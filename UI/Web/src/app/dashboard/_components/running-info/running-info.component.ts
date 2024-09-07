import {Component, Input} from '@angular/core';
import {InfoStat} from "../../../_models/stats";
import {DownloadService} from "../../../_services/download.service";
import {StopRequest} from "../../../_models/search";
import {ToastrService} from "ngx-toastr";
import {ContentTitlePipe} from "../../../_pipes/content-title.pipe";
import {NgIcon} from "@ng-icons/core";
import {SpeedPipe} from "../../../_pipes/speed.pipe";
import {SpeedTypePipe} from "../../../_pipes/speed-type.pipe";
import {DirectoryBrowserComponent} from "../../../directory-browser/directory-browser.component";
import {dropAnimation} from "../../../_animations/drop-animation";

@Component({
  selector: 'app-running-info',
  standalone: true,
  imports: [
    ContentTitlePipe,
    NgIcon,
    SpeedPipe,
    SpeedTypePipe,
    DirectoryBrowserComponent
  ],
  templateUrl: './running-info.component.html',
  styleUrl: './running-info.component.css',
  animations: [dropAnimation]
})
export class RunningInfoComponent {

  @Input({required: true}) info!: InfoStat;

  showDirBrowser = false;

  constructor(private downloadService: DownloadService,
              private toastR: ToastrService,
              private contentTitle: ContentTitlePipe
  ) {
  }

  toggleDirBrowser() {
    this.showDirBrowser = !this.showDirBrowser;
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
        this.toastR.error(`Failed to stop download: ${err.error.error}`, "Error")
      }
    })
  }

}
