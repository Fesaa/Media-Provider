import {ApplicationConfig, importProvidersFrom, provideZoneChangeDetection} from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import {HTTP_INTERCEPTORS, provideHttpClient, withInterceptorsFromDi} from "@angular/common/http";
import {AuthInterceptor} from "./_interceptors/auth-headers.interceptor";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {AuthRedirectInterceptor} from "./_interceptors/auth-redirect.interceptor";
import {NgIconsModule} from "@ng-icons/core";
import { heroChevronDoubleRight, heroChevronUp, heroChevronDown, heroArrowDownTray } from '@ng-icons/heroicons/outline';
import {CommonModule} from "@angular/common";

export const appConfig: ApplicationConfig = {
  providers: [
    CommonModule,
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes),
    { provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true },
    { provide: HTTP_INTERCEPTORS, useClass: AuthRedirectInterceptor, multi: true },
    provideHttpClient(withInterceptorsFromDi()),
    importProvidersFrom(BrowserAnimationsModule, NgIconsModule.withIcons({heroChevronDoubleRight, heroChevronUp, heroChevronDown, heroArrowDownTray}))
  ]
};
