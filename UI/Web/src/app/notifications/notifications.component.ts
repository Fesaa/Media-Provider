import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {TableModule} from "primeng/table";
import {NotificationService} from "../_services/notification.service";
import {GroupWeight, Notification, NotificationGroup} from "../_models/notifications";
import {Tag} from "primeng/tag";
import {Button} from "primeng/button";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from "../_services/message.service";
import {Dialog} from "primeng/dialog";
import {Card} from "primeng/card";
import {DialogService} from "../_services/dialog.service";
import { SortedList } from '../shared/data-structures/sorted-list';
import {Select} from "primeng/select";
import {FormsModule} from "@angular/forms";
import {NavService} from "../_services/nav.service";
import {Checkbox} from "primeng/checkbox";

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
    private messageService: MessageService,
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
        this.messageService.error("Failed to read notifications", err.error.message);
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
        this.messageService.error("Failed to unread notifications", err.error.message);
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
      this.messageService.warning("No notifications selected");
      return;
    }

    if (!await this.dialogService.openDialog(
      `Are you sure you want to mark ${this.selectedNotifications.length} notification(s) as read?`)) {
      return;
    }

    this.notificationService.readMany(this.selectedNotifications).subscribe({
      next: () => {
        this.messageService.success(`Marked ${this.selectedNotifications.length} notification(s) as read`);
        this.notifications.set(this.notifications.items().map(n => {
          if (this.selectedNotifications.includes(n.ID)) {
            n.read = true;
          }
          return n;
        }))
        this.selectedNotifications = [];
      },
      error: err => {
        this.messageService.error("Failed to read notifications", err.error.message);
      }
    })
  }

  async deleteSelected() {
    if (this.selectedNotifications.length === 0) {
      this.messageService.warning("No notifications selected");
      return;
    }

    if (!await this.dialogService.openDialog(
      `Are you sure you want to delete ${this.selectedNotifications.length} notification(s)?`)) {
      return;
    }

    this.notificationService.deleteMany(this.selectedNotifications).subscribe({
      next: () => {
        this.messageService.success(`Deleted ${this.selectedNotifications.length} notification(s)`);
        this.notifications.set(this.notifications.items().
        filter(n => !this.selectedNotifications.includes(n.ID)))
        this.selectedNotifications = [];
      },
      error: err => {
        this.messageService.error("Failed to delete notifications", err.error.message);
      }
    })
  }

  async delete(notification: Notification) {
    if (!await this.dialogService.openDialog(`Are you sure you want to delete ${notification.title} (${notification.ID})`)) {
      return;
    }

    this.notificationService.deleteNotification(notification.ID).subscribe({
      next: () => {
        this.notifications.removeFunc((n: Notification) => n.ID == notification.ID);
      },
      error: err => {
        this.messageService.error("Failed to delete notifications", err.error.message);
      }
    })
  }

}
