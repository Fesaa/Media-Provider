import {Component, HostListener, inject, OnInit, ViewContainerRef} from '@angular/core';
import {Router, RouterOutlet} from '@angular/router';
import {AccountService} from "./_services/account.service";
import {NavHeaderComponent} from "./nav-header/nav-header.component";
import {Title} from "@angular/platform-browser";
import {EventType, SignalRService} from "./_services/signal-r.service";
import {Notification} from "./_models/notifications";
import {OidcEvents, OidcService} from "./_services/oidc.service";
import {NavService} from "./_services/nav.service";
import {ToastrService} from "ngx-toastr";

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, NavHeaderComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit {
  title = 'Media Provider';

  private readonly oidcService = inject(OidcService);
  private readonly toastr = inject(ToastrService);

  constructor(
    protected accountService: AccountService,
    private titleService: Title,
    private signalR: SignalRService,
    private navService: NavService,
    private router: Router,
  ) {
    this.titleService.setTitle(this.title);

    this.oidcService.events$.subscribe(event => {
      if (event.type !== OidcEvents.TokenRefreshed) return;

      const user = this.accountService.currentUserSignal();
      if (user) {
        user.oidcToken = this.oidcService.token;
        return;
      }

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
    this.updateVh();

    this.signalR.events$.subscribe(event => {
      switch (event.type) {
        case EventType.Notification:
          const notification = event.data as Notification;
          // TODO: Correct level
          this.toastr.info(notification.title, notification.summary);
      }
    });

    this.accountService.currentUser$.subscribe(user => {
      if (!user) {
        return;
      }

      if (user.oidcToken && !this.oidcService.hasValidAccessToken()) return;

      this.signalR.startConnection(user);
    })
  }

  @HostListener('window:resize')
  @HostListener('window:orientationchange')
  setDocHeight() {
    this.updateVh();
  }

  private updateVh(): void {
    // Sets a CSS variable for the actual device viewport height. Needed for mobile dev.
    const vh = window.innerHeight * 0.01;
    document.documentElement.style.setProperty('--vh', `${vh}px`);
  }
}
