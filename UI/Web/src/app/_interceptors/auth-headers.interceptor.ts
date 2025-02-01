import {Injectable} from '@angular/core';
import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Observable} from 'rxjs';
import {switchMap, take} from 'rxjs/operators';
import {AccountService} from "../_services/account.service";

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private accountService: AccountService) {
  }

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return this.accountService.currentUser$.pipe(
      take(1),
      switchMap(user => {
        if (user) {
          const authReq = req.clone({
            setHeaders: {
              Authorization: `Bearer ${user.token}`
            }
          });
          return next.handle(authReq);
        }
        return next.handle(req);
      })
    );
  }
}
