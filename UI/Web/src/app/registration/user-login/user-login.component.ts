import {ChangeDetectorRef, Component, computed, effect, inject, OnInit, signal} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {AccountService} from "../../_services/account.service";
import {ActivatedRoute, Router} from "@angular/router";
import {forkJoin, Observable, take} from "rxjs";
import {NavService} from "../../_services/nav.service";
import {User} from "../../_models/user";
import {ToastService} from "../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";
import {SettingsService} from "../../_services/settings.service";
import {tap} from "rxjs/operators";
import {Oidc} from "../../_models/config";

@Component({
  selector: 'app-login',
  imports: [
    ReactiveFormsModule,
    TranslocoDirective,
    TitleCasePipe
  ],
  templateUrl: './user-login.component.html',
  styleUrl: './user-login.component.scss'
})
export class UserLoginComponent implements OnInit {

  private readonly route = inject(ActivatedRoute);
  private readonly accountService = inject(AccountService);
  private readonly router = inject(Router);
  private readonly cdRef = inject(ChangeDetectorRef);
  private readonly navService = inject(NavService);
  private readonly toastService = inject(ToastService);
  private readonly settingsService = inject(SettingsService);

  loginForm: FormGroup = new FormGroup({
    username: new FormControl("", [Validators.required]),
    password: new FormControl('', [Validators.required]),
    remember: new FormControl(false),
  });
  /**
   * If there are no admins on the server, this will enable the registration to kick in.
   */
  firstTimeFlow = signal(true);
  /**
   * Used for first time the page loads to ensure no flashing
   */
  isLoaded = signal(false);
  isSubmitting = signal(false);
  oidcConfig = signal<Oidc | undefined>(undefined);
  /**
   * undefined until query params are read
   */
  skipAutoLogin = signal<boolean | undefined>(undefined);
  /**
   * Display the login form, regardless if the password authentication is disabled (admins can still log in)
   * Set from query
   */
  forceShowPasswordLogin = signal(false);
  /**
   * Display the login form
   */
  showPasswordLogin = computed(() => {
    const loaded = this.isLoaded();
    const config = this.oidcConfig();
    const force = this.forceShowPasswordLogin();
    if (force) return true;

    return loaded && config && !config.disablePasswordLogin;
  });

  constructor() {
    this.navService.setNavVisibility(false);

    effect(() => {
      const skipAutoLogin = this.skipAutoLogin();
      const oidcConfig = this.oidcConfig();
      if (!oidcConfig || skipAutoLogin === undefined) return;

      if (oidcConfig.autoLogin && !skipAutoLogin) {
        window.location.href = '/oidc/login'
      }
    });
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(false);

    const user = this.accountService.currentUserSignal();
    if (user) {
      this.router.navigateByUrl('/home');
      return;
    }

    forkJoin([
      this.accountService.anyUserExists(),
      this.settingsService.getPublicOidcConfig(),
    ]).pipe(
      tap(([check, oidcConfig]) => {
        this.firstTimeFlow.set(!check);
        this.oidcConfig.set(oidcConfig);
        this.isLoaded.set(true);
      })
    ).subscribe()

    this.route.queryParamMap.subscribe(params => {
      this.skipAutoLogin.set(params.get('skipAutoLogin') === 'true')
      this.forceShowPasswordLogin.set(params.get('forceShowPassword') === 'true');
    });
  }

  login() {
    const model = this.loginForm.getRawValue();
    this.isSubmitting.set(true);

    let obs: Observable<User>;
    if (this.firstTimeFlow()) {
      obs = this.accountService.register(model);
    } else {
      obs = this.accountService.login(model);
    }

    obs.subscribe({
      next: () => {
        this.loginForm.reset();
        this.navService.handleLogin();

        this.isSubmitting.set(false);
        this.cdRef.markForCheck();
      },
      error: (_) => {
        this.toastService.errorLoco("login.toasts.login-failed");
        this.isSubmitting.set(false);
        this.cdRef.markForCheck();
      }
    })

  }


}
