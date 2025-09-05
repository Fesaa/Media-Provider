import {Component, inject} from '@angular/core';
import {PageService} from "../../../_services/page.service";
import {RouterLink} from "@angular/router";
import {dropAnimation} from "../../../_animations/drop-animation";
import {Page} from "../../../_models/page";
import {ToastService} from "../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";

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

  protected readonly pageService = inject(PageService);
  private readonly toastService = inject(ToastService);

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
