import {DestroyRef, inject, Injectable, signal} from '@angular/core';
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



  private readonly _currentUser = signal<User | undefined>(undefined);
  public readonly currentUserSignal = this._currentUser.asReadonly();
  private currentUser: User | undefined;

  constructor(private httpClient: HttpClient, private router: Router) {
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

  getMe() {
    return this.httpClient.get<User>(this.baseUrl+"user/me").pipe(
      tap((user) => {
        this.setCurrentUser(user);
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
    this._currentUser.set(user);

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
    this._currentUser.set(undefined);
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

      this.setCurrentUser({
        ...this.currentUser,
        name: dto.name,
        roles: dto.roles,
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
