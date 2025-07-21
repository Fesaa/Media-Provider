import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  OnInit,
  signal,
  computed, HostListener
} from '@angular/core';
import {PageService} from "../_services/page.service";
import {ActivatedRoute, RouterLink} from "@angular/router";
import {AccountService} from "../_services/account.service";
import {NavService} from "../_services/nav.service";
import {NotificationService} from "../_services/notification.service";
import {SignalRService, EventType} from "../_services/signal-r.service";
import {TranslocoPipe, TranslocoService} from "@jsverse/transloco";
import {OidcService} from "../_services/oidc.service";
import {User} from "../_models/user";
import {Page} from "../_models/page";
import {AsyncPipe} from "@angular/common";
import {animate, style, transition, trigger} from "@angular/animations";

interface NavItem {
  label: string;
  icon?: string;
  routerLink?: string;
  queryParams?: Record<string, any>;
  command?: () => void;
}

const drawerAnimation = trigger('drawerAnimation', [
  transition(':enter', [
    style({ transform: 'translateX(-100%)', opacity: 0 }),
    animate('250ms ease-out', style({ transform: 'translateX(0)', opacity: 1 })),
  ]),
  transition(':leave', [
    animate('200ms ease-in', style({ transform: 'translateX(-100%)', opacity: 0 })),
  ]),
]);

const dropdownAnimation = trigger('dropdownAnimation', [
  transition(':enter', [
    style({ opacity: 0, transform: 'translateY(-8px)' }),
    animate('150ms ease-out', style({ opacity: 1, transform: 'translateY(0)' })),
  ]),
  transition(':leave', [
    animate('100ms ease-in', style({ opacity: 0, transform: 'translateY(-8px)' })),
  ]),
]);



@Component({
  selector: 'app-nav-header',
  templateUrl: './nav-header.component.html',
  styleUrls: ['./nav-header.component.scss'],
  imports: [
    RouterLink,
    AsyncPipe
  ],
  animations: [drawerAnimation, dropdownAnimation],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class NavHeaderComponent implements OnInit {

  notifications = signal(0);
  currentUser = signal<User | null>(null);
  pageItems = signal<Page[]>([]);
  accountItems = signal<NavItem[]>([]);

  isMobileMenuOpen = signal(false);
  isAccountDropdownOpen = signal(false);

  severity = computed((): 'info' | 'warn' | 'danger' => {
    const count = this.notifications();
    if (count < 4) return 'info';
    if (count < 10) return 'warn';
    return 'danger';
  });

  constructor(
    private pageService: PageService,
    private route: ActivatedRoute,
    private cdRef: ChangeDetectorRef,
    private accountService: AccountService,
    protected navService: NavService,
    private notificationService: NotificationService,
    private signalR: SignalRService,
    private transLoco: TranslocoService,
    private oidcService: OidcService
  ) {}

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      const index = params['index'];
      if (index) {
        // Not used here, preserved for logic continuity
      }
    });

    this.accountService.currentUser$.subscribe(user => {
      if (!user) return;
      this.currentUser.set(user);

      this.transLoco.events$.subscribe(event => {
        if (event.type === "translationLoadSuccess") {
          this.loadPages();
          this.setAccountItems(user);
        }
      });

      this.notificationService.amount().subscribe(amount => {
        this.notifications.set(amount);
      });
    });

    this.signalR.events$.subscribe(event => {
      if (event.type === EventType.NotificationAdd) {
        this.notifications.update(n => n + 1);
      }
      if (event.type === EventType.NotificationRead) {
        const amount: number = event.data.amount;
        this.notifications.update(n => Math.max(0, n - amount));
      }
    });
  }

  loadPages() {
    this.pageService.pages$.subscribe(pages => {
      this.pageItems.set([
        {
          title: this.transLoco.translate("nav-bar.home"),
          ID: -1,
          icon: "pi-home",
          dirs: [],
          custom_root_dir: '',
          modifiers: [],
          providers: [],
          sortValue: -100,
        },
        ...pages
      ]);
      this.cdRef.markForCheck();
    });
  }

  setAccountItems(user: User) {
    const items: NavItem[] = [
      {
        label: this.transLoco.translate("nav-bar.subscriptions"),
        icon: "pi-wave-pulse",
        routerLink: "/subscriptions"
      },
      {
        label: this.transLoco.translate("nav-bar.notifications"),
        icon: "pi-inbox",
        routerLink: "/notifications"
      },
      {
        label: this.transLoco.translate("nav-bar.settings"),
        icon: "pi-cog",
        routerLink: "/settings"
      },
      {
        label: this.transLoco.translate("nav-bar.sign-out"),
        icon: "pi-sign-out",
        command: () => this.logout()
      }
    ];

    this.accountItems.set(items);
  }

  logout() {
    this.oidcService.logout();
    this.accountService.logout();
  }

  toggleMobileMenu() {
    this.isMobileMenuOpen.update(v => !v);
  }

  toggleAccountDropdown() {
    this.isAccountDropdownOpen.update(v => !v);
  }

  @HostListener('document:click', ['$event'])
  onClickOutside(event: MouseEvent): void {
    const target = event.target as HTMLElement;
    if (
      this.isAccountDropdownOpen() &&
      !target.closest('.account-dropdown') &&
      !target.closest('.account-toggle')
    ) {
      this.isAccountDropdownOpen.set(false);
    }
  }

}
