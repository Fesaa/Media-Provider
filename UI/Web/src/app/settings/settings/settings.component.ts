import {Component, computed, effect, inject, signal} from '@angular/core';
import {ActivatedRoute, Router} from '@angular/router';
import {NavService} from '../../_services/nav.service';
import {AccountService} from '../../_services/account.service';
import {hasPermission, Perm, User} from '../../_models/user';
import {PreferenceSettingsComponent} from "./_components/preference-settings/preference-settings.component";
import {PagesSettingsComponent} from "./_components/pages-settings/pages-settings.component";
import {ServerSettingsComponent} from "./_components/server-settings/server-settings.component";
import {UserSettingsComponent} from "./_components/user-settings/user-settings.component";
import {TranslocoDirective} from "@jsverse/transloco";
import {AccountSettingsComponent} from "./_components/account-settings/account-settings.component";

export enum SettingsID {
  Account = "account",
  Server = "server",
  Preferences = "preferences",
  Pages = "pages",
  User = "user"
}

interface SettingsTab {
  id: SettingsID,
  title: string,
  icon: string,
  perm: Perm,
}

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [
    PreferenceSettingsComponent,
    PagesSettingsComponent,
    ServerSettingsComponent,
    UserSettingsComponent,
    TranslocoDirective,
    AccountSettingsComponent

  ],
  templateUrl: './settings.component.html',
  styleUrls: ['./settings.component.scss']
})
export class SettingsComponent {
  private navService = inject(NavService);
  private accountService = inject(AccountService);
  private router = inject(Router);
  private route = inject(ActivatedRoute);

  readonly SettingsID = SettingsID;

  user = signal<User | null>(null);
  selected = signal<SettingsID>(SettingsID.Account);
  showMobileConfig = signal(false);

  readonly settings: SettingsTab[] = [
    { id: SettingsID.Account, title: "Account", icon: 'fa fa-user', perm: Perm.All },
    { id: SettingsID.Preferences, title: "Preferences", icon: 'fa fa-heart', perm: Perm.WriteConfig },
    { id: SettingsID.Pages, title: 'Pages', icon: 'fa fa-thumbtack', perm: Perm.All },
    { id: SettingsID.Server, title: 'Server', icon: 'fa fa-server', perm: Perm.WriteConfig },
    { id: SettingsID.User, title: 'Users', icon: 'fa fa-users', perm: Perm.WriteUser },
  ];

  readonly visibleSettings = computed(() => {
    this.user(); // Compute when user changes

    return this.settings.filter(setting => this.canSee(setting.id));
  });

  constructor() {
    this.navService.setNavVisibility(true);

    this.accountService.currentUser$.subscribe(user => {
      if (!user) {
        this.router.navigateByUrl('/login');
        return;
      }
      this.user.set(user);

      if (!this.canSee(this.selected())) {
        this.selected.set(this.visibleSettings()[0].id);
      }
    });

    this.route.fragment.subscribe(fragment => {
      if (fragment && Object.values(SettingsID).includes(fragment as SettingsID)) {
        this.selected.set(fragment as SettingsID);
      }
    });

    effect(() => {
      this.router.navigate([], { fragment: this.selected() });
    });
  }

  toggleMobile() {
    this.showMobileConfig.update(v => !v);
  }

  setSettings(id: SettingsID) {
    this.selected.set(id);
    this.showMobileConfig.set(false);
  }

  canSee(id: SettingsID): boolean {
    const user = this.user();
    if (!user) return false;

    const setting = this.settings.find(s => s.id === id);
    if (!setting) return false;

    return hasPermission(user, setting.perm);
  }

  isMobile(): boolean {
    return window.innerWidth <= 768;
  }
}
