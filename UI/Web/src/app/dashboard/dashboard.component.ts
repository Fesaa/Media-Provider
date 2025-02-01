import {ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {SuggestionDashboardComponent} from "./_components/suggestion-dashboard/suggestion-dashboard.component";
import {ContentService} from "../_services/content.service";
import {ContentState, InfoStat} from "../_models/stats";
import {TableModule} from "primeng/table";
import {Tag} from 'primeng/tag';
import {ContentTitlePipe} from "../_pipes/content-title.pipe";
import {Button} from "primeng/button";
import {Tooltip} from "primeng/tooltip";
import {SpeedPipe} from "../_pipes/speed.pipe";
import {SpeedTypePipe} from "../_pipes/speed-type.pipe";
import {TimePipe} from "../_pipes/time.pipe";
import {StopRequest} from "../_models/search";
import {DialogService} from "../_services/dialog.service";
import {ContentStatePipe} from "../_pipes/content-state.pipe";
import {Dialog} from "primeng/dialog";
import {ContentPickerDialogComponent} from "./_components/content-picker-dialog/content-picker-dialog.component";
import {MessageService} from "../_services/message.service";

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
    TimePipe,
    ContentStatePipe,
    Dialog,
    ContentPickerDialogComponent
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit, OnDestroy {

  loading = true;
  info: InfoStat[] | [] = [];
  infoString: string = '';

  displayContentPicker: { [key: string]: boolean } = {};
  protected readonly ContentState = ContentState;

  constructor(private navService: NavService,
              private contentService: ContentService,
              private cdRef: ChangeDetectorRef,
              private msgService: MessageService,
              private contentTitle: ContentTitlePipe,
              private dialogService: DialogService,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnDestroy(): void {
    this.contentService.loadStats(false);
  }

  ngOnInit(): void {
    this.contentService.loadStats();

    this.contentService.stats$.subscribe(stats => {
      this.loading = false;

      const newInfo = (stats.running || []).sort((a, b) => {
        if (a.contentState == b.contentState) {
          return a.id.localeCompare(b.id)
        }

        return a.contentState - b.contentState;
      });

      if (this.infoString !== JSON.stringify(newInfo)) {
        this.info = newInfo;
        this.infoString = JSON.stringify(this.info)
      }
    })
  }

  async stop(info: InfoStat) {
    if (!await this.dialogService.openDialog(`Are you sure you want to stop ${this.contentTitle.transform(info.name)}`)) {
      return;
    }

    const req: StopRequest = {
      provider: info.provider,
      delete: true,
      id: info.id,
    }

    this.contentService.stop(req).subscribe({
      next: () => {
        this.msgService.success("Success", `Download stopped ${this.contentTitle.transform(info.name)}`)
      },
      error: (err) => {
        this.msgService.error("Error", `Failed to stop download: ${err.error.message}`)
      }
    })
  }

  browse(info: InfoStat) {
    this.dialogService.openDirBrowser(info.download_dir, {showFiles: true, width: '40rem',})
  }

  markReady(info: InfoStat) {
    this.contentService.startDownload(info.provider, info.id).subscribe({
      next: () => {
        this.msgService.success("Success", "Content marked as ready for download, will begin as soon as possible")
      },
      error: (err) => {
        this.msgService.error("Error", `Failed to mark as ready:\n ${err.error.message}`)
      }
    })
  }

  pickContent(info: InfoStat) {
    this.displayContentPicker[info.id] = true;
  }

  getSeverity(info: InfoStat): "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | undefined {
    switch (info.contentState) {
      case ContentState.Downloading:
        return "success";
      case ContentState.Ready:
      case ContentState.Waiting:
        return "info";
      case ContentState.Queued:
      case ContentState.Loading:
      default:
        return "secondary"
    }
  }
}
