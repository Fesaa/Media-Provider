import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page} from "../../../../../../_models/page";
import {PageService} from "../../../../../../_services/page.service";
import {Card} from "primeng/card";
import {Fieldset} from "primeng/fieldset";
import {FormControl, FormGroup, FormsModule} from "@angular/forms";
import {NgForOf} from "@angular/common";
import {Router} from "@angular/router";
import {DialogService} from "../../../../../../_services/dialog.service";
import {ProviderNamePipe} from "../../../../../../_pipes/provider-name.pipe";
import {MessageService} from "../../../../../../_services/message.service";

@Component({
  selector: 'app-page-wizard-save',
  imports: [
    Card,
    Fieldset,
    FormsModule,
    NgForOf,
    ProviderNamePipe,
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
    private msgService: MessageService,
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

    if (!await this.dialogService.openDialog("Are you sure you want save? ")) {
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
        this.msgService.success("Success!")
        this.router.navigate(["/page"], {
          queryParams: {
            index: page.ID,
          }
        })
      },
      error: (error) => {
        this.msgService.error("Error!", `An error occurred:\n${error.error.message}`)
      }
    })

  }

  dirsCheck(): boolean {
    if (this.page.dirs.length == 0) {
      this.msgService.error("You must provide at least one download directory");
      return false;
    }

    return true;
  }

  generalCheck(): boolean {
    if (this.page.title === '') {
      this.msgService.error("You most provide a title")
      return false;
    }

    if (this.page.providers.length == 0) {
      this.msgService.error("You most provide at least one provider")
      return false;
    }

    return true;
  }

  modifierCheck(): boolean {
    for (const mod of this.page.modifiers) {
      if (mod.key === '' || mod.title === '') {
        const title = mod.title === '' ? mod.key : mod.title;
        this.msgService.error(`Invalid modifier ${title}`, "Ensure all modifiers have their key and title set");
        return false;
      }

      for (const val of mod.values) {
        if (val.key === '' || val.value === '') {
          this.msgService.error(`Invalid modifier ${mod.title}`, "Ensure all modifier values have their key and value set")
          return false;
        }
      }
    }

    return true;
  }

}
