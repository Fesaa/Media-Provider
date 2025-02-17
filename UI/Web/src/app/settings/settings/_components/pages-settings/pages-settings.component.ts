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
import {ToastService} from "../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-pages-settings',
  imports: [
    RouterLink,
    ReactiveFormsModule,
    Button,
    TableModule,
    Tooltip,
    TranslocoDirective,
  ],
  templateUrl: './pages-settings.component.html',
  styleUrl: './pages-settings.component.css',
  animations: [dropAnimation]
})
export class PagesSettingsComponent {

  user: User | null = null;
  pages: Page[] = []
  loading: boolean = true;
  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;

  constructor(private toastService: ToastService,
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
    if (!await this.dialogService.openDialog("settings.pages.confirm-delete", {title: page.title})) {
      return;
    }

    this.pageService.removePage(page.ID).subscribe({
      next: () => {
        this.toastService.successLoco("settings.pages.toasts.delete.success", {}, {title: page.title});
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
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
    return this.pages[this.pages.length - 1].ID === page.ID;
  }

  moveUp(page: Page) {
    const index = this.pages.indexOf(page);
    const other = this.pages[index - 1];
    this.swap(page, other);
  }

  moveDown(page: Page) {
    const index = this.pages.indexOf(page);
    const other = this.pages[index + 1];
    this.swap(page, other);
  }

  swap(page1: Page, page2: Page) {
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.toastService.successLoco("settings.pages.toasts.swapped.success", {},
          {page1: page1.title, page2: page2.title});
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    });
  }
}
