import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {AsyncPipe} from "@angular/common";
import {SuggestionDashboardComponent} from "./_components/suggestion-dashboard/suggestion-dashboard.component";
import {DownloadService} from "../_services/download.service";
import {InfoStat, QueueStat} from "../_models/stats";
import {combineLatest} from "rxjs";

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [
    AsyncPipe,
    SuggestionDashboardComponent
  ],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.css'
})
export class DashboardComponent implements OnInit {

  loading = true;
  running: InfoStat[] | [] = [];
  queued: QueueStat[] | [] = [];

  constructor(private navService: NavService,
              protected pageService: PageService,
              private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnInit(): void {
    this.downloadService.loadStats();

    combineLatest([
      this.downloadService.running$,
      this.downloadService.queued$
    ]).subscribe(([running, queued]) => {
      if (running) {
        this.running = running;
      }
      if (queued) {
        this.queued = queued;
      }
      this.loading = false;
      this.cdRef.detectChanges();
    });
  }

}
