import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Button} from "primeng/button";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {InputText} from "primeng/inputtext";
import {NgForOf} from "@angular/common";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Select} from "primeng/select";
import {Modifier, ModifierType, ModifierValue} from "../../../../../../../../_models/page";
import {TranslocoDirective} from "@jsverse/transloco";
import {Tooltip} from "primeng/tooltip";
import {ToastService} from "../../../../../../../../_services/toast.service";
import {DialogService} from "../../../../../../../../_services/dialog.service";
import {Checkbox} from "primeng/checkbox";

@Component({
  selector: 'app-modifier-edit-modal',
  imports: [
    Button,
    Dialog,
    FloatLabel,
    IconField,
    InputIcon,
    InputText,
    NgForOf,
    ReactiveFormsModule,
    Select,
    FormsModule,
    TranslocoDirective,
    Tooltip,
    Checkbox
  ],
  templateUrl: './modifier-edit-modal.component.html',
  styleUrl: './modifier-edit-modal.component.scss'
})
export class ModifierEditModalComponent {

  typeOptions = [
    {label: "Dropdown", value: ModifierType.DROPDOWN},
    {label: "Multi select", value: ModifierType.MULTI},
  ]

  @Input({required: true}) mod!: Modifier;
  @Input({required: true}) modifierVisible!: { [key: string]: boolean };
  @Output() onClose: EventEmitter<void> = new EventEmitter<void>();

  constructor(
    private dialogService: DialogService,
    private toastService: ToastService,
  ) {
  }

  addNewModifierValue(mod: Modifier) {
    if (mod.values.filter(v => v.key == '' || v.value == '').length > 0) {
      this.toastService.warningLoco("settings.pages.toasts.already-adding-new");
      return;
    }

    mod.values.push({
      key: '',
      value: '',
      default: false,
    });
  }

  async deleteModifierValue(mod: Modifier, key: string) {
    if (!await this.dialogService.openDialog("settings.pages.wizard.confirm-delete-modifier-value", {name: mod.title, key: key})) {
      return;
    }

    mod.values = mod.values.filter(value => value.key !== key);
  }

  toggleOthers(mod: Modifier, val: ModifierValue) {
    if (mod.type !== ModifierType.DROPDOWN) return;

    for (const key in mod.values) {
      if (mod.values[key].key !== val.key) {
        mod.values[key].default = false;
      }
    }
  }

}
