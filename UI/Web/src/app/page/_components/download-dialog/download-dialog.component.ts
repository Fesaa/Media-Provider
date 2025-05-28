import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {DownloadMetadata, DownloadMetadataDefinition, DownloadMetadataFormType} from "../../../_models/page";
import {DownloadRequest, DownloadRequestMetadata} from "../../../_models/search";
import {ContentService} from "../../../_services/content.service";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {Select} from "primeng/select";
import {ToggleSwitch} from "primeng/toggleswitch";
import {MultiSelect} from "primeng/multiselect";
import {ToastService} from "../../../_services/toast.service";
import {Tooltip} from "primeng/tooltip";
import {TranslocoDirective} from "@jsverse/transloco";
import {NgTemplateOutlet, TitleCasePipe} from "@angular/common";
import {Tab, TabList, TabPanel, TabPanels, Tabs} from "primeng/tabs";
import {DirectorySelectorComponent} from "../../../shared/_component/directory-selector/directory-selector.component";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";

enum TabId {
  General = "general",
  Advanced = "advanced",
}

@Component({
  selector: 'app-download-dialog',
  imports: [
    Button,
    ReactiveFormsModule,
    FloatLabel,
    InputText,
    FormsModule,
    Select,
    ToggleSwitch,
    MultiSelect,
    Tooltip,
    TranslocoDirective,
    TitleCasePipe,
    Tabs,
    TabList,
    Tab,
    TabPanel,
    TabPanels,
    DirectorySelectorComponent,
    IconField,
    InputIcon,
    NgTemplateOutlet
  ],
  templateUrl: './download-dialog.component.html',
  styleUrl: './download-dialog.component.css'
})
export class DownloadDialogComponent implements OnInit {

  @Input({required: true}) visible!: boolean;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  @Input({required: true}) downloadDir!: string;
  @Input({required: true}) searchResult!: SearchInfo;
  @Input({required: true}) metadata!: DownloadMetadata | undefined;
  showPicker: boolean = false;

  requestMetadata: DownloadRequestMetadata = {
    extra: {},
    startImmediately: true,
  }
  protected readonly DownloadMetadataFormType = DownloadMetadataFormType;
  protected readonly Boolean = Boolean;

  constructor(
    private downloadService: ContentService,
    private toastService: ToastService,
  ) {
  }

  ngOnInit(): void {
    if (!this.metadata || !this.metadata.definitions) {
      return;
    }

    for (const met of this.metadata.definitions) {
      if (met.defaultOption !== "") {
        this.requestMetadata.extra[met.key] = [met.defaultOption]
      }
    }
  }

  advanced() {
    if (!this.metadata || !this.metadata.definitions) return [];
    return this.metadata.definitions.filter(definition => definition.advanced);
  }

  simple() {
    if (!this.metadata || !this.metadata.definitions) return [];
    return this.metadata.definitions.filter(definition => !definition.advanced)
  }

  changeChoice(meta: DownloadMetadataDefinition, value: string | boolean | string[]) {
    if (value instanceof Array) {
      this.requestMetadata.extra[meta.key] = value;
    } else {
      this.requestMetadata.extra[meta.key] = [String(value)];
    }
  }

  download() {
    const req: DownloadRequest = {
      provider: this.searchResult.Provider,
      title: this.searchResult.Name,
      id: this.searchResult.InfoHash,
      dir: this.downloadDir,
      downloadMetadata: this.requestMetadata,
    }

    this.downloadService.download(req).subscribe({
      next: () => {
        this.toastService.successLoco("page.download-dialog.toasts.download-success", {}, {name: this.searchResult.Name});
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    }).add(() => {
      this.close()
    })
  }

  close() {
    this.visibleChange.emit(false);
  }

  protected readonly TabId = TabId;
}
