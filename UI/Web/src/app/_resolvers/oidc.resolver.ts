import { inject, Injectable } from '@angular/core';
import {ActivatedRouteSnapshot, Resolve, ResolveFn, RouterStateSnapshot} from '@angular/router';
import {OidcService} from "../_services/oidc.service";
import {catchError, filter, Observable, of, take, timeout} from 'rxjs';
import {ToastService} from "../_services/toast.service";

@Injectable({
  providedIn: 'root'
})
export class OidcResolver implements Resolve<any> {

  private oidcService = inject(OidcService);
  private toastR = inject(ToastService);

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<any> {
    return this.oidcService.loaded$.pipe(
      filter(value => value),
      take(1),
      timeout(5000),
      catchError(err => {
        console.log(err);
        this.toastR.errorLoco("oidc.timeout");
        return of(true);
      }));
  }
}
