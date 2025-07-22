import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {AccountService} from "../../_services/account.service";
import {Router} from '@angular/router';
import {NavService} from "../../_services/nav.service";
import {take} from "rxjs";
import {LoadingSpinnerComponent} from "../../shared/_component/loading-spinner/loading-spinner.component";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-oidc-callback',
  imports: [
    LoadingSpinnerComponent,
    TranslocoDirective
  ],
  templateUrl: './oidc-callback.component.html',
  styleUrl: './oidc-callback.component.scss'
})
export class OidcCallbackComponent implements OnInit {

  constructor(
    private accountService: AccountService,
    private router: Router,
    private navService: NavService,
    private readonly cdRef: ChangeDetectorRef,
  ) {
    this.navService.setNavVisibility(false);
  }

  ngOnInit(): void {
    this.accountService.currentUser$.pipe(take(1)).subscribe(user => {
      if (user) {
        this.navService.setNavVisibility(true);
        this.router.navigateByUrl('/home');
        this.cdRef.markForCheck();
      }
    });
  }

}
