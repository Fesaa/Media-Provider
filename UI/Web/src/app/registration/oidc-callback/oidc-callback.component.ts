import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {AccountService} from "../../_services/account.service";
import {Router } from '@angular/router';
import {NavService} from "../../_services/nav.service";
import {take} from "rxjs";

@Component({
  selector: 'app-oidc-callback',
  imports: [],
  templateUrl: './oidc-callback.component.html',
  styleUrl: './oidc-callback.component.css'
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
