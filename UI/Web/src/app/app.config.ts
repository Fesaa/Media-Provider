import {ApplicationConfig, importProvidersFrom, isDevMode, provideZoneChangeDetection} from '@angular/core';
import {provideRouter} from '@angular/router';

import {routes} from './app.routes';
import {HTTP_INTERCEPTORS, provideHttpClient, withInterceptorsFromDi} from "@angular/common/http";
import {AuthInterceptor} from "./_interceptors/auth-headers.interceptor";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {AuthRedirectInterceptor} from "./_interceptors/auth-redirect.interceptor";
import {CommonModule} from "@angular/common";
import {ContentTitlePipe} from "./_pipes/content-title.pipe";
import {provideAnimationsAsync} from '@angular/platform-browser/animations/async';
import {providePrimeNG} from "primeng/config";
import Aura from '@primeng/themes/aura';
import {ProviderNamePipe} from "./_pipes/provider-name.pipe";
import {MessageService} from "primeng/api";
import {SubscriptionExternalUrlPipe} from "./_pipes/subscription-external-url.pipe";
import {provideTransloco} from "@jsverse/transloco";
import {TranslocoLoaderImpl} from "./_services/transloco-loader";

export const appConfig: ApplicationConfig = {
  providers: [
    CommonModule,
    ContentTitlePipe,
    ProviderNamePipe,
    SubscriptionExternalUrlPipe,
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
    })
  ]
};
