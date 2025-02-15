import {Component, OnInit, ViewContainerRef} from '@angular/core';
import {RouterOutlet} from '@angular/router';
import {AccountService} from "./_services/account.service";
import {NavHeaderComponent} from "./nav-header/nav-header.component";
import {Title} from "@angular/platform-browser";
import {DialogService} from "./_services/dialog.service";
import {EventType, SignalRService} from "./_services/signal-r.service";
import {Toast} from "primeng/toast";
import {MessageService} from "primeng/api";
import {Notification} from "./_models/notifications";

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, NavHeaderComponent, Toast],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent implements OnInit {
  title = 'Media Provider';

  constructor(
    protected accountService: AccountService,
    private titleService: Title,
    private vcr: ViewContainerRef,
    private ds: DialogService,
    private signalR: SignalRService,
    private messageService: MessageService,
  ) {
    this.titleService.setTitle(this.title);
    this.ds.viewContainerRef = this.vcr;
  }

  ngOnInit(): void {
    this.accountService.currentUser$.subscribe(user => {
      if (!user) {
        return;
      }

      this.signalR.startConnection(user);
      this.signalR.events$.subscribe(event => {
        switch (event.type) {
          case EventType.Notification:
            const notification = event.data as Notification;
            this.messageService.add({
              severity: notification.colour,
              summary: notification.title,
              detail: notification.summary, // I know they're switched here
            });
        }
      })
    })
  }
}
