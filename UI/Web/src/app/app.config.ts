import {
  ApplicationConfig,
  importProvidersFrom,
  inject,
  isDevMode,
  provideAppInitializer,
  provideZoneChangeDetection
} from '@angular/core';
import {provideRouter} from '@angular/router';

import {routes} from './app.routes';
import {HTTP_INTERCEPTORS, provideHttpClient, withInterceptorsFromDi} from "@angular/common/http";
import {AuthInterceptor} from "./_interceptors/auth-headers.interceptor";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {AuthRedirectInterceptor} from "./_interceptors/auth-redirect.interceptor";
import {APP_BASE_HREF, CommonModule, PlatformLocation} from "@angular/common";
import {ContentTitlePipe} from "./_pipes/content-title.pipe";
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {providePrimeNG} from "primeng/config";
import Aura from '@primeng/themes/aura';
import {ProviderNamePipe} from "./_pipes/provider-name.pipe";
import {MessageService} from "primeng/api";
import {SubscriptionExternalUrlPipe} from "./_pipes/subscription-external-url.pipe";
import {provideTransloco} from "@jsverse/transloco";
import {TranslocoLoaderImpl} from "./_services/transloco-loader";
import {provideOAuthClient} from "angular-oauth2-oidc";
import {OidcEvents, OidcService} from "./_services/oidc.service";
import {ToastService} from "./_services/toast.service";
import {AccountService} from './_services/account.service';
import {NavService} from "./_services/nav.service";
import {catchError, filter, firstValueFrom, Observable, of, switchMap, tap, timeout} from "rxjs";
import {User} from './_models/user';

function getBaseHref(platformLocation: PlatformLocation): string {
  return platformLocation.getBaseHrefFromDOM();
}

function setupOidcListener(oidcService: OidcService, accountService: AccountService, navService: NavService) {
  return oidcService.events$.pipe(
    filter(event => event.type === OidcEvents.TokenRefreshed),
    switchMap(() => syncOidcUser(oidcService, accountService, navService))
  ).subscribe();
}


function syncOidcUser(oidcService: OidcService, accountService: AccountService, navService: NavService): Observable<User> {
  const currentUser = accountService.currentUserSignal();
  const inStorage = accountService.getUserFromLocalStorage() !== undefined;

  return accountService.loginByToken(oidcService.token).pipe(
    tap(() => {
      navService.handleLogin(!currentUser && !inStorage);
    }),
    catchError(err => {
      console.error("Failed to sync OIDC user:", err);
      throw err;
    })
  );
}

function preLoadOidcAndUser() {
  const oidc = inject(OidcService);
  const toastr = inject(ToastService);
  const accountService = inject(AccountService);
  const navService = inject(NavService);

  return firstValueFrom(oidc.setupOidc().pipe(
    switchMap((isConfigured) => {
      if (!isConfigured) return of(null);

      return oidc.refreshTokenIfAvailable().pipe(
        switchMap(tokenRefreshed => {
          if (!tokenRefreshed) return of(null);

          return syncOidcUser(oidc, accountService, navService);
        })
      );
    }),
    tap(user => {
      if (!user) accountService.setCurrentUser(accountService.getUserFromLocalStorage());
    }),
    tap(() => setupOidcListener(oidc, accountService, navService)),
    timeout(2000),
    catchError(err => {
      console.error("OIDC setup failed:", err);
      if (err.name === 'TimeoutError') {
        toastr.errorLoco('errors.oidc.timeout');
      } else {
        toastr.errorLoco('errors.generic');
      }

      return of(null);
    }),
  )).then(() => void 0);
}

export const appConfig: ApplicationConfig = {
  providers: [
    CommonModule,
    ContentTitlePipe,
    ProviderNamePipe,
    SubscriptionExternalUrlPipe,
    provideOAuthClient(),
    provideZoneChangeDetection({eventCoalescing: true}),
    provideRouter(routes),
    {provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true},
    {provide: HTTP_INTERCEPTORS, useClass: AuthRedirectInterceptor, multi: true},
    provideHttpClient(withInterceptorsFromDi()),
    importProvidersFrom(BrowserAnimationsModule), provideAnimationsAsync(),
    providePrimeNG({
      theme: {
        preset: Aura
      }
    }),
    MessageService,
    provideTransloco({
      config: {
        availableLangs: ['en'],
        defaultLang: 'en',
        missingHandler: {
          useFallbackTranslation: true,
          allowEmpty: true,
        },
        reRenderOnLangChange: true,
        prodMode: !isDevMode(),
      },
      loader: TranslocoLoaderImpl,
    }),
    {
      provide: APP_BASE_HREF,
      useFactory: getBaseHref,
      deps: [PlatformLocation]
    },
    provideAppInitializer(() => preLoadOidcAndUser()),
  ]
};
