import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page, Provider, providerNames, providerValues} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule} from "@angular/forms";
import {TitleCasePipe} from "@angular/common";

@Component({
  selector: 'app-page-wizard-general',
  imports: [
    Card,
    FloatLabel,
    InputText,
    FormsModule,
    TitleCasePipe
  ],
  templateUrl: './page-wizard-general.component.html',
  styleUrl: './page-wizard-general.component.css'
})
export class PageWizardGeneralComponent {

  @Input({required:true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();

  hasProvider(provider: Provider) {
    return this.page.providers.includes(provider);
  }

  onProviderCheckboxChange(provider: number) {
    if (this.hasProvider(provider)) {
      this.page.providers = this.page.providers.filter(p => p !== provider);
    } else {
      this.page.providers = [...this.page.providers, provider];
    }
  }

  protected readonly providerValues = providerValues;
  protected readonly providerNames = providerNames;
}
