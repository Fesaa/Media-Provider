import {Component, Input} from '@angular/core';
import {Page, Provider, providerNames, providerValues} from "../../../../_models/page";
import {TitleCasePipe} from "@angular/common";
import {FormGroup} from "@angular/forms";

@Component({
  selector: 'app-provider-settings',
  standalone: true,
  imports: [
    TitleCasePipe
  ],
  templateUrl: './provider-settings.component.html',
  styleUrl: './provider-settings.component.css'
})
export class ProviderSettingsComponent {

  @Input({required: true}) pageForm!: FormGroup;
  @Input({required: true}) page!: Page;

  hasProvider(provider: Provider) {
    return this.page.providers.includes(provider);
  }

  onProviderCheckboxChange(provider: number) {
    const formArray = this.pageForm.controls['providers'];
    if (formArray.value.includes(provider)) {
      formArray.patchValue(formArray.value.filter((v: number) => v !== provider));
    } else {
      formArray.patchValue([...formArray.value, provider]);
    }
  }


  protected readonly providerValues = providerValues;
  protected readonly providerNames = providerNames;
}
