import { Component } from '@angular/core';
import {Page} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {ConfigService} from "../../../../_services/config.service";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {ToastrService} from "ngx-toastr";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {ConfirmService} from "../../../../_services/confirm.service";

@Component({
  selector: 'app-pages-settings',
  standalone: true,
  imports: [
    RouterLink,
    NgIcon
  ],
  templateUrl: './pages-settings.component.html',
  styleUrl: './pages-settings.component.css',
  animations: [dropAnimation],
})
export class PagesSettingsComponent {

  pages: Page[] = []

  cooldown = false;
  selectedPage: Page | null = null;

  constructor(private configService: ConfigService,
              private toastR: ToastrService,
              private pageService: PageService,
              private confirmService: ConfirmService,
  ) {
    this.pageService.pages$.subscribe(pages => this.pages = pages);
  }

  setSelectedPage(page?: Page) {
    if (page === undefined) {
      page = {
        dirs: [],
        title: '',
        modifiers: {},
        custom_root_dir: '',
        provider: [],
      }
    }
    this.selectedPage = page;
    this.cooldown = true;
    setTimeout(() => this.cooldown = false, 700);
  }

  async remove(index: number) {
    const confirm = await this.confirmService.confirm(`Are you sure you want to remove ${this.pages[index].title}?`);
    if (!confirm) {
      console.log('cancelled');
      return
    }


   /* this.configService.removePage(index).subscribe({
      next: () => {
        const page = this.pages[index];
        this.toastR.success(`${page.title} removed`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });*/
  }

  moveUp(index: number) {
    this.configService.movePage(index, index - 1).subscribe({
      next: () => {
        const temp = this.pages[index];
        this.toastR.success(`${temp.title} moved up`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

  moveDown(index: number) {
    this.configService.movePage(index, index + 1).subscribe({
      next: () => {
        const temp = this.pages[index];
        this.toastR.success(`${temp.title} moved down`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

}
