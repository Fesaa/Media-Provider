import {Component, EventEmitter, Input, Output} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {DownloadMetadata, DownloadMetadataDefinition, DownloadMetadataFormType, Page} from "../../../_models/page";
import {DownloadRequest} from "../../../_models/search";
import {DownloadService} from "../../../_services/download.service";
import {ToastrService} from "ngx-toastr";
import {FormGroup, FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {FormInputComponent} from "../../../shared/form/form-input/form-input.component";
import {NgIcon} from "@ng-icons/core";
import {DialogService} from "../../../_services/dialog.service";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputText} from "primeng/inputtext";
import {InputIcon} from "primeng/inputicon";
import {Select} from "primeng/select";
import {ToggleSwitch} from "primeng/toggleswitch";

@Component({
  selector: 'app-download-dialog',
  imports: [
    Button,
    ReactiveFormsModule,
    FloatLabel,
    IconField,
    InputText,
    InputIcon,
    FormsModule,
    Select,
    ToggleSwitch
  ],
  templateUrl: './download-dialog.component.html',
  styleUrl: './download-dialog.component.css'
})
export class DownloadDialogComponent {

  @Input({required: true}) visible!: boolean;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  @Input({required: true}) downloadDir!: string;
  @Input({required: true}) searchResult!: SearchInfo;
  @Input({required: true}) metadata!: DownloadMetadata | undefined;

  metadataChoices: { [key: string]: string[] } = {};

  constructor(
    private downloadService: DownloadService,
    private toastR: ToastrService,
    private dialogService: DialogService,
  ) {
  }

  changeChoice(meta: DownloadMetadataDefinition, value: string | boolean) {
    this.metadataChoices[meta.key] = [String(value)];
  }

  download() {
    const req: DownloadRequest = {
      provider: this.searchResult.Provider,
      title: this.searchResult.Name,
      id: this.searchResult.InfoHash,
      dir: this.downloadDir,
      downloadMetadata: this.metadataChoices,
    }

    console.log(req)

    /*this.downloadService.download(req).subscribe({
      next: () => {
        this.toastR.success(`Downloaded started for ${this.searchResult.Name}`, "Success")
      },
      error: (err) => {
        this.toastR.error(`Download failed ${err.error.message}`, "Error")
      }
    }).add(() => {
      this.close()
    })*/
  }

  close() {
    this.visibleChange.emit(false);
  }

  protected readonly DownloadMetadataFormType = DownloadMetadataFormType;
}
