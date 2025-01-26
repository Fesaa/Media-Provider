import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {DownloadMetadata, DownloadMetadataDefinition, DownloadMetadataFormType} from "../../../_models/page";
import {DownloadRequest, DownloadRequestMetadata} from "../../../_models/search";
import {DownloadService} from "../../../_services/download.service";
import {ToastrService} from "ngx-toastr";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {Select} from "primeng/select";
import {ToggleSwitch} from "primeng/toggleswitch";
import {MultiSelect} from "primeng/multiselect";

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
    MultiSelect
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

  requestMetadata: DownloadRequestMetadata = {
    extra: {},
    startImmediately: true,
  }

  constructor(
    private downloadService: DownloadService,
    private toastR: ToastrService,
  ) {
  }

  ngOnInit(): void {
    if (!this.metadata) {
      return;
    }

    for (const met of this.metadata.definitions) {
      if (met.defaultOption !== "") {
        this.requestMetadata.extra[met.key] = [met.defaultOption]
      }
    }
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

    console.log(req)

    this.downloadService.download(req).subscribe({
      next: () => {
        this.toastR.success(`Downloaded started for ${this.searchResult.Name}`, "Success")
      },
      error: (err) => {
        this.toastR.error(`Download failed ${err.error.message}`, "Error")
      }
    }).add(() => {
      this.close()
    })
  }

  close() {
    this.visibleChange.emit(false);
  }

  protected readonly DownloadMetadataFormType = DownloadMetadataFormType;
  protected readonly Boolean = Boolean;
  protected readonly HTMLInputElement = HTMLInputElement;
}
