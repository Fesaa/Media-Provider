import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Modifier, ModifierType, Page} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {Fieldset} from "primeng/fieldset";
import {FormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {InputText} from "primeng/inputtext";
import {Tooltip} from "primeng/tooltip";
import {FloatLabel} from "primeng/floatlabel";
import {Select} from "primeng/select";
import {NgForOf} from "@angular/common";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {DialogService} from "../../../../../../_services/dialog.service";
import {MessageService} from "../../../../../../_services/message.service";

@Component({
  selector: 'app-page-wizard-modifiers',
  imports: [
    Card,
    Fieldset,
    FormsModule,
    Button,
    InputText,
    Tooltip,
    FloatLabel,
    Select,
    NgForOf,
    IconField,
    InputIcon,
  ],
  templateUrl: './page-wizard-modifiers.component.html',
  styleUrl: './page-wizard-modifiers.component.css'
})
export class PageWizardModifiersComponent {

  typeOptions = [
    {label: "Dropdown", value: ModifierType.DROPDOWN},
    {label: "Multi select", value: ModifierType.MULTI},
  ]

  @Input({required:true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();
  @Output() back: EventEmitter<void> = new EventEmitter();

  constructor(
    private msgService: MessageService,
    private dialogService: DialogService,
  ) {
  }

  nextCallback() {
    for (const mod of this.page.modifiers) {
      if (mod.key === '' || mod.title === '') {
        const title = mod.title === '' ? mod.key : mod.title;
        this.msgService.error(`Invalid modifier ${title}`, "Ensure all modifiers have their key and title set");
        return;
      }

      for (const val of mod.values) {
        if (val.key === '' || val.value === '') {
          this.msgService.error(`Invalid modifier ${mod.title}`, "Ensure all modifier values have their key and value set")
          return;
        }
      }
    }

    this.next.emit();
  }

  addNewModifierValue(mod: Modifier) {
    if (mod.values.filter(v => v.key == '' || v.value == '').length > 0) {
      this.msgService.warning("Cannot add modifier value", "All modifiers must be filled in before adding a new one")
      return;
    }

    mod.values.push({
      key: '',
      value: ''
    });
  }

  async deleteModifierValue(mod: Modifier, key: string) {
    if (! await this.dialogService.openDialog(`Certain you want to delete ${mod.title} > ${key}`)) {
      return;
    }

    mod.values = mod.values.filter(value => value.key !== key);
  }

  async delete(toDelete: Modifier) {
    if (! await this.dialogService.openDialog(`Certain you want to delete ${toDelete.title}`)) {
      return;
    }

    this.page.modifiers = this.page.modifiers.filter(modifier => modifier !== toDelete);
  }

  addNewModifier(): void {
    this.page.modifiers = [{
      key: '',
      id: 0,
      title: '',
      type: ModifierType.DROPDOWN,
      values: []
    }, ...this.page.modifiers];
  }
}
