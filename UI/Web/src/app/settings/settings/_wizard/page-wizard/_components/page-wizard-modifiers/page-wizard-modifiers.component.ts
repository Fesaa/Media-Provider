import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Modifier, ModifierType, Page} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {Fieldset} from "primeng/fieldset";
import {FormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {NgForOf, TitleCasePipe} from "@angular/common";
import {DialogService} from "../../../../../../_services/dialog.service";
import {ToastService} from "../../../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {CdkDrag, CdkDragDrop, CdkDragHandle, CdkDropList, moveItemInArray} from "@angular/cdk/drag-drop";
import {ModifierEditModalComponent} from "./_components/modifier-edit-modal/modifier-edit-modal.component";

@Component({
  selector: 'app-page-wizard-modifiers',
  imports: [
    Card,
    Fieldset,
    FormsModule,
    Button,
    NgForOf,
    TranslocoDirective,
    TitleCasePipe,
    CdkDropList,
    CdkDrag,
    CdkDragHandle,
    ModifierEditModalComponent,
  ],
  templateUrl: './page-wizard-modifiers.component.html',
  styleUrl: './page-wizard-modifiers.component.scss'
})
export class PageWizardModifiersComponent {

  @Input({required: true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();
  @Output() back: EventEmitter<void> = new EventEmitter();

  modifierVisible: { [key: string]: boolean } = {};
  editModifier: Modifier | null = null;

  constructor(
    private toastService: ToastService,
    private dialogService: DialogService,
  ) {
  }

  show(mod: Modifier) {
    this.modifierVisible = {}
    this.modifierVisible[mod.ID] = true;
  }

  nextCallback() {
    for (const mod of this.page.modifiers) {
      if (mod.key === '' || mod.title === '') {
        const title = mod.title === '' ? mod.key : mod.title;
        this.toastService.errorLoco("settings.pages.toasts.invalid-modifier", {title: mod.title});
        return;
      }

      for (const val of mod.values) {
        if (val.key === '' || val.value === '') {
          this.toastService.errorLoco("settings.pages.toasts.invalid-modifier-values", {title: mod.title});
          return;
        }
      }
    }

    this.next.emit();
  }

  async delete(toDelete: Modifier) {
    if (!await this.dialogService.openDialog("settings.pages.wizard.confirm-delete-modifier", {name: toDelete.title})) {
      return;
    }

    this.page.modifiers = this.page.modifiers.filter(modifier => modifier !== toDelete);
  }

  addNewModifier(): void {
    if (this.editModifier !== null) {
      return;
    }

    this.editModifier = {
      key: '',
      ID: 0,
      title: '',
      type: ModifierType.DROPDOWN,
      values: [
        {key: '', value: '', default: false},
      ]
    }
    this.show(this.editModifier);
  }

  closeEdit() {
    if (this.editModifier === null) {
      return;
    }

    if (this.editModifier.title.length === 0
      || this.editModifier.key.length === 0
      || this.editModifier.values.length === 0
    ) {
      this.toastService.warningLoco("settings.pages.wizard.modifiers.modifier-needs");
    } else {
      this.page.modifiers.push(this.editModifier);
    }

    this.editModifier = null;
  }

  drop(event: CdkDragDrop<any, any>) {
    moveItemInArray(this.page.modifiers, event.previousIndex, event.currentIndex)
  }
}
