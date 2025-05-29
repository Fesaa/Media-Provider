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
import {CdkDrag, CdkDragDrop, CdkDragHandle, CdkDropList, moveItemInArray} from "@angular/cdk/drag-drop";
import {Card} from "primeng/card";
import {NgForOf} from "@angular/common";

@Component({
  selector: 'app-pages-settings',
  imports: [
    RouterLink,
    ReactiveFormsModule,
    Button,
    TableModule,
    Tooltip,
    TranslocoDirective,
    CdkDropList,
    Card,
    NgForOf,
    CdkDragHandle,
    CdkDrag,
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

  drop($event: CdkDragDrop<any, any>) {
    const page1 = this.pages[$event.previousIndex];
    const page2 = this.pages[$event.currentIndex];

    // Assume no error will occur
    moveItemInArray(this.pages, $event.previousIndex, $event.currentIndex);
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.pageService.refreshPages();
      },
      error: (err) => {
        // Could not move, set back
        moveItemInArray(this.pages, $event.currentIndex, $event.previousIndex)
        this.toastService.genericError(err.error.message);
      }
    });
  }
}
