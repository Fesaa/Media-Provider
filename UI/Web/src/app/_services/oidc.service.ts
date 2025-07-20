import {computed, DestroyRef, inject, Injectable, signal} from '@angular/core';
import {OAuthErrorEvent, OAuthService} from "angular-oauth2-oidc";
import {catchError, from, map, Observable, of} from "rxjs";
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {takeUntilDestroyed, toObservable} from "@angular/core/rxjs-interop";
import {APP_BASE_HREF} from "@angular/common";
import {ToastService} from "./toast.service";
import {Oidc} from "../_models/config";
import {switchMap, tap} from "rxjs/operators";

/**
 * Enum mirror of angular-oauth2-oidc events which are used in Kavita
 */
export enum OidcEvents {
  /**
   * Fired on token refresh, and when the first token is recieved
   */
  TokenRefreshed = "token_refreshed"
}

@Injectable({
  providedIn: 'root'
})
export class OidcService {

  private readonly oauth2 = inject(OAuthService);
  private readonly httpClient = inject(HttpClient);
  private readonly destroyRef = inject(DestroyRef);
  private readonly toastR = inject(ToastService);

  protected readonly baseUrl = inject(APP_BASE_HREF);
  apiBaseUrl = environment.apiUrl;

  public events$ = this.oauth2.events;

  /**
   * True when the OIDC discovery document has been loaded, and login tried. Or no OIDC has been set up
   */
  private readonly loaded = signal(false);

  public readonly inUse = computed(() => {
    const loaded = this.loaded();
    const settings = this.settings();
    return loaded && settings && settings.authority.trim() !== '';
  });

  /**
   * Public OIDC settings
   */
  private readonly _settings = signal<Oidc | undefined>(undefined);
  public readonly settings = this._settings.asReadonly();

  constructor() {
    window.addEventListener('online', this.tryRefreshOnOnline)

    this.oauth2.setStorage(localStorage);

    // log events in dev
    if (!environment.production) {
      this.oauth2.events.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(event => {
        if (event instanceof OAuthErrorEvent) {
          console.error('OAuthErrorEvent:', event);
        } else {
          console.debug('OAuthEvent:', event);
        }
      });
    }
  }

  /**
   * Retrieves OIDC config and sets up OAuth service
   */
  setupOidc(): Observable<boolean> {
    return this.getPublicOidcConfig().pipe(
      switchMap((oidcSettings) => {
        this._settings.set(oidcSettings);

        if (!oidcSettings.authority) {
          this.loaded.set(true);
          return of(false);
        }

        return this.setupOAuthService(oidcSettings);
      }),
      tap(() => this.loaded.set(true))
    );
  }

  /**
   * Attempts to refresh the token if available
   * Returns observable that completes when refresh is done (success or failure)
   */
  refreshTokenIfAvailable(): Observable<boolean> {
    if (!this.oauth2.getRefreshToken()) {
      return of(false);
    }

    return from(this.oauth2.refreshToken()).pipe(
      map(() => true),
      catchError(err => {
        console.error("Failed to refresh token on startup", err);
        return of(false);
      })
    );
  }

  /**
   * Sets up the OAuthService, and loads the discovery document
   */
  setupOAuthService(oidcSettings: Oidc) {
    this.oauth2.configure({
      issuer: oidcSettings.authority,
      clientId: oidcSettings.clientId,
      // Require https in production unless localhost
      requireHttps: environment.production ? 'remoteOnly' : false,
      redirectUri: window.location.origin + this.baseUrl + "oidc/callback",
      postLogoutRedirectUri: window.location.origin + this.baseUrl + "login",
      showDebugInformation: !environment.production,
      responseType: 'code',
      scope: "openid profile email roles offline_access",
      // Not all OIDC providers follow this nicely
      strictDiscoveryDocumentValidation: false,
    });
    this.oauth2.setupAutomaticSilentRefresh();

    return from(this.oauth2.loadDiscoveryDocumentAndTryLogin());
  }

  tryRefreshOnOnline() {
    if (this.oauth2.hasValidAccessToken()) return;

    if (!this.oauth2.getRefreshToken()) return;

    this.oauth2.refreshToken().catch(e => console.error(e));
  }


  login() {
    this.oauth2.initLoginFlow();
  }

  logout() {
    if (this.token) {
      this.oauth2.logOut(true,);
    }
  }

  getPublicOidcConfig() {
    return this.httpClient.get<Oidc>(this.apiBaseUrl + "config/oidc");
  }

  get token() {
    return this.oauth2.getAccessToken();
  }

  hasValidAccessToken() {
    return this.oauth2.hasValidAccessToken();
  }

}
