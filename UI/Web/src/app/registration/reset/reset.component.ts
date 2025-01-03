import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {ActivatedRoute, Router} from "@angular/router";
import {AccountService} from "../../_services/account.service";
import {NavService} from "../../_services/nav.service";
import {ToastrService} from "ngx-toastr";

@Component({
    selector: 'app-reset',
    imports: [
        ReactiveFormsModule
    ],
    templateUrl: './reset.component.html',
    styleUrl: './reset.component.css'
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
              private toastR: ToastrService,
  ) {}

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
        this.toastR.error(`Failed to reset password: ${err.error.message}`, "Error");
      }
    });
  }

}
