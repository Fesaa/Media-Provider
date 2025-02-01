import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {RouterLink} from "@angular/router";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {DialogService} from "../../../../../../_services/dialog.service";
import {NgForOf} from "@angular/common";
import {Button} from "primeng/button";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {Fieldset} from "primeng/fieldset";
import {MessageService} from "../../../../../../_services/message.service";

@Component({
  selector: 'app-page-wizard-dirs',
  imports: [
    Card,
    RouterLink,
    FloatLabel,
    InputText,
    ReactiveFormsModule,
    FormsModule,
    NgForOf,
    Button,
    IconField,
    InputIcon,
    Fieldset
  ],
  templateUrl: './page-wizard-dirs.component.html',
  styleUrl: './page-wizard-dirs.component.css'
})
export class PageWizardDirsComponent {

  @Input({required:true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();
  @Output() back: EventEmitter<void> = new EventEmitter();

  constructor(private dialogService: DialogService,
              private msgService: MessageService,
  ) {
  }

  nextCallback(): void {
    if (this.page.dirs.length == 0) {
      this.msgService.error("You must provide at least one download directory");
      return;
    }

    this.next.emit();
  }

  removeDir(index: number) {
    this.page.dirs.splice(index, 1);
  }

  async updateDir(index: number) {
    const newDir = await this.dialogService.openDirBrowser("");
    if (newDir === undefined) {
      return;
    }

    if (newDir === "") {
      this.msgService.warning("Cannot add empty directory.");
      return;
    }

    if (this.page.dirs.includes(newDir)) {
      this.msgService.warning("Not adding duplicate directory.");
      return;
    }

    if (index >= this.page.dirs.length) {
      this.page.dirs.push(newDir);
    } else {
      this.page.dirs[index] = newDir;
    }
  }

  async updateCustomDir() {
    const newDir = await this.dialogService.openDirBrowser("");
    if (newDir === undefined) {
      return;
    }

    this.page.custom_root_dir = newDir;
  }

}
