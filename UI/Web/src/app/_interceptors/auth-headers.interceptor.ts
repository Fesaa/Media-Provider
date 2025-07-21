import {Injectable} from '@angular/core';
import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {Observable} from 'rxjs';
import {AccountService} from "../_services/account.service";

@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  constructor(private accountService: AccountService) {
  }

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    const user = this.accountService.currentUserSignal();
    if (user) {
      req = req.clone({
        setHeaders: {
          Authorization: `Bearer ${user.oidcToken ?? user.token}`
        }
      });
    }

    return next.handle(req);
  }
}
