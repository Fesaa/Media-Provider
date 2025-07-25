import {Component, computed, DestroyRef, inject, OnInit, signal} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {ActivatedRoute, Router} from "@angular/router";
import {takeUntilDestroyed} from '@angular/core/rxjs-interop';
import {AccountService} from "../../_services/account.service";
import {NavService} from "../../_services/nav.service";
import {ToastService} from "../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {LoadingSpinnerComponent} from "../../shared/_component/loading-spinner/loading-spinner.component";

@Component({
  selector: 'app-reset',
  imports: [
    ReactiveFormsModule,
    TranslocoDirective,
    LoadingSpinnerComponent
  ],
  templateUrl: './reset.component.html',
  styleUrl: './reset.component.scss'
})
export class ResetComponent implements OnInit {
  private readonly route = inject(ActivatedRoute);
  private readonly accountService = inject(AccountService);
  private readonly navService = inject(NavService);
  private readonly router = inject(Router);
  private readonly toastService = inject(ToastService);
  private readonly destroyRef = inject(DestroyRef);

  readonly key = signal<string>('');
  readonly isLoading = signal<boolean>(false);
  readonly showPassword = signal<boolean>(false);

  readonly resetForm: FormGroup = new FormGroup({
    password: new FormControl('', [
      Validators.required,
      Validators.minLength(6)
    ]),
  });

  readonly isFormValid = computed(() => this.resetForm.valid);
  readonly canSubmit = computed(() =>
    this.isFormValid() && !this.isLoading() && this.key().length > 0
  );

  ngOnInit(): void {
    this.navService.setNavVisibility(false);
    this.setupRouteParams();
  }

  private setupRouteParams(): void {
    this.route.queryParams
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe(params => {
        const keyParam = params['key'];

        if (!keyParam) {
          this.navigateToLogin();
          return;
        }

        this.key.set(keyParam);
      });
  }

  togglePasswordVisibility(): void {
    this.showPassword.update(current => !current);
  }

  reset(): void {
    if (!this.canSubmit()) {
      return;
    }

    this.isLoading.set(true);

    const model = {
      key: this.key(),
      password: this.resetForm.value.password,
    };

    this.accountService.resetPassword(model)
      .pipe(takeUntilDestroyed(this.destroyRef))
      .subscribe({
        next: () => {
          this.toastService.successLoco('Password reset successfully!');
          this.navigateToLogin();
        },
        error: (err) => {
          this.isLoading.set(false);
          const errorMessage = err?.error?.message || 'An error occurred while resetting your password';
          this.toastService.genericError(errorMessage);
        },
        complete: () => {
          this.isLoading.set(false);
        }
      });
  }

  private navigateToLogin(): void {
    this.router.navigateByUrl('/login');
  }
}
