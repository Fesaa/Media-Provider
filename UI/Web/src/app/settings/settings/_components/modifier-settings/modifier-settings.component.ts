import {ChangeDetectorRef, Component, HostListener, Input, OnInit} from '@angular/core';
import {KeyValuePipe} from "@angular/common";
import {NgIcon} from "@ng-icons/core";
import {Modifier} from "../../../../_models/page";
import {FormBuilder, FormGroup, Validators} from "@angular/forms";
import {DialogService} from "../../../../_services/dialog.service";
import {ToastrService} from "ngx-toastr";

@Component({
  selector: 'app-modifier-settings',
  standalone: true,
  imports: [
    KeyValuePipe,
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

  getModifiers() {
    const modifiers: {[key: string]: Modifier} = {};
    if (this.pageForm === undefined) {
      return modifiers;
    }

    const form = this.pageForm.controls['modifiers'] as FormGroup;
    for (const [key, value] of Object.entries(form.controls)) {
      modifiers[key] = value.value;
    }

    return modifiers;
  }

  addModifier() {
    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    modifierGroup.addControl('modifier', this.fb.group({
      title: this.fb.control('', [Validators.required]),
      type: this.fb.control('string', [Validators.required]),
      values: this.fb.control({}),
    }));
  }

  updateModifierTitle(key: string, e: Event) {
    const title = (e.target as HTMLInputElement).value;

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    modifier.controls['title'].patchValue(title);
  }

  updateModifierKey(key: string, e: Event) {
    const newKey = (e.target as HTMLInputElement).value;

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    modifierGroup.removeControl(key);
    modifierGroup.addControl(newKey, modifier);
  }

  async removeModifier(key: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${key}?`)) {
      return;
    }

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    modifierGroup.removeControl(key);
    this.toastR.warning(`Removed modifier ${key}`, 'Success');
  }

  async removeModifierValue(key: string, valueKey: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${valueKey}?`)) {
      return;
    }

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    const values = modifier.controls['values'];
    delete values.value[valueKey];
    values.patchValue(values.value);
    this.toastR.warning(`Removed value ${valueKey}`, 'Success');
  }

  addModifierValue(key: string) {
    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    const values = modifier.controls['values'];
    values.value['key'] = 'value';
  }

}
