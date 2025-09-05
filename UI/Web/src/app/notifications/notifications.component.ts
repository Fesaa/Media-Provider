import {Component, computed, inject, OnInit, signal} from '@angular/core';
import {NotificationService} from "../_services/notification.service";
import {GroupWeight, Notification} from "../_models/notifications";
import {ToastService} from "../_services/toast.service";
import {FormControl, FormGroup, FormsModule, NonNullableFormBuilder, ReactiveFormsModule} from "@angular/forms";
import {NavService} from "../_services/nav.service";
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {UtcToLocalTimePipe} from "../_pipes/utc-to-local.pipe";
import {TableComponent} from "../shared/_component/table/table.component";
import {BadgeComponent} from "../shared/_component/badge/badge.component";
import {ModalService} from "../_services/modal.service";
import {NgbTooltip} from "@ng-bootstrap/ng-bootstrap";
import {NotificationInfoModalComponent} from "./_components/notification-info-modal/notification-info-modal.component";
import {DefaultModalOptions} from "../_models/default-modal-options";
import {debounce, debounceTime, distinctUntilChanged, map, switchMap, tap} from "rxjs";

@Component({
  selector: 'app-notifications',
  imports: [
    FormsModule,
    TranslocoDirective,
    UtcToLocalTimePipe,
    TableComponent,
    BadgeComponent,
    NgbTooltip,
    ReactiveFormsModule,
  ],
  templateUrl: './notifications.component.html',
  styleUrl: './notifications.component.scss'
})
export class NotificationsComponent implements OnInit {

  private readonly modalService = inject(ModalService);
  private readonly fb = inject(NonNullableFormBuilder);

  notifications = signal<Notification[]>([]);

  sortedNotifications = computed(() => {
    const notifications = this.notifications();
    return notifications.sort((n1: Notification, n2: Notification) => {
      const d1 = new Date(n1.CreatedAt)
      const d2 = new Date(n2.CreatedAt);

      if (n1.group === n2.group) {
        return d2.getTime() - d1.getTime();
      }

      return GroupWeight(n2.group) - GroupWeight(n1.group);
    });
  });

  timeAgoForm = this.fb.group({
    timeAgo: this.fb.control<number>(999_999_999),
  });

  selectedNotifications: number[] = [];
  allCheck: boolean = false;

  timeAgoOptions = [
    {
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

  constructor(
    private notificationService: NotificationService,
    private toastService: ToastService,
    private navService: NavService,
  ) {

    this.timeAgoForm.valueChanges.pipe(
      distinctUntilChanged(),
      debounceTime(100),
      map(() => this.timeAgoForm.get('timeAgo')!.value),
      map(timeAgo => {
        if (timeAgo === -1) return undefined;

        const date = new Date();
        date.setDate(date.getDate() - timeAgo);
        return date;
      }),
      switchMap(date => this.notificationService.all(date)),
      tap(notifications => this.notifications.set(notifications))
    ).subscribe();

  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true);
    this.timeAgoForm.get('timeAgo')!.setValue(30);
  }

  toggleAll() {
    if (this.allCheck) {
      this.selectedNotifications = this.notifications().map(n => n.ID)
    } else {
      this.selectedNotifications = []
    }
  }

  toggleSelect(id: number) {
    if (this.selectedNotifications.indexOf(id) === -1) {
      this.selectedNotifications.push(id);
    } else {
      this.selectedNotifications = this.selectedNotifications.filter(i => i !== id);
    }
  }

  show(n: Notification) {
    const [_, component] = this.modalService.open(NotificationInfoModalComponent, DefaultModalOptions);
    component.notification.set(n);
  }

  markRead(notification: Notification) {
    this.notificationService.markAsRead(notification.ID).subscribe({
      next: () => {
        this.notifications.update(notifications => notifications.map(n => {
          if (n.ID !== notification.ID) return n;

          n.read = true;
          return n;
        }));
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  markUnRead(notification: Notification) {
    this.notificationService.markAsUnread(notification.ID).subscribe({
      next: () => {
        this.notifications.update(notifications => notifications.map(n => {
          if (n.ID !== notification.ID) return n;

          n.read = false;
          return n;
        }));
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async readSelected() {
    // Filter out read notifications
    this.selectedNotifications = this.selectedNotifications.filter(n => {
      const not = this.notifications().find(n => n.ID === n.ID);
      return not && !not.read
    })

    if (this.selectedNotifications.length === 0) {
      this.toastService.warningLoco("notifications.toasts.no-selected");
      return;
    }

    if (!await this.modalService.confirm({
      question: translate('notifications.confirm-read-many', {amount: this.selectedNotifications.length})
    })) {
      return;
    }

    this.notificationService.readMany(this.selectedNotifications).subscribe({
      next: () => {
        this.toastService.successLoco("notifications.toasts.read-many-success", {amount: this.selectedNotifications.length})
        this.notifications.set(this.notifications().map(n => {
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

    if (!await this.modalService.confirm({
      question: translate('notifications.confirm-delete-many', {amount: this.selectedNotifications.length})
    })) {
      return;
    }

    this.notificationService.deleteMany(this.selectedNotifications).subscribe({
      next: () => {
        this.toastService.successLoco("notifications.toasts.delete-success", {amount: this.selectedNotifications.length})
        this.notifications.set(this.notifications().
        filter(n => !this.selectedNotifications.includes(n.ID)))
        this.selectedNotifications = [];
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async delete(notification: Notification) {
    if (!await this.modalService.confirm({
      question: translate('notifications.confirm-delete', {title: notification.title})
    })) {
      return;
    }

    this.notificationService.deleteNotification(notification.ID).subscribe({
      next: () => {
        this.notifications.update(notifications => notifications.filter(n => n.ID !== notification.ID))
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  trackBy(idx: number, notification: Notification): string {
    return `${notification.ID}`
  }

}
