import {Component, Input} from '@angular/core';
import {DownloadMetadata, DownloadMetadataDefinition, DownloadMetadataFormType} from "../../../_models/page";
import {NgIf} from "@angular/common";
import {Tooltip} from "primeng/tooltip";
import {ToggleSwitch} from "primeng/toggleswitch";
import {Select} from "primeng/select";
import {MultiSelect} from "primeng/multiselect";
import {FormsModule} from "@angular/forms";
import {InputText} from "primeng/inputtext";
import {CdkFixedSizeVirtualScroll, CdkVirtualForOf, CdkVirtualScrollViewport} from "@angular/cdk/scrolling";
import { Subscription } from '../../../_models/subscription';
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-provider-metadata-options',
  imports: [
    NgIf,
    CdkFixedSizeVirtualScroll,
    CdkVirtualForOf,
    Tooltip,
    CdkFixedSizeVirtualScroll,
    CdkVirtualScrollViewport,
    ToggleSwitch,
    Select,
    MultiSelect,
    FormsModule,
    InputText,
    TranslocoDirective
  ],
  templateUrl: './provider-metadata-options.component.html',
  styleUrl: './provider-metadata-options.component.css'
})
export class ProviderMetadataOptionsComponent {

  @Input({required: true}) subscription!: Subscription;
  @Input() metadata!: DownloadMetadata | undefined;
  @Input() canHideMetadata: boolean = false;
  @Input() hideMetadata: boolean = false;

  changeChoice(meta: DownloadMetadataDefinition, value: string | boolean | string[]) {
    if (value instanceof Array) {
      this.subscription.metadata.extra[meta.key] = value;
    } else {
      this.subscription.metadata.extra[meta.key] = [String(value)];
    }
  }

  getValues(def: DownloadMetadataDefinition) {
    const values = this.subscription.metadata.extra[def.key] || [];
    if (values.length == 0) {
      return def.defaultOption;
    }

    switch (def.formType) {
      case DownloadMetadataFormType.MULTI:
        return values;
      case DownloadMetadataFormType.SWITCH:
        return values[0] == "true";
      case DownloadMetadataFormType.DROPDOWN:
      case DownloadMetadataFormType.TEXT:
        return values[0];
    }
  }


  protected readonly DownloadMetadataFormType = DownloadMetadataFormType;
  protected readonly Boolean = Boolean;
}
