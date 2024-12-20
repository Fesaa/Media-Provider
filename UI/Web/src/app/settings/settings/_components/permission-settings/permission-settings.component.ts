import {Component, Input} from '@angular/core';
import {Page, Provider, providerNames, providerValues} from "../../../../_models/page";
import {TitleCasePipe} from "@angular/common";
import {FormGroup} from "@angular/forms";
import {hasPermission, Perm, permissionNames, permissionValues, UserDto} from "../../../../_models/user";

@Component({
    selector: 'app-permissions-settings',
    imports: [
        TitleCasePipe
    ],
    templateUrl: './permission-settings.component.html',
    styleUrl: './permission-settings.component.css'
})
export class PermissionSettingsComponent {

  @Input({required: true}) pageForm!: FormGroup;
  @Input({required: true}) user!: UserDto;

  hasPermission(perm: Perm) {
    return hasPermission(this.user, perm);
  }

  onProviderCheckboxChange(perm: Perm) {
    const formArray = this.pageForm.controls['permissions'];
    if (formArray.value.includes(perm)) {
      formArray.patchValue(formArray.value.filter((v: Perm) => v !== perm));
    } else {
      formArray.patchValue([...formArray.value, perm]);
    }
  }

  protected readonly permissionValues = permissionValues;
  protected readonly permissionNames = permissionNames;
}
