import {Component, inject, OnInit} from '@angular/core';
import {NotificationService} from '../../../_services/notification.service';
import {Notification} from "../../../_models/notifications";
import {TranslocoDirective} from "@jsverse/transloco";
import {UtcToLocalTimePipe} from "../../../_pipes/utc-to-local.pipe";
import {ToastService} from "../../../_services/toast.service";

@Component({
  selector: 'app-recently-downloaded',
  imports: [
    TranslocoDirective,
    UtcToLocalTimePipe,
  ],
  templateUrl: './recently-downloaded.component.html',
  styleUrl: './recently-downloaded.component.scss'
})
export class RecentlyDownloadedComponent implements OnInit{

  private readonly notificationService = inject(NotificationService);
  private readonly toastService = inject(ToastService);

  downloads: Notification[] = [];
  infoVisibility: {[key: number]: boolean} = {};

  ngOnInit(): void {
    this.notificationService.recent().subscribe((recent) => {
      this.downloads = recent;
    });
  }

  markRead(download: Notification) {
    this.notificationService.markAsRead(download.ID).subscribe({
      next: () => {
        this.downloads = this.downloads.filter(d => d.ID !== download.ID);
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  show(id: number) {
    this.infoVisibility = {} // close others
    this.infoVisibility[id] = true;
  }

  formattedBody(notification: Notification) {
    let body = notification.body;
    body = body ? body.replace(/\n/g, '<br>') : '';
    return body;
  }

}
