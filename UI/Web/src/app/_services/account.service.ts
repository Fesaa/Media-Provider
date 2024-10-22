import {DestroyRef, inject, Injectable} from '@angular/core';
import {environment} from "../../environments/environment";
import {Observable, ReplaySubject, tap} from "rxjs";
import {User} from "../_models/user";
import {HttpClient} from "@angular/common/http";
import {Router} from "@angular/router";
import {takeUntilDestroyed} from "@angular/core/rxjs-interop";

@Injectable({
  providedIn: 'root'
})
export class AccountService {

  private readonly destroyRef = inject(DestroyRef);

  baseUrl = environment.apiUrl;
  userKey = 'mp-user';

  private currentUserSource = new ReplaySubject<User | undefined>(1);
  private currentUser: User | undefined;
  public currentUser$ = this.currentUserSource.asObservable();

  constructor(private httpClient: HttpClient, private router: Router) {
    const user = localStorage.getItem(this.userKey);
    if (user) {
      this.currentUser = JSON.parse(user);
      this.currentUserSource.next(this.currentUser);
    } else {
      this.currentUser = undefined;
      this.currentUserSource.next(undefined);
    }
  }

  login(model: {username: string, password: string, remember: boolean}): Observable<User> {
    return this.httpClient.post<User>(this.baseUrl + 'login', model).pipe(
      tap((user: User) => {
        if (user) {
          this.setCurrentUser(user)
        }
      }),
      takeUntilDestroyed(this.destroyRef)
    );
  }

  register(model: {username: string, password: string, remember: boolean}): Observable<User> {
    return this.httpClient.post<User>(this.baseUrl + 'register', model).pipe(
      tap((user: User) => {
        if (user) {
          this.setCurrentUser(user)
        }
      }),
      takeUntilDestroyed(this.destroyRef)
    );
  }

  setCurrentUser(user?: User) {
    if (user) {
      localStorage.setItem(this.userKey, JSON.stringify(user));
    }

    this.currentUser = user;
    this.currentUserSource.next(user);
  }

  logout() {
    localStorage.removeItem(this.userKey);
    this.currentUser = undefined;
    this.currentUserSource.next(undefined);
    this.router.navigate(['/login']);
  }

  anyUserExists() {
    return this.httpClient.get<boolean>(this.baseUrl + 'any-user-exists');
  }
}
