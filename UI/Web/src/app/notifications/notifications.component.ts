import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {TableModule} from "primeng/table";
import {NotificationService} from "../_services/notification.service";
import {GroupWeight, Notification, NotificationGroup} from "../_models/notifications";
import {Tag} from "primeng/tag";
import {Button} from "primeng/button";
import {Tooltip} from "primeng/tooltip";
import {ToastService} from "../_services/toast.service";
import {Dialog} from "primeng/dialog";
import {Card} from "primeng/card";
import {DialogService} from "../_services/dialog.service";
import { SortedList } from '../shared/data-structures/sorted-list';
import {Select} from "primeng/select";
import {FormsModule} from "@angular/forms";
import {NavService} from "../_services/nav.service";
import {Checkbox} from "primeng/checkbox";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-notifications',
  imports: [
    TableModule,
    Tag,
    Button,
    Tooltip,
    Dialog,
    Card,
    Select,
    FormsModule,
    Checkbox,
    TranslocoDirective,
  ],
  templateUrl: './notifications.component.html',
  styleUrl: './notifications.component.css'
})
export class NotificationsComponent implements OnInit {

  notifications: SortedList<Notification> = new SortedList<Notification>(
    (n1: Notification, n2: Notification) => {
      const d1 = new Date(n1.CreatedAt)
      const d2 = new Date(n2.CreatedAt);

      if (d1.getDay() !== d2.getDay()) {
        return d1.getDay() - d2.getDay();
      }

      if (n1.group === n2.group) {
        return new Date(n1.CreatedAt).getTime() - new Date(n2.CreatedAt).getTime()
      }

      return GroupWeight(n2.group) - GroupWeight(n1.group);
    }
  );
  infoVisibility: {[key: number]: boolean} = {};
  selectedNotifications: number[] = [];

  timeAgoOptions = [{
    label: 'Last 24 hours',
    value: 1
  }, {
    label: "Last 7 days",
    value: 7
  }, {
    label: "Last 30 days",
    value: 30
  }, {
    label: "All",
    value: -1
  }]
  timeAgo: number = 7;

  constructor(
    private notificationService: NotificationService,
    private toastService: ToastService,
    private dialogService: DialogService,
    private navService: NavService,
  ) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
    this.refresh()
  }

  refresh() {
    let date: Date | undefined = undefined;
    if (this.timeAgo !== -1) {
      date = new Date();
      date.setDate(date.getDate() - this.timeAgo);
    }
    this.notificationService.all(date).subscribe((notifications) => {
      this.notifications.set(notifications)
    })
  }

  show(id: number) {
    this.infoVisibility = {} // close others
    this.infoVisibility[id] = true;
  }

  groupSeverity(group: NotificationGroup) {
    switch (group) {
      case NotificationGroup.Content:
      case NotificationGroup.General:
        return "info"
      case NotificationGroup.Error:
        return "danger"
      case NotificationGroup.Security:
        return "warn"
    }
  }

  markRead(notification: Notification) {
    this.notificationService.markAsRead(notification.ID).subscribe({
      next: () => {
        this.notifications.removeFunc((n: Notification) => n.ID == notification.ID);
        notification.read = true;
        this.notifications.add(notification);
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  markUnRead(notification: Notification) {
    this.notificationService.markAsUnread(notification.ID).subscribe({
      next: () => {
        this.notifications.removeFunc((n: Notification) => n.ID == notification.ID);
        notification.read = false;
        this.notifications.add(notification);
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async readSelected() {
    // Filter out read notifications
    this.selectedNotifications = this.selectedNotifications.filter(n => {
      const not = this.notifications.getFunc((n2: Notification) => n2.ID === n)
      return not && !not.read
    })

    if (this.selectedNotifications.length === 0) {
      this.toastService.warningLoco("notifications.toasts.no-selected");
      return;
    }

    if (!await this.dialogService.openDialog("notifications.confirm-read-many",
      {amount: this.selectedNotifications.length})) {
      return;
    }

    this.notificationService.readMany(this.selectedNotifications).subscribe({
      next: () => {
        this.toastService.successLoco("notifications.toasts.read-many-success", {amount: this.selectedNotifications.length})
        this.notifications.set(this.notifications.items().map(n => {
          if (this.selectedNotifications.includes(n.ID)) {
            n.read = true;
          }
          return n;
        }))
        this.selectedNotifications = [];
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async deleteSelected() {
    if (this.selectedNotifications.length === 0) {
      this.toastService.warningLoco("notifications.toasts.no-selected");
      return;
    }

    if (!await this.dialogService.openDialog("notifications.confirm-delete-many",
      {amount: this.selectedNotifications.length})) {
      return;
    }

    this.notificationService.deleteMany(this.selectedNotifications).subscribe({
      next: () => {
        this.toastService.successLoco("notifications.toasts.delete-success", {amount: this.selectedNotifications.length})
        this.notifications.set(this.notifications.items().
        filter(n => !this.selectedNotifications.includes(n.ID)))
        this.selectedNotifications = [];
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async delete(notification: Notification) {
    if (!await this.dialogService.openDialog("notifications.confirm-delete", {title: notification.title})) {
      return;
    }

    this.notificationService.deleteNotification(notification.ID).subscribe({
      next: () => {
        this.notifications.removeFunc((n: Notification) => n.ID == notification.ID);
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

}
