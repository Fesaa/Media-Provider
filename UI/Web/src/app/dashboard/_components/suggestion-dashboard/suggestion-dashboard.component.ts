import {Component} from '@angular/core';
import {PageService} from "../../../_services/page.service";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {dropAnimation} from "../../../_animations/drop-animation";
import {Page} from "../../../_models/page";
import {MessageService} from "../../../_services/message.service";

@Component({
  selector: 'app-suggestion-dashboard',
  imports: [
    RouterLink,
    NgIcon
  ],
  templateUrl: './suggestion-dashboard.component.html',
  styleUrl: './suggestion-dashboard.component.css',
  animations: [dropAnimation]
})
export class SuggestionDashboardComponent {

  pages: Page[] = []

  constructor(protected pageService: PageService,
              private msgService: MessageService,
  ) {
    this.pageService.pages$.subscribe(pages => {
      this.pages = pages;
    });
  }

  loadDefault() {
    this.pageService.loadDefault().subscribe({
      next: () => {
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.msgService.error('Error', err.error.message);
      }
    })
  }

}
