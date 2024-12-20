import {Component, Input} from '@angular/core';
import {InfoStat, QueueStat} from "../../../_models/stats";
import {DownloadService} from "../../../_services/download.service";
import {ToastrService} from "ngx-toastr";
import {ContentTitlePipe} from "../../../_pipes/content-title.pipe";
import {StopRequest} from "../../../_models/search";
import {NgIcon} from "@ng-icons/core";
import {SpeedPipe} from "../../../_pipes/speed.pipe";
import {SpeedTypePipe} from "../../../_pipes/speed-type.pipe";

@Component({
    selector: 'app-queued-info',
    imports: [
        ContentTitlePipe,
        NgIcon,
        SpeedPipe,
        SpeedTypePipe
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
        this.toastR.error(`Failed to stop download: ${err.message}`, "Error")
      }
    })
  }
}
