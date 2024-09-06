import {ChangeDetectorRef, Component, Input} from '@angular/core';
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {NgIcon} from "@ng-icons/core";
import {FormArray, FormBuilder, FormGroup} from "@angular/forms";
import {DialogService} from "../../../../_services/dialog.service";
import {ToastrService} from "ngx-toastr";

@Component({
  selector: 'app-directory-settings',
  standalone: true,
  imports: [
    FormInputComponent,
    NgIcon
  ],
  templateUrl: './directory-settings.component.html',
  styleUrl: './directory-settings.component.css'
})
export class DirectorySettingsComponent {

  @Input({required: true}) pageForm!: FormGroup;

  constructor(private cdRef: ChangeDetectorRef,
              private fb: FormBuilder,
              private dialogService: DialogService,
              private toastR: ToastrService,

  ) {
  }

  async updateCustomDir() {
    const newDir = await this.dialogService.openDirBrowser("");
    if (newDir === undefined) {
      return;
    }
    this.pageForm?.controls['custom_root_dir'].patchValue(newDir);
  }

  getDirs() {
    return this.pageForm?.controls['dirs'].value;
  }

  updateDir(index: number, e: Event) {
    const dir = (e.target as HTMLInputElement).value;
    this.updateInArray(this.pageForm?.controls['dirs'] as FormArray, dir, index);
  }

  async getNewDir(index: number) {
    const dir = await this.dialogService.openDirBrowser("");
    if (dir === undefined) {
      return;
    }
    this.updateInArray(this.pageForm?.controls['dirs'] as FormArray, dir, index);
  }

  async removeDir(index: number) {
    const dirs = this.pageForm?.controls['dirs'] as FormArray;
    const values = dirs.value;
    if (index >= values.length) {
      this.toastR.error('Invalid index', 'Error');
      return;
    }

    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${values[index]}?`)) {
      return;
    }

    if (dirs.value.length === 1) {
      this.toastR.error('You must have at least one directory', 'Error');
      return;
    }

    dirs.patchValue(values.filter((_: any, i: number) => i !== index));
    this.toastR.warning(`Removed directory ${values[index]}`, 'Success');
  }

  private updateInArray(formArray: FormArray, value: any, index: number) {
    const values = formArray.value;

    if (index >= values.length) {
      const find = values.find((v: any) => v === value);
      if (find !== undefined) {
        this.toastR.info('Directory already added', 'Nothing happened');
        return;
      }

      values.push(value);
      formArray.patchValue(values);
      this.toastR.success(`Added directory ${value}`, 'Success');
      return;
    }

    values[index] = value;
    formArray.patchValue(values);
  }

}
