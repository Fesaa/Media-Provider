import {computed, DestroyRef, inject, Injectable, signal} from '@angular/core';
import {OAuthErrorEvent, OAuthService} from "angular-oauth2-oidc";
import {from} from "rxjs";
import {HttpClient} from "@angular/common/http";
import {environment} from "../../environments/environment";
import {takeUntilDestroyed, toObservable} from "@angular/core/rxjs-interop";
import {APP_BASE_HREF} from "@angular/common";
import {ToastService} from "./toast.service";
import {Oidc} from "../_models/config";

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
  private readonly _loaded = signal(false);
  public readonly loaded = this._loaded.asReadonly();
  public readonly loaded$ = toObservable(this.loaded);

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

    this.config().subscribe(oidcSetting => {
      this._settings.set(oidcSetting);
      if (!oidcSetting.authority) {
        this._loaded.set(true);
        return
      }

      this.oauth2.configure({
        issuer: oidcSetting.authority,
        clientId: oidcSetting.clientId,
        // Require https in production unless localhost
        requireHttps: environment.production ? 'remoteOnly' : false,
        redirectUri: window.location.origin + this.baseUrl + "oidc/callback",
        postLogoutRedirectUri: window.location.origin + this.baseUrl + "login",
        showDebugInformation: !environment.production,
        responseType: 'code',
        scope: "openid profile email roles offline_access",
        // Not all OIDC providers follow this nicely
        strictDiscoveryDocumentValidation: false,
        useSilentRefresh: false,
      });
      this.oauth2.setupAutomaticSilentRefresh();

      from(this.oauth2.loadDiscoveryDocumentAndTryLogin()).subscribe({
        next: _ => {
          this._loaded.set(true);

          if (!this.hasValidAccessToken() && this.oauth2.getRefreshToken()) {
            this.oauth2.refreshToken().catch(e => console.error(e));
          }
        },
        error: error => {
          console.log(error);
          this.toastR.errorLoco("oidc.error-loading-info")
        }
      });
    })
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

  config() {
    return this.httpClient.get<Oidc>(this.apiBaseUrl + "config/oidc");
  }

  get token() {
    return this.oauth2.getAccessToken();
  }

  hasValidAccessToken() {
    return this.oauth2.hasValidAccessToken();
  }

}
