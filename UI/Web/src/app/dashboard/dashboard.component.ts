import {Component, DestroyRef, OnInit} from '@angular/core';
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
import {ContentPickerDialogComponent} from "./_components/content-picker-dialog/content-picker-dialog.component";
import {ToastService} from "../_services/toast.service";
import {EventType, SignalRService} from "../_services/signal-r.service";
import {ContentProgressUpdate, ContentSizeUpdate, ContentStateUpdate, DeleteContent} from "../_models/signalr";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";
import {takeUntilDestroyed} from "@angular/core/rxjs-interop";
import {SortedList} from "../shared/data-structures/sorted-list";

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
    ContentPickerDialogComponent,
    TranslocoDirective,
    TitleCasePipe
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {

  loading = true;
  dashboardItems: SortedList<InfoStat> = new SortedList((a, b) => {
    if (a.contentState == b.contentState) {
      return a.id.localeCompare(b.id)
    }

    // Bigger first
    return b.contentState - a.contentState;
  });

  displayContentPicker: { [key: string]: boolean } = {};
  protected readonly ContentState = ContentState;

  constructor(private navService: NavService,
              private contentService: ContentService,
              private toastService: ToastService,
              private contentTitle: ContentTitlePipe,
              private dialogService: DialogService,
              private signalR: SignalRService,
              private destroyRef: DestroyRef,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnInit(): void {
    this.contentService.infoStats().subscribe(info => {
      this.loading = false;
      this.dashboardItems.set(info.running || []);
    })

    this.signalR.events$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(event => {
      switch (event.type) {
        case EventType.ContentStateUpdate:
          this.updateState(event.data as ContentStateUpdate);
          break;
        case EventType.ContentSizeUpdate:
          this.updateSize(event.data as ContentSizeUpdate);
          break;
        case EventType.DeleteContent:
          this.dashboardItems.removeFunc(item => item.id !== (event.data as DeleteContent).contentId);
          break;
        case EventType.ContentProgressUpdate:
          this.updateProgress(event.data as ContentProgressUpdate);
          break;
        case EventType.AddContent:
          this.addContent(event.data as InfoStat);
          break;
        case EventType.ContentInfoUpdate:
          this.updateInfo(event.data as InfoStat);
          break;
      }
    })
  }

  private updateInfo(info: InfoStat) {
    this.dashboardItems.setFunc(i => {
      if (i.id !== info.id) {
        return i;
      }
      return info;
    });
  }

  private addContent(event: InfoStat) {
    if (this.dashboardItems.getFunc(is => is.id == event.id)) {
      return;
    }

    this.dashboardItems.add(event);
  }

  private updateSize(event: ContentSizeUpdate) {
    const content = this.dashboardItems.getFunc(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.size = event.size;
  }

  private updateProgress(event: ContentProgressUpdate) {
    const content = this.dashboardItems.getFunc(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.progress = event.progress;
    content.estimated = event.estimated;
    content.speed = event.speed;
    content.speed_type = event.speed_type;
  }

  private updateState(event: ContentStateUpdate) {
    const content = this.dashboardItems.getFunc(is => is.id == event.contentId)
    if (!content) {
      return;
    }
    content.contentState = event.contentState;
  }

  async stop(info: InfoStat) {
    if (!await this.dialogService.openDialog("dashboard.confirm-stop", {name: this.contentTitle.transform(info.name)})) {
      return;
    }

    const req: StopRequest = {
      provider: info.provider,
      delete: true,
      id: info.id,
    }

    this.contentService.stop(req).subscribe({
      next: () => {
        this.toastService.successLoco("dashboard.toasts.stopped-success", {}, {title: this.contentTitle.transform(info.name)});
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  async browse(info: InfoStat) {
    await this.dialogService.openDirBrowser(info.download_dir, {showFiles: true, width: '40rem',})
  }

  markReady(info: InfoStat) {
    this.contentService.startDownload(info.provider, info.id).subscribe({
      next: () => {
        this.toastService.successLoco("dashboard.toasts.mark-ready.success")
      },
      error: (err) => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  pickContent(info: InfoStat) {
    this.displayContentPicker = {} // Close others
    this.displayContentPicker[info.id] = true;
  }

  getSeverity(info: InfoStat): "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | undefined {
    switch (info.contentState) {
      case ContentState.Cleanup:
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
