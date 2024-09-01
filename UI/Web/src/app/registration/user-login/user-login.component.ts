import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {AccountService} from "../../_services/account.service";
import {Router} from "@angular/router";
import {take} from "rxjs";
import {AuthGuard} from "../../_guards/auth.guard";
import {NavService} from "../../_services/nav.service";

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [
    ReactiveFormsModule
  ],
  templateUrl: './user-login.component.html',
  styleUrl: './user-login.component.css'
})
export class UserLoginComponent implements OnInit {

  loginForm: FormGroup = new FormGroup({
    password: new FormControl('', [Validators.required]),
    remember: new FormControl(false),
  });

  isSubmitting = false;
  isLoaded = false;

  constructor(private accountService: AccountService,
              private router: Router,
              private readonly cdRef: ChangeDetectorRef,
              private navService: NavService
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
      this.isLoaded = true;
      this.cdRef.markForCheck()
    });
  }

  login() {
    const model = this.loginForm.getRawValue();
    this.isSubmitting = true;

    this.accountService.login(model).subscribe(() => {
      this.loginForm.reset();

      const pageResume = localStorage.getItem(AuthGuard.urlKey);
      localStorage.setItem(AuthGuard.urlKey, '');
      if (pageResume && pageResume != '/login') {
        this.router.navigateByUrl(pageResume);
      } else {
        this.router.navigateByUrl('/home');
      }

      this.isSubmitting = false;
      this.cdRef.markForCheck()
    }, err => {
      console.error(err);
      this.isSubmitting = false;
      this.cdRef.markForCheck()
    })
  }


}
