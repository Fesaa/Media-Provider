import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {PageService} from "../_services/page.service";
import {Page} from "../_models/page";
import {ActivatedRoute} from "@angular/router";
import {AsyncPipe} from "@angular/common";
import {AccountService} from "../_services/account.service";
import {NavService} from "../_services/nav.service";
import {dropAnimation} from "../_animations/drop-animation";
import {MenuItem} from "primeng/api";
import {Menubar} from "primeng/menubar";
import {NotificationService} from "../_services/notification.service";
import {EventType, SignalRService} from "../_services/signal-r.service";
import {BadgeDirective} from "primeng/badge";
import {TranslocoDirective, TranslocoService} from "@jsverse/transloco";
import {User} from "../_models/user";
import {OidcService} from "../_services/oidc.service";

@Component({
  selector: 'app-nav-header',
  imports: [
    AsyncPipe,
    Menubar,
    BadgeDirective,
    TranslocoDirective
  ],
  templateUrl: './nav-header.component.html',
  styleUrl: './nav-header.component.scss',
  animations: [dropAnimation]
})
export class NavHeaderComponent implements OnInit {

  isMenuOpen = false;
  index: number | undefined;
  path: string | undefined;

  pages: Page[] = [];
  accountItems: MenuItem[] | undefined;
  pageItems: MenuItem[] | undefined;

  notifications: number = 0;

  constructor(private pageService: PageService,
              private route: ActivatedRoute,
              private cdRef: ChangeDetectorRef,
              protected accountService: AccountService,
              protected navService: NavService,
              private notificationService: NotificationService,
              private signalR: SignalRService,
              private transLoco: TranslocoService,
              private oidcService: OidcService,
  ) {
  }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      const index = params['index'];
      if (index) {
        this.index = parseInt(index);
      } else {
        this.index = undefined;
      }
    })

    this.accountService.currentUser$.subscribe(user => {
      if (!user) {
        return;
      }

      this.transLoco.events$.subscribe(event => {
        if (event.type !== "translationLoadSuccess") {
          return;
        }

        this.setPageItems()
        this.setAccountItems(user)
      });

      this.notificationService.amount().subscribe(amount => {
        this.notifications = amount;
      });
    });

    this.signalR.events$.subscribe(event => {
      if (event.type == EventType.NotificationAdd) {
        this.notifications++;
      }
      if (event.type === EventType.NotificationRead) {
        const amount: number = event.data.amount;
        this.notifications -= amount;
      }
    })

  }

  setPageItems() {
    this.pageService.pages$.subscribe(pages => {
      this.pages = pages;

      this.pageItems = this.pages.map(page => {
        return {
          label: page.title,
          routerLink: 'page',
          queryParams: {index: page.ID},
          icon: page.icon === '' ? undefined : 'pi ' + page.icon,
        }
      })
      this.pageItems = [{
        label: this.transLoco.translate("nav-bar.home"),
        routerLink: 'home',
        icon: 'pi pi-home',
      }, ...this.pageItems]


      this.cdRef.detectChanges();
    });
  }

  setAccountItems(user: User) {
    this.accountItems = [
      {
        label: user.name,
        icon: "pi pi-user",
        items: [
          {
            label: this.transLoco.translate("nav-bar.subscriptions"),
            routerLink: "subscriptions",
            icon: "pi pi-wave-pulse"
          },
          {
            label: this.transLoco.translate("nav-bar.notifications"),
            routerLink: "notifications",
            icon: "pi pi-inbox"
          },
          {
            label: this.transLoco.translate("nav-bar.settings"),
            routerLink: "settings",
            icon: "pi pi-cog",
          },
          {
            label: this.transLoco.translate("nav-bar.sign-out"),
            command: () => {
              this.oidcService.logout();
              this.accountService.logout()
            },
            icon: "pi pi-sign-out"
          }
        ]
      }
    ];

    if (window.innerWidth <= 768) {
      this.accountItems = this.accountItems[0].items;
    }
  }

  severity(): "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | undefined {
    if (this.notifications < 4) {
      return "info"
    }

    if (this.notifications < 10) {
      return "warn"
    }

    return "danger"
  }

}
