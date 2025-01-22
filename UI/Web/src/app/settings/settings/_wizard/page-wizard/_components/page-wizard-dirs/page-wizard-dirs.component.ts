import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {RouterLink} from "@angular/router";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {NgIcon} from "@ng-icons/core";
import {DialogService} from "../../../../../../_services/dialog.service";

@Component({
  selector: 'app-page-wizard-dirs',
  imports: [
    Card,
    RouterLink,
    FloatLabel,
    InputText,
    ReactiveFormsModule,
    FormsModule,
    NgIcon
  ],
  templateUrl: './page-wizard-dirs.component.html',
  styleUrl: './page-wizard-dirs.component.css'
})
export class PageWizardDirsComponent {

  @Input({required:true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();
  @Output() back: EventEmitter<void> = new EventEmitter();

  constructor(private dialogService: DialogService ) {
  }

  async updateCustomDir() {
    const newDir = await this.dialogService.openDirBrowser("");
    if (newDir === undefined) {
      return;
    }

    this.page.custom_root_dir = newDir;
  }

}
