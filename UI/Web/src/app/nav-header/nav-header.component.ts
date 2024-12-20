import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {PageService} from "../_services/page.service";
import {Page} from "../_models/page";
import {ActivatedRoute, RouterLink} from "@angular/router";
import {AsyncPipe, NgClass} from "@angular/common";
import {animate, query, sequence, stagger, state, style, transition, trigger} from "@angular/animations";
import {AccountService} from "../_services/account.service";
import {NavService} from "../_services/nav.service";
import {Observable} from "rxjs";
import {dropAnimation} from "../_animations/drop-animation";

@Component({
    selector: 'app-nav-header',
    imports: [
        RouterLink,
        NgClass,
        AsyncPipe
    ],
    templateUrl: './nav-header.component.html',
    styleUrl: './nav-header.component.css',
    animations: [dropAnimation]
})
export class NavHeaderComponent implements OnInit {

  isMenuOpen = false;
  index: number | undefined;
  path: string | undefined;

  pages: Page[] = [];

  constructor(private pageService: PageService,
              private route: ActivatedRoute,
              private cdRef: ChangeDetectorRef,
              protected accountService: AccountService,
              protected navService: NavService
  ) {

    this.pageService.pages$.subscribe(pages => {
      this.pages = pages;
      this.cdRef.detectChanges();
    });
  }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      const index = params['index'];
      if (index) {
        this.index = parseInt(index);
      } else {
        this.index = undefined;
      }
    })
  }

  clickMenu() {
    this.isMenuOpen = !this.isMenuOpen;
    this.cdRef.detectChanges();
  }

  mobileMenuState() {
    return this.isMenuOpen ? 'open' : 'closed';
  }

}
