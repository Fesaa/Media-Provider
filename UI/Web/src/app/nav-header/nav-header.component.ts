import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {PageService} from "../_services/page.service";
import {Page} from "../_models/page";
import {ActivatedRoute} from "@angular/router";
import {AsyncPipe} from "@angular/common";
import {AccountService} from "../_services/account.service";
import {NavService} from "../_services/nav.service";
import {dropAnimation} from "../_animations/drop-animation";
import {MenuItem} from "primeng/api";
import {Menubar} from "primeng/menubar";

@Component({
  selector: 'app-nav-header',
  imports: [
    AsyncPipe,
    Menubar
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
  accountItems: MenuItem[] | undefined;
  pageItems: MenuItem[] | undefined;

  constructor(private pageService: PageService,
              private route: ActivatedRoute,
              private cdRef: ChangeDetectorRef,
              protected accountService: AccountService,
              protected navService: NavService
  ) {

    this.pageService.pages$.subscribe(pages => {
      this.pages = pages;

      this.pageItems = this.pages.map(page => {
        return {
          label: page.title,
          routerLink: 'page',
          queryParams: {index: page.ID},
          icon: page.icon === '' ? undefined : 'pi ' + page.icon,
        }
      })
      this.pageItems = [{
        label: 'Home',
        routerLink: 'home',
        icon: 'pi pi-home',
      }, ...this.pageItems]


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

    this.accountService.currentUser$.subscribe(user => {
      if (!user) {
        return;
      }

      this.accountItems = [
        {
          label: user.name,
          icon: "pi pi-user",
          items: [
            {
              label: "Subscriptions",
              routerLink: "subscriptions",
              icon: "pi pi-wave-pulse"
            },
            {
              label: "Settings",
              routerLink: "settings",
              icon: "pi pi-cog",
            },
            {
              label: "Log out",
              command: () => {
                this.accountService.logout()
              },
              icon: "pi pi-sign-out"
            }
          ]
        }
      ];

      if (window.innerWidth <= 768) {
        this.accountItems = this.accountItems[0].items;
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
