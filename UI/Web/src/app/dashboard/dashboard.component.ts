import {ChangeDetectorRef, Component, OnDestroy, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {AsyncPipe} from "@angular/common";
import {SuggestionDashboardComponent} from "./_components/suggestion-dashboard/suggestion-dashboard.component";
import {DownloadService} from "../_services/download.service";
import {InfoStat, QueueStat} from "../_models/stats";
import {combineLatest} from "rxjs";
import {RunningInfoComponent} from "./_components/running-info/running-info.component";
import {QueuedInfoComponent} from "./_components/queued-info/queued-info.component";

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    AsyncPipe,
    SuggestionDashboardComponent,
    RunningInfoComponent,
    QueuedInfoComponent
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit,OnDestroy {

  loading = true;
  running: InfoStat[] | [] = [];
  queued: QueueStat[] | [] = [];

  constructor(private navService: NavService,
              private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnDestroy(): void {
    this.downloadService.loadStats(false);
  }

  ngOnInit(): void {
    this.downloadService.loadStats();
    this.downloadService.running$.subscribe(running => {
      if (running) {
        this.running = running.sort((a, b) => a.id.localeCompare(b.id));
      }
      this.cdRef.detectChanges();
    });

    this.downloadService.queued$.subscribe(queued => {
      if (queued) {
        this.queued = queued.sort((a, b) => a.id.localeCompare(b.id));
      }
      this.cdRef.detectChanges();
    });
  }

}
