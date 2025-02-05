import {Component, OnInit} from '@angular/core';
import {TableModule} from "primeng/table";
import {NotificationService} from "../../../../_services/notification.service";
import {EventType, SignalRService} from "../../../../_services/signal-r.service";
import {Notification, NotificationGroup} from "../../../../_models/notifications";
import {Tag} from "primeng/tag";
import {Button} from "primeng/button";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from "../../../../_services/message.service";
import {Dialog} from "primeng/dialog";
import {Card} from "primeng/card";
import {DialogService} from "../../../../_services/dialog.service";

@Component({
  selector: 'app-notifications',
  imports: [
    TableModule,
    Tag,
    Button,
    Tooltip,
    Dialog,
    Card,
  ],
  templateUrl: './notifications.component.html',
  styleUrl: './notifications.component.css'
})
export class NotificationsComponent implements OnInit {

  // Make into sorted list
  notifications: {[key: string]: Notification[]} = {};
  infoVisibility: {[key: number]: boolean} = {};

  constructor(
    private notificationService: NotificationService,
    private messageService: MessageService,
    private dialogService: DialogService,
  ) {
    this.notifications[NotificationGroup.Content] = [];
    this.notifications[NotificationGroup.General] = [];
    this.notifications[NotificationGroup.Security] = [];
    this.notifications[NotificationGroup.Error] = [];
  }

  ngOnInit(): void {
    this.notificationService.all().subscribe((notifications) => {
      for (const notification of notifications) {
        this.notifications[notification.group].push(notification)
      }
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

  groupedNotifications(): Notification[] {
    const notifications: Notification[] = [];
    for (const group of [NotificationGroup.Error, NotificationGroup.Content, NotificationGroup.Security, NotificationGroup.General]) {
      const copy = this.notifications[group]
      copy.sort((a, b) => new Date(a.CreatedAt).getTime() - new Date(b.CreatedAt).getTime())
      for (const notification of copy) {
        notifications.push(notification)
      }
    }
    return notifications;
  }

  markRead(notification: Notification) {
    this.notificationService.markAsRead(notification.ID).subscribe({
      next: () => {
        this.notifications[notification.group] = this.notifications[notification.group].map(n => {
          if (n.ID !== notification.ID) {
            return n
          }
          n.read = true
          return n;
        })
      },
      error: err => {
        this.messageService.error("Failed to read notifications", err.error.message);
      }
    })
  }

  markUnRead(notification: Notification) {
    this.notificationService.markAsUnread(notification.ID).subscribe({
      next: () => {
        this.notifications[notification.group] = this.notifications[notification.group].map(n => {
          if (n.ID !== notification.ID) {
            return n
          }
          n.read = false
          return n;
        })
      },
      error: err => {
        this.messageService.error("Failed to unread notifications", err.error.message);
      }
    })
  }

  async delete(notification: Notification) {
    if (!await this.dialogService.openDialog(`Are you sure you want to delete ${notification.title} (${notification.ID})`)) {
      return;
    }

    this.notificationService.deleteNotification(notification.ID).subscribe({
      next: () => {
        this.notifications[notification.group] = this.notifications[notification.group].filter(n => n.ID !== notification.ID)
      },
      error: err => {
        this.messageService.error("Failed to delete notifications", err.error.message);
      }
    })
  }

}
