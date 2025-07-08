import {Component, effect, inject, OnInit, ViewContainerRef} from '@angular/core';
import {Router, RouterOutlet} from '@angular/router';
import {AccountService} from "./_services/account.service";
import {NavHeaderComponent} from "./nav-header/nav-header.component";
import {Title} from "@angular/platform-browser";
import {DialogService} from "./_services/dialog.service";
import {EventType, SignalRService} from "./_services/signal-r.service";
import {Toast} from "primeng/toast";
import {MessageService} from "primeng/api";
import {Notification} from "./_models/notifications";
import {OidcService} from "./_services/oidc.service";
import {NavService} from "./_services/nav.service";

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, NavHeaderComponent, Toast],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent implements OnInit {
  title = 'Media Provider';

  private readonly oidcService = inject(OidcService);

  constructor(
    protected accountService: AccountService,
    private titleService: Title,
    private vcr: ViewContainerRef,
    private ds: DialogService,
    private signalR: SignalRService,
    private messageService: MessageService,
    private navService: NavService,
    private router: Router,
  ) {
    this.titleService.setTitle(this.title);
    this.ds.viewContainerRef = this.vcr;

    // Login automatically when a token is available
    effect(() => {
      const inUse = this.oidcService.inUse();
      const user = this.accountService.currentUserSignal();
      if (!inUse || !this.oidcService.token || user) return;

      this.accountService.loginByToken(this.oidcService.token).subscribe({
        next: () => {
          this.navService.handleLogin();
        },
        error: err => {
          console.error(err);
          this.accountService.logout();
          this.oidcService.logout();
          this.router.navigate(['login']);
        }
      });
    });

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
