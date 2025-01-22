import {ChangeDetectorRef, Component, HostListener, OnInit} from '@angular/core';
import {Page} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {ConfigService} from "../../../../_services/config.service";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {ToastrService} from "ngx-toastr";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {DialogService} from "../../../../_services/dialog.service";
import {FormBuilder, ReactiveFormsModule} from "@angular/forms";
import {hasPermission, Perm, User} from "../../../../_models/user";
import {AccountService} from "../../../../_services/account.service";

@Component({
    selector: 'app-pages-settings',
    imports: [
        RouterLink,
        NgIcon,
        ReactiveFormsModule,
    ],
    templateUrl: './pages-settings.component.html',
    styleUrl: './pages-settings.component.css',
    animations: [dropAnimation]
})
export class PagesSettingsComponent {

  user: User | null = null;
  pages: Page[] = []

  constructor(private configService: ConfigService,
              private toastR: ToastrService,
              private pageService: PageService,
              private dialogService: DialogService,
              private fb: FormBuilder,
              private cdRef: ChangeDetectorRef,
              private accountService: AccountService,
  ) {
    this.configService.getConfig().subscribe();
    this.pageService.pages$.subscribe(pages => this.pages = pages);
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.user = user;
      }
    });
  }


  async remove(page: Page) {
    if (!await this.dialogService.openDialog('Are you sure you want to remove this page?')) {
      return;
    }

   this.pageService.removePage(page.ID).subscribe({
      next: () => {
        this.toastR.success(`${page.title} removed`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

  moveUp(index: number) {
    const page1 = this.pages[index];
    const page2 = this.pages[index-1];
    this.swap(page1, page2);
  }

  moveDown(index: number) {
    const page1 = this.pages[index];
    const page2 = this.pages[index+1];
    this.swap(page1, page2);
  }

  swap(page1: Page, page2: Page) {
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.toastR.success(`Swapped ${page1.title} and ${page2.title}`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.messsage, 'Error');
      }
    });
  }


  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
}
