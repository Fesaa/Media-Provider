import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Page, Provider, providerNames, providerValues} from "../../../../../../_models/page";
import {Card} from "primeng/card";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule} from "@angular/forms";
import {TitleCasePipe} from "@angular/common";
import {Fieldset} from "primeng/fieldset";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {ToastService} from "../../../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-page-wizard-general',
  imports: [
    Card,
    FloatLabel,
    InputText,
    FormsModule,
    TitleCasePipe,
    Fieldset,
    IconField,
    InputIcon,
    TranslocoDirective
  ],
  templateUrl: './page-wizard-general.component.html',
  styleUrl: './page-wizard-general.component.css'
})
export class PageWizardGeneralComponent {

  @Input({required: true}) page!: Page;
  @Output() next: EventEmitter<void> = new EventEmitter();
  protected readonly providerValues = providerValues;
  protected readonly providerNames = providerNames;

  constructor(private toastService: ToastService,) {
  }

  nextCallback() {
    if (this.page.title === '') {
      this.toastService.errorLoco("settings.pages.toasts.name-required");
      return;
    }

    if (this.page.providers.length == 0) {
      this.toastService.errorLoco("settings.pages.toasts.provider-required");
      return;
    }

    this.next.emit();
  }

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
}
