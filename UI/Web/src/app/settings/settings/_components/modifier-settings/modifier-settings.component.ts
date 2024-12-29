import {ChangeDetectorRef, Component, HostListener, Input, OnInit} from '@angular/core';
import {NgIcon} from "@ng-icons/core";
import {Modifier, ModifierType} from "../../../../_models/page";
import {FormBuilder, FormControl, FormGroup} from "@angular/forms";
import {DialogService} from "../../../../_services/dialog.service";
import {ToastrService} from "ngx-toastr";

@Component({
    selector: 'app-modifier-settings',
    imports: [
        NgIcon
    ],
    templateUrl: './modifier-settings.component.html',
    styleUrl: './modifier-settings.component.css'
})
export class ModifierSettingsComponent implements OnInit {

  @Input({required: true}) pageForm!: FormGroup;

  showModifiers = false;
  isMobile = false;

  constructor(private cdRef: ChangeDetectorRef,
              private fb: FormBuilder,
              private dialogService: DialogService,
              private toastR: ToastrService,

  ) {
  }

  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  ngOnInit(): void {
    this.isMobile = window.innerWidth < 768;
  }

  toggleModifiers() {
    this.showModifiers = !this.showModifiers;
    this.cdRef.detectChanges();
  }

  private get controlGroup(): FormControl<Modifier[]> {
    return this.pageForm!.controls['modifiers'] as FormControl<Modifier[]>;
  }

  getModifiers() {
    const modifiers: Modifier[] = [];
    if (this.pageForm === undefined) {
      return modifiers;
    }

    return this.controlGroup.value;
  }

  addModifier() {
    const mod = this.controlGroup.value.find(m => m.key === "");
    if (mod) {
      this.toastR.warning("Cannot add a new modifier, while one with an empty key exists", "error");
      return;
    }

    this.controlGroup.value.push({
      id: -1,
      key: '',
      title: '',
      values: [],
      type: ModifierType.DROPDOWN
    })
  }

  updateModifierTitle(key: string, e: Event) {
    this.controlGroup.value.find(m => m.key === key)!.title = (e.target as HTMLInputElement).value;
  }

  updateModifierKey(key: string, e: Event) {
    const newKey = (e.target as HTMLInputElement).value;
    const conflict = this.controlGroup.value.find(m => m.key === newKey);
    if (conflict) {
      this.toastR.warning("Key was not saved, other modifier already uses it", "warning");
      return
    }

    this.controlGroup.value.find(m => m.key === key)!.key = newKey;
  }

  async removeModifier(key: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${key}?`)) {
      return;
    }

    this.controlGroup.setValue(this.controlGroup.value.filter(m => m.key !== key));
    this.toastR.warning(`Removed modifier ${key}`, 'Success');
  }

  updateModifierValueKey(key: string, valueKey: string, e: Event) {
    const modifier = this.controlGroup.value.find(m => m.key === key)!;
    const newKey = (e.target as HTMLInputElement).value;
    const conflict = this.controlGroup.value.find(m => m.key === newKey);
    if (conflict) {
      this.toastR.warning("Key was not saved, other modifier already uses it", "warning");
      return;
    }

    const mv = modifier.values.find(m => m.key === valueKey);
    if (mv) {
      modifier.values = modifier.values.filter(m => m.key !== valueKey);
      modifier.values.push({key: (e.target as HTMLInputElement).value, value: mv.value})
    }
  }

  updateModifierValueValue(key: string, valueKey: string, e: Event) {
    const modifier = this.controlGroup.value.find(m => m.key === key)!;
    const mv = modifier.values.find(mv => mv.key === valueKey)
    if (mv) {
      mv.value = (e.target as HTMLInputElement).value
    }
  }

  async removeModifierValue(key: string, valueKey: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${valueKey}?`)) {
      return;
    }

    this.controlGroup.setValue(this.controlGroup.value.filter(m => m.key !== valueKey));
    this.toastR.warning(`Removed value ${valueKey}`, 'Success');
  }

  addModifierValue(key: string) {
    const modifier = this.controlGroup.value.find(m => m.key === key)!;
    const val = new Map(Object.entries(modifier.values)).get('');
    if (val !== undefined) {
      this.toastR.warning(`Cannot add a new value to modifier, while one with an empty key exists`, 'Error');
      return;
    }

    modifier.values.push({key: '', value: ''});
    this.cdRef.detectChanges();
  }

}
