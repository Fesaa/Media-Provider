import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {ActivatedRoute, Router} from "@angular/router";
import {AccountService} from "../../_services/account.service";
import {NavService} from "../../_services/nav.service";
import {ToastService} from "../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-reset',
    imports: [
        ReactiveFormsModule,
        TranslocoDirective
    ],
  templateUrl: './reset.component.html',
  styleUrl: './reset.component.scss'
})
export class ResetComponent implements OnInit {

  resetForm: FormGroup = new FormGroup({
    password: new FormControl('', [Validators.required]),
  });

  key = '';

  constructor(private route: ActivatedRoute,
              private accountService: AccountService,
              private navService: NavService,
              private router: Router,
              private readonly cdRef: ChangeDetectorRef,
              private toastService: ToastService,
  ) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(false);

    this.route.queryParams.subscribe(params => {
      const key = params['key']
      if (!key) {
        this.router.navigateByUrl('/login');
        this.cdRef.markForCheck()
        return;
      }

      this.key = key;
    })

  }

  reset() {
    const model = {
      key: this.key,
      password: this.resetForm.value.password,
    }

    this.accountService.resetPassword(model).subscribe({
      next: () => {
        this.router.navigateByUrl('/login');
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    });
  }

}
