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
import {EventType, SignalRService} from "../_services/signal-r.service";
import {ContentProgressUpdate, ContentSizeUpdate, ContentStateUpdate, DeleteContent} from "../_models/signalr";

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
export class DashboardComponent implements OnInit {

  loading = true;
  info: InfoStat[] = [];
  infoString: string = '';

  displayContentPicker: { [key: string]: boolean } = {};
  protected readonly ContentState = ContentState;

  constructor(private navService: NavService,
              private contentService: ContentService,
              private cdRef: ChangeDetectorRef,
              private msgService: MessageService,
              private contentTitle: ContentTitlePipe,
              private dialogService: DialogService,
              private signalR: SignalRService,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnInit(): void {
    this.contentService.infoStats().subscribe(info => {
      this.loading = false;
      this.info = info.running || [];
      this.sortInfo()
    })

    this.signalR.events$.subscribe(event => {
      switch (event.type) {
        case EventType.ContentStateUpdate:
          this.updateState(event.data as ContentStateUpdate);
          break;
        case EventType.ContentSizeUpdate:
          this.updateSize(event.data as ContentSizeUpdate);
          break;
        case EventType.DeleteContent:
          this.info = this.info.filter(item => item.id !== (event.data as DeleteContent).contentId);
          break;
        case EventType.ContentProgressUpdate:
          this.updateProgress(event.data as ContentProgressUpdate);
          break;
        case EventType.AddContent:
          this.addContent(event.data as InfoStat);
          break;
      }
    })
  }

  private addContent(event: InfoStat) {
    if (this.info.find(is => is.id == event.id)) {
      return;
    }

    this.info.push(event);
    this.sortInfo()
  }

  private updateSize(event: ContentSizeUpdate) {
    const content = this.info.find(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.size = event.size;
  }

  private updateProgress(event: ContentProgressUpdate) {
    const content = this.info.find(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.progress = event.progress;
    content.estimated = event.estimated;
    content.speed = event.speed;
    content.speed_type = event.speed_type;
  }

  private updateState(event: ContentStateUpdate) {
    const content = this.info.find(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.contentState = event.contentState;
  }

  private sortInfo() {
    this.info = this.info.sort((a, b) => {
      if (a.contentState == b.contentState) {
        return a.id.localeCompare(b.id)
      }

      return a.contentState - b.contentState;
    });
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
