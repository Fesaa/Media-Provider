import {Component, computed, inject, OnInit, signal} from '@angular/core';
import {Page} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {RouterLink} from "@angular/router";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {ReactiveFormsModule} from "@angular/forms";
import {hasPermission, Perm} from "../../../../_models/user";
import {AccountService} from "../../../../_services/account.service";
import {ToastService} from "../../../../_services/toast.service";
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {CdkDragDrop, CdkDragHandle, moveItemInArray} from "@angular/cdk/drag-drop";
import {TableComponent} from "../../../../shared/_component/table/table.component";
import {ModalService} from "../../../../_services/modal.service";
import {EditPageModalComponent} from "./_components/edit-page-modal/edit-page-modal.component";
import {DefaultModalOptions} from "../../../../_models/default-modal-options";

@Component({
  selector: 'app-pages-settings',
  imports: [
    RouterLink,
    ReactiveFormsModule,
    TranslocoDirective,
    CdkDragHandle,
    TableComponent,
  ],
  templateUrl: './pages-settings.component.html',
  styleUrl: './pages-settings.component.scss',
  animations: [dropAnimation]
})
export class PagesSettingsComponent implements OnInit {

  private readonly modalService = inject(ModalService);
  private readonly toastService = inject(ToastService);
  private readonly pageService = inject(PageService);
  private readonly accountService = inject(AccountService);

  user = computed(() => this.accountService.currentUserSignal());
  pages = signal<Page[]>([]);
  loading = signal(true);

  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;

  ngOnInit(): void {
    this.loadPages();
  }

  loadPages() {
    this.pageService.pages$.subscribe(pages => {
      this.pages.set(pages)
      this.loading.set(false);
    });
  }

  edit(page: Page | null) {
    const [modal, component] = this.modalService.open(EditPageModalComponent, DefaultModalOptions);
    component.page.set(page ?? {
      ID: -1,
      customRootDir: '',
      title: '',
      dirs: [],
      providers: [],
      modifiers: [],
      icon: '',
      sortValue: 0,
    });

    modal.result.then(() => this.loadPages());
  }

  async remove(page: Page) {
    if (!await this.modalService.confirm({
      question: translate("settings.pages.confirm-delete", {title: page.title})
    })) {
      return;
    }

    this.pageService.removePage(page.ID).subscribe({
      next: () => {
        this.toastService.successLoco("settings.pages.toasts.delete.success", {}, {title: page.title});
        this.pageService.refreshPages().subscribe();
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    });
  }

  drop($event: CdkDragDrop<any, any>) {
    const pages = [...this.pages()];

    const page1 = pages[$event.previousIndex];
    const page2 = pages[$event.currentIndex];

    // Assume no error will occur
    moveItemInArray(pages, $event.previousIndex, $event.currentIndex);
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.pageService.refreshPages().subscribe();
      },
      error: (err) => {
        // Could not move, set back
        moveItemInArray(pages, $event.currentIndex, $event.previousIndex)
        this.toastService.genericError(err.error.message);
      }
    }).add(() => this.pages.set(pages));
  }

  trackBy(idx: number, page: Page) {
    return `${page.ID}`
  }
}
