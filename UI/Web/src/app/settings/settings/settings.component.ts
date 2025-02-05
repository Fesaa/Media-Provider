import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../../_services/nav.service";
import {ServerSettingsComponent} from "./_components/server-settings/server-settings.component";
import {PagesSettingsComponent} from "./_components/pages-settings/pages-settings.component";
import {dropAnimation} from "../../_animations/drop-animation";
import {ActivatedRoute, Router} from "@angular/router";
import {hasPermission, Perm, User} from "../../_models/user";
import {AccountService} from "../../_services/account.service";
import {UserSettingsComponent} from "./_components/user-settings/user-settings.component";
import {PreferenceSettingsComponent} from "./_components/preference-settings/preference-settings.component";
import {Button} from "primeng/button";
import {NotificationsComponent} from "./_components/notifications/notifications.component";
import {NotificationService} from "../../_services/notification.service";
import {EventType, SignalRService} from "../../_services/signal-r.service";

export enum SettingsID {

  Server = "server",
  Preferences = "preferences",
  Notifications = "notifications",
  Pages = "pages",
  User = "user"

}

@Component({
  selector: 'app-settings',
  imports: [
    ServerSettingsComponent,
    PagesSettingsComponent,
    UserSettingsComponent,
    PreferenceSettingsComponent,
    Button,
    NotificationsComponent
  ],
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.css',
  animations: [dropAnimation]
})
export class SettingsComponent implements OnInit {
  showMobileConfig = false;

  user: User | null = null;
  selected: SettingsID = SettingsID.Server;
  settings: { id: SettingsID, title: string, icon: string, perm: Perm, badge?: number }[] = [
    {
      id: SettingsID.Server,
      title: 'Server',
      icon: 'pi pi-server',
      perm: Perm.WriteConfig
    },
    {
      id: SettingsID.Preferences,
      title: "Preferences",
      icon: 'pi pi-ethereum',
      perm: Perm.WriteConfig,
    },
    {
      id: SettingsID.Notifications,
      title: "Notifications",
      icon: "pi pi-inbox",
      perm: Perm.All,
    },
    {
      id: SettingsID.Pages,
      title: 'Pages',
      icon: 'pi pi-thumbtack',
      perm: Perm.All,
    },
    {
      id: SettingsID.User,
      title: 'Users',
      icon: 'pi pi-users',
      perm: Perm.WriteUser,
    },
  ]
  protected readonly SettingsID = SettingsID;

  constructor(private navService: NavService,
              private cdRef: ChangeDetectorRef,
              private activatedRoute: ActivatedRoute,
              private router: Router,
              private accountService: AccountService,
              private notificationService: NotificationService,
              private signalR: SignalRService,
  ) {
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.user = user;
      } else {
        this.router.navigateByUrl('/login');
        return;
      }

      if (!this.canSee(this.selected)) {
        this.setSettings(this.settings.find(s => this.canSee(s.id))!.id)
      }
    })

    this.activatedRoute.fragment.subscribe(fragment => {
      if (fragment) {
        if (Object.values(SettingsID).find(id => id === fragment)) {
          this.selected = fragment as SettingsID;
        }
      }
    })

    this.notificationService.amount().subscribe(amount => {
      this.settings = this.settings.map(s => {
        if (s.id !== SettingsID.Notifications) {
          return s;
        }
        s.badge = amount
        return s;
      })
    })

    this.signalR.events$.subscribe(event => {
      if (event.type === EventType.Notification) {
        this.settings = this.settings.map(s => {
          if (s.id !== SettingsID.Notifications) {
            return s;
          }
          s.badge = (s.badge || 0) + 1;
          return s;
        })
      }

      if (event.type === EventType.NotificationRead) {
        this.settings = this.settings.map(s => {
          if (s.id !== SettingsID.Notifications) {
            return s;
          }
          s.badge = (s.badge || 1) - 1;
          return s;
        })
      }
    })
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
  }

  toggleMobile() {
    this.showMobileConfig = !this.showMobileConfig;
    this.cdRef.markForCheck();
  }

  setSettings(id: SettingsID) {
    this.selected = id;
    this.router.navigate([], {fragment: id});
    this.cdRef.markForCheck();
  }

  canSee(id: SettingsID): boolean {
    if (!this.user) {
      return false;
    }

    const setting = this.settings.find(setting => setting.id === id);
    if (!setting) {
      return false;
    }

    return hasPermission(this.user, setting.perm);
  }

  protected readonly String = String;
}
