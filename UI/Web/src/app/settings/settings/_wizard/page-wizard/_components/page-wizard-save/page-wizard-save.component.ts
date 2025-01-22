import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page} from "../../../../../../_models/page";
import {PageService} from "../../../../../../_services/page.service";
import {ToastrService} from "ngx-toastr";
import {Card} from "primeng/card";
import {Fieldset} from "primeng/fieldset";
import {FormControl, FormGroup, FormsModule} from "@angular/forms";
import {NgForOf} from "@angular/common";
import {Router} from "@angular/router";
import {DialogService} from "../../../../../../_services/dialog.service";

@Component({
  selector: 'app-page-wizard-save',
  imports: [
    Card,
    Fieldset,
    FormsModule,
    NgForOf,
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
    private toastr: ToastrService,
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

    if (! await this.dialogService.openDialog("Are you sure you want save? ")) {
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
        this.toastr.success("Success!")
        this.router.navigate(["/page"], {
          queryParams: {
            index: page.ID,
          }
        })
      },
      error: (error) => {
        this.toastr.error(`An error occurred:\n${error.error.message}`, "Error!")
      }
    })

  }

  dirsCheck(): boolean {
    if (this.page.dirs.length == 0) {
      this.toastr.error("You must provide at least one download directory");
      return false;
    }

    return true;
  }

  generalCheck(): boolean {
    if (this.page.title === '') {
      this.toastr.error("You most provide a title")
      return false;
    }

    if (this.page.providers.length == 0) {
      this.toastr.error("You most provide at least one provider")
      return false;
    }

    return true;
  }

  modifierCheck(): boolean {
    for (const mod of this.page.modifiers) {
      if (mod.key === '' || mod.title === '') {
        const title = mod.title === '' ? mod.key : mod.title;
        this.toastr.error("Ensure all modifiers have their key and title set", `Invalid modifier ${title}`);
        return false;
      }

      for (const val of mod.values) {
        if (val.key === '' || val.value === '') {
          this.toastr.error("Ensure all modifier values have their key and value set", `Invalid modifier ${mod.title}`)
          return false;
        }
      }
    }

    return true;
  }

}
