import {ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {SuggestionDashboardComponent} from "./_components/suggestion-dashboard/suggestion-dashboard.component";
import {DownloadService} from "../_services/download.service";
import {ContentStatus, ContentStatusWeight, InfoStat, QueueStat} from "../_models/stats";
import {TableModule} from "primeng/table";
import {Tag} from 'primeng/tag';
import {ContentTitlePipe} from "../_pipes/content-title.pipe";
import {Button} from "primeng/button";
import {Tooltip} from "primeng/tooltip";
import {SpeedPipe} from "../_pipes/speed.pipe";
import {SpeedTypePipe} from "../_pipes/speed-type.pipe";
import {TimePipe} from "../_pipes/time.pipe";
import {StopRequest} from "../_models/search";
import {ToastrService} from "ngx-toastr";
import {DialogService} from "../_services/dialog.service";

@Component({
    selector: 'app-dashboard',
  imports: [
    SuggestionDashboardComponent,
    TableModule,
    ContentTitlePipe,
    Tag,
    Button,
    Tooltip,
    SpeedPipe,
    SpeedTypePipe,
    TimePipe
  ],
    templateUrl: './dashboard.component.html',
    styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit,OnDestroy {

  loading = true;
  info: InfoStat[] | [] = [];

  constructor(private navService: NavService,
              private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
              private toastR: ToastrService,
              private contentTitle: ContentTitlePipe,
              private dialogService: DialogService,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnDestroy(): void {
    this.downloadService.loadStats(false);
  }

  ngOnInit(): void {
    this.downloadService.loadStats();

    this.downloadService.stats$.subscribe(stats => {
      this.loading = false;
      this.info = (stats.running || []).sort((a, b) => {
        if (a.contentStatus == b.contentStatus) {
          return a.id.localeCompare(b.id)
        }

        return ContentStatusWeight(a.contentStatus) - ContentStatusWeight(a.contentStatus);
      });
    })
  }

  async stop(info: InfoStat) {
    if (! await this.dialogService.openDialog(`Are you sure you want to stop ${this.contentTitle.transform(info.name)}`)) {
      return;
    }

    const req: StopRequest = {
      provider: info.provider,
      delete: true,
      id: info.id,
    }

    this.downloadService.stop(req).subscribe({
      next: () => {
        this.toastR.success(`Download stopped ${this.contentTitle.transform(info.name)}`, "Success")
      },
      error: (err) => {
        this.toastR.error(`Failed to stop download: ${err.error.message}`, "Error")
      }
    })
  }

  browse(info: InfoStat) {
    this.dialogService.openDirBrowser(info.download_dir, {showFiles: true, width: '40rem',})
  }

  getSeverity(info: InfoStat): "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | undefined {
    switch (info.contentStatus) {
      case ContentStatus.Downloading:
        return "success";
      case ContentStatus.Waiting:
        return "info";
      case ContentStatus.Queued:
      case ContentStatus.Loading:
      default:
        return "secondary"
    }
  }

}
