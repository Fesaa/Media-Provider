import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page} from "../../../../../../_models/page";
import {PageService} from "../../../../../../_services/page.service";
import {Card} from "primeng/card";
import {Fieldset} from "primeng/fieldset";
import {FormControl, FormGroup, FormsModule} from "@angular/forms";
import {NgForOf, TitleCasePipe} from "@angular/common";
import {Router} from "@angular/router";
import {DialogService} from "../../../../../../_services/dialog.service";
import {ProviderNamePipe} from "../../../../../../_pipes/provider-name.pipe";
import {ToastService} from "../../../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-page-wizard-save',
  imports: [
    Card,
    Fieldset,
    FormsModule,
    NgForOf,
    ProviderNamePipe,
    TranslocoDirective,
    TitleCasePipe,
  ],
  templateUrl: './page-wizard-save.component.html',
  styleUrl: './page-wizard-save.component.css'
})
export class PageWizardSaveComponent {

  @Input({required: true}) page!: Page;
  @Output() back: EventEmitter<void> = new EventEmitter();

  formGroup: FormGroup = new FormGroup({
    query: new FormControl(),
    dir: new FormControl(),
    customDir: new FormControl(),
  });

  constructor(
    private pageService: PageService,
    private toastService: ToastService,
    private router: Router,
    private dialogService: DialogService,
  ) {
  }

  async save() {

    if (!this.generalCheck()) {
      return;
    }

    if (!this.dirsCheck()) {
      return;
    }

    if (!this.modifierCheck()) {
      return;
    }

    if (!await this.dialogService.openDialog("settings.pages.wizard.confirm-save")) {
      return;
    }


    let obs;
    if (this.page.ID === 0) {
      obs = this.pageService.new(this.page);
    } else {
      obs = this.pageService.update(this.page);
    }

    obs.subscribe({
      next: (page) => {
        this.toastService.successLoco("settings.pages.toasts.save.success");
        this.router.navigate(["/page"], {
          queryParams: {
            index: page.ID,
          }
        })
      },
      error: (error) => {
        this.toastService.errorLoco("settings.pages.toasts.save.error", {}, {msg: error.error.message});
      }
    })

  }

  dirsCheck(): boolean {
    if (this.page.dirs.length == 0) {
      this.toastService.errorLoco("settings.pages.toasts.dir-required");
      return false;
    }

    return true;
  }

  generalCheck(): boolean {
    if (this.page.title === '') {
      this.toastService.errorLoco("settings.pages.toasts.name-required");
      return false;
    }

    if (this.page.providers.length == 0) {
      this.toastService.errorLoco("settings.pages.toasts.provider-required");
      return false;
    }

    return true;
  }

  modifierCheck(): boolean {
    for (const mod of this.page.modifiers) {
      if (mod.key === '' || mod.title === '') {
        const title = mod.title === '' ? mod.key : mod.title;
        this.toastService.errorLoco("settings.pages.toasts.invalid-modifier", {title: mod.title});
        return false;
      }

      for (const val of mod.values) {
        if (val.key === '' || val.value === '') {
          this.toastService.errorLoco("settings.pages.toasts.invalid-modifier-values", {title: mod.title});
          return false;
        }
      }
    }

    return true;
  }

}
