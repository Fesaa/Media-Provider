import {ApplicationConfig, importProvidersFrom, provideZoneChangeDetection} from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import {HTTP_INTERCEPTORS, provideHttpClient, withInterceptorsFromDi} from "@angular/common/http";
import {AuthInterceptor} from "./_interceptors/auth-headers.interceptor";
import {BrowserAnimationsModule} from "@angular/platform-browser/animations";
import {AuthRedirectInterceptor} from "./_interceptors/auth-redirect.interceptor";
import {NgIconsModule} from "@ng-icons/core";
import {
  heroChevronDoubleRight,
  heroChevronUp,
  heroChevronDown,
  heroArrowDownTray,
  heroTrash,
  heroArrowUp,
  heroFolder,
  heroDocument,
  heroSquare3Stack3d,
  heroClipboard,
  heroPlus,
  heroAdjustmentsHorizontal,
  heroServerStack,
  heroPlusCircle,
  heroMinus,
  heroXMark,
  heroChevronLeft,
  heroArrowPath,
  heroEye,
  heroEyeSlash,
  heroChevronRight,
  heroUser,
  heroCheckCircle,
  heroXCircle, heroPencil,
} from '@ng-icons/heroicons/outline';
import {CommonModule} from "@angular/common";
import {provideToastr} from "ngx-toastr";
import {ContentTitlePipe} from "./_pipes/content-title.pipe";
import { provideAnimationsAsync } from '@angular/platform-browser/animations/async';

export const appConfig: ApplicationConfig = {
  providers: [
    CommonModule,
    provideToastr({
      countDuplicates: true,
      preventDuplicates: true,
      maxOpened: 5,
      resetTimeoutOnDuplicate: true,
      includeTitleDuplicates: true,
      progressBar: true,
      positionClass: 'toast-bottom-right',
    }),
    ContentTitlePipe,
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes),
    { provide: HTTP_INTERCEPTORS, useClass: AuthInterceptor, multi: true },
    { provide: HTTP_INTERCEPTORS, useClass: AuthRedirectInterceptor, multi: true },
    provideHttpClient(withInterceptorsFromDi()),
    importProvidersFrom(BrowserAnimationsModule, NgIconsModule.withIcons({
      heroChevronDoubleRight,
      heroChevronUp,
      heroChevronDown,
      heroArrowDownTray,
      heroTrash,
      heroArrowUp,
      heroFolder,
      heroDocument,
      heroSquare3Stack3d,
      heroClipboard,
      heroPlus,
      heroAdjustmentsHorizontal,
      heroServerStack,
      heroPlusCircle,
      heroMinus,
      heroXMark,
      heroChevronLeft,
      heroArrowPath,
      heroEye,
      heroEyeSlash,
      heroChevronRight,
      heroUser,
      heroCheckCircle,
      heroXCircle,
      heroPencil
    })), provideAnimationsAsync()
  ]
};
