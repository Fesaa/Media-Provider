import {Component, computed, inject} from '@angular/core';
import {PageService} from "../../../_services/page.service";
import {RouterLink} from "@angular/router";
import {dropAnimation} from "../../../_animations/drop-animation";
import {Page} from "../../../_models/page";
import {ToastService} from "../../../_services/toast.service";
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {ModalService} from "../../../_services/modal.service";
import {ManualContentAddModalComponent} from "../manual-content-add-modal/manual-content-add-modal.component";
import {DefaultModalOptions} from "../../../_models/default-modal-options";

interface Option {
  ID: number,
  title: string,
  action?: () => void,
}

@Component({
  selector: 'app-suggestion-dashboard',
  imports: [
    RouterLink,
    TranslocoDirective,
  ],
  templateUrl: './suggestion-dashboard.component.html',
  styleUrl: './suggestion-dashboard.component.scss',
  animations: [dropAnimation]
})
export class SuggestionDashboardComponent {

  private readonly pageService = inject(PageService);
  private readonly modalService = inject(ModalService);
  private readonly toastService = inject(ToastService);

  options = computed(() => {
    const options: Option[] = this.pageService.pages().map(p => {
      return {
        ID: p.ID,
        title: p.title,
      };
    });

    if (options.length === 0) {
      return options // Ensure we display the load defaults
    }

    options.push({
      ID: -1,
      title: translate('dashboard.manual-add.title'),
      action: this.manualAdd.bind(this),
    })

    return options;
  })

  manualAdd() {
    this.modalService.open(ManualContentAddModalComponent, DefaultModalOptions);
  }

  loadDefault() {
    this.pageService.loadDefault().subscribe({
      next: () => {
        this.pageService.refreshPages().subscribe();
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

}
