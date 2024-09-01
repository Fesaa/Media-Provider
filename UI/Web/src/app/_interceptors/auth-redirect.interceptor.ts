import {HttpEvent, HttpHandler, HttpInterceptor, HttpRequest} from '@angular/common/http';
import {catchError, Observable} from 'rxjs';
import {AccountService} from "../_services/account.service";
import {Injectable} from "@angular/core";

@Injectable()
export class AuthRedirectInterceptor implements HttpInterceptor {

  constructor(private accountService: AccountService) {
  }

  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    return next.handle(req).pipe(
      catchError(err => {
        if (err.status === 401) {
          this.accountService.logout();
        }
        throw err;
      })
    )
  }

}
