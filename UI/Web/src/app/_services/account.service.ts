import {DestroyRef, inject, Injectable} from '@angular/core';
import {environment} from "../../environments/environment";
import {Observable, ReplaySubject, tap} from "rxjs";
import {User, UserDto} from "../_models/user";
import {HttpClient, HttpHeaders} from "@angular/common/http";
import {Router} from "@angular/router";
import {takeUntilDestroyed, toSignal} from "@angular/core/rxjs-interop";
import {PasswordReset} from "../_models/password_reset";
import {SignalRService} from "./signal-r.service";

@Injectable({
  providedIn: 'root'
})
export class AccountService {

  private readonly destroyRef = inject(DestroyRef);
  private readonly signalR = inject(SignalRService);

  baseUrl = environment.apiUrl;
  userKey = 'mp-user';


  private currentUserSource = new ReplaySubject<User | undefined>(1);
  public currentUser$ = this.currentUserSource.asObservable();
  public currentUserSignal = toSignal(this.currentUser$);
  private currentUser: User | undefined;

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

  login(model: { username: string, password: string, remember: boolean }): Observable<User> {
    return this.httpClient.post<User>(this.baseUrl + 'login', model).pipe(
      tap((user: User) => {
        if (user) {
          this.setCurrentUser(user)
        }
      }),
      takeUntilDestroyed(this.destroyRef)
    );
  }

  loginByToken(token: string): Observable<User> {
    const headers = new HttpHeaders({
      "Authorization": `Bearer ${token}`
    })
    return this.httpClient.get<User>(`${this.baseUrl}user/me`, {headers}).pipe(
      tap((user: User) => {
        if (user) {
          user.oidcToken = token;
          this.setCurrentUser(user)
        }
      }),
      takeUntilDestroyed(this.destroyRef)
    );
  }

  updateMe(model: {username: string, email: string}) {
    return this.httpClient.post<User>(`${this.baseUrl}user/me`, model).pipe(
      tap(() => {
        if (!this.currentUser) return;

        const user = {...this.currentUser};
        user.name = model.username;
        user.email = model.email;
        this.setCurrentUser(user);
      })
    );
  }

  updatePassword(model: {oldPassword: string, newPassword: string}) {
    return this.httpClient.post(`${this.baseUrl}user/password`, model)
  }

  register(model: { username: string, password: string, remember: boolean }): Observable<User> {
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

    if (user) {
      this.signalR.stopConnection()
        .then(() => this.signalR.startConnection(user));
    }
  }

  logout() {
    if (!this.currentUser) {
      return;
    }

    localStorage.removeItem(this.userKey);
    this.currentUser = undefined;
    this.currentUserSource.next(undefined);
    this.router.navigate(['/login']);
  }

  anyUserExists() {
    return this.httpClient.get<boolean>(this.baseUrl + 'any-user-exists');
  }

  all() {
    return this.httpClient.get<UserDto[]>(this.baseUrl + 'user/all');
  }

  updateOrCreate(dto: UserDto) {
    return this.httpClient.post<UserDto>(this.baseUrl + 'user/update', dto).pipe(tap(dto => {
      if (dto.id !== this.currentUser?.id) {
        return;
      }

      // TODO: This should be changed to refresh, with a refresh token. I think
      this.setCurrentUser({
        ...this.currentUser,
        name: dto.name,
        permissions: dto.permissions,
      })
    }))
  }

  delete(id: number) {
    return this.httpClient.delete(this.baseUrl + 'user/' + id);
  }

  generateReset(id: number) {
    return this.httpClient.post<PasswordReset>(this.baseUrl + 'user/reset/' + id, {})
  }

  resetPassword(model: { key: string, password: string }) {
    return this.httpClient.post(this.baseUrl + 'reset-password', model)
  }

  refreshApiKey() {
    return this.httpClient.get<{ ApiKey: string }>(this.baseUrl + 'user/refresh-api-key').pipe(tap(res => {
      this.currentUser!.apiKey = res.ApiKey
      this.setCurrentUser(this.currentUser)
    }));
  }

}
