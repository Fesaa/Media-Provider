import {Component} from '@angular/core';
import {Page} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {RouterLink} from "@angular/router";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {DialogService} from "../../../../_services/dialog.service";
import {ReactiveFormsModule} from "@angular/forms";
import {hasPermission, Perm, User} from "../../../../_models/user";
import {AccountService} from "../../../../_services/account.service";
import {Button} from "primeng/button";
import {TableModule} from "primeng/table";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from "../../../../_services/message.service";

@Component({
    selector: 'app-pages-settings',
  imports: [
    RouterLink,
    ReactiveFormsModule,
    Button,
    TableModule,
    Tooltip,
  ],
    templateUrl: './pages-settings.component.html',
    styleUrl: './pages-settings.component.css',
    animations: [dropAnimation]
})
export class PagesSettingsComponent {

  user: User | null = null;
  pages: Page[] = []
  loading: boolean = true;

  constructor(private msgService: MessageService,
              private pageService: PageService,
              private dialogService: DialogService,
              private accountService: AccountService,
  ) {
    this.pageService.pages$.subscribe(pages => {
      this.pages = pages
      this.loading = false;
    });
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.user = user;
      }
    });
  }


  async remove(page: Page) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove page > ${page.title}?`)) {
      return;
    }

   this.pageService.removePage(page.ID).subscribe({
      next: () => {
        this.msgService.success('Success', `${page.title} removed`);
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.msgService.error('Error', err.error.message);
      }
    });
  }

  isFirst(page: Page): boolean {
    if (this.pages.length == 0) {
      return false;
    }
    return this.pages[0].ID === page.ID;
  }

  isLast(page: Page): boolean {
    if (this.pages.length == 0) {
      return false;
    }
    return this.pages[this.pages.length-1].ID === page.ID;
  }

  moveUp(page: Page) {
    const index = this.pages.indexOf(page);
    const other = this.pages[index-1];
    this.swap(page, other);
  }

  moveDown(page: Page) {
    const index = this.pages.indexOf(page);
    const other = this.pages[index+1];
    this.swap(page, other);
  }

  swap(page1: Page, page2: Page) {
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.msgService.success('Success', `Swapped ${page1.title} and ${page2.title}`);
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.msgService.error('Error', err.error.messsage);
      }
    });
  }


  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
}
