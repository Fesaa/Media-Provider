import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {AccountService} from "../../_services/account.service";
import {Router} from "@angular/router";
import {Observable, take} from "rxjs";
import {AuthGuard} from "../../_guards/auth.guard";
import {NavService} from "../../_services/nav.service";
import {PageService} from "../../_services/page.service";
import {User} from "../../_models/user";
import {ToastService} from "../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";

@Component({
  selector: 'app-login',
  imports: [
    ReactiveFormsModule,
    TranslocoDirective,
    TitleCasePipe
  ],
  templateUrl: './user-login.component.html',
  styleUrl: './user-login.component.css'
})
export class UserLoginComponent implements OnInit {

  loginForm: FormGroup = new FormGroup({
    username: new FormControl("", [Validators.required]),
    password: new FormControl('', [Validators.required]),
    remember: new FormControl(false),
  });

  isSubmitting = false;
  isLoaded = false;

  hasAccount = false;

  constructor(private accountService: AccountService,
              private router: Router,
              private readonly cdRef: ChangeDetectorRef,
              private navService: NavService,
              private toastService: ToastService,
              private pageService: PageService,
  ) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(false);
    this.accountService.currentUser$.pipe(take(1)).subscribe(user => {
      if (user) {
        this.router.navigateByUrl('/home');
        this.cdRef.markForCheck()
        return;
      }

      this.accountService.anyUserExists().subscribe(check => {
        this.isLoaded = true;
        this.hasAccount = check;
        this.cdRef.markForCheck();
      })

    });
  }

  login() {
    const model = this.loginForm.getRawValue();
    this.isSubmitting = true;

    let obs: Observable<User>;
    if (this.hasAccount) {
      obs = this.accountService.login(model);
    } else {
      obs = this.accountService.register(model);
    }

    obs.subscribe({
      next: () => {
        this.loginForm.reset();
        this.pageService.refreshPages();
        this.navService.setNavVisibility(true);

        const pageResume = localStorage.getItem(AuthGuard.urlKey);
        localStorage.setItem(AuthGuard.urlKey, '');
        if (pageResume && pageResume != '/login') {
          this.router.navigateByUrl(pageResume);
        } else {
          this.router.navigateByUrl('/home');
        }

        this.isSubmitting = false;
        this.cdRef.markForCheck()
      }, error: (_) => {
        this.toastService.errorLoco("login.toasts.login-failed");
        this.isSubmitting = false;
        this.cdRef.markForCheck()
      }
    })

  }


}
