import {Component, Input} from '@angular/core';
import {QueueStat} from "../../../_models/stats";
import {DownloadService} from "../../../_services/download.service";
import {ToastrService} from "ngx-toastr";
import {ContentTitlePipe} from "../../../_pipes/content-title.pipe";
import {StopRequest} from "../../../_models/search";
import {NgIcon} from "@ng-icons/core";

@Component({
    selector: 'app-queued-info',
    imports: [
        ContentTitlePipe,
        NgIcon,
    ],
    templateUrl: './queued-info.component.html',
    styleUrl: './queued-info.component.css'
})
export class QueuedInfoComponent {
  @Input({required: true}) info!: QueueStat;

  constructor(private downloadService: DownloadService,
              private toastR: ToastrService,
              private contentTitle: ContentTitlePipe
  ) {
  }


  stop() {
    const req: StopRequest = {
      provider: this.info.provider,
      delete: true,
      id: this.info.id,
    }

    this.downloadService.stop(req).subscribe({
      next: () => {
        this.toastR.success(`Download stopped ${this.contentTitle.transform(this.info.name)}`, "Success")
      },
      error: (err) => {
        this.toastR.error(`Failed to stop download: ${err.error.message}`, "Error")
      }
    })
  }
}
