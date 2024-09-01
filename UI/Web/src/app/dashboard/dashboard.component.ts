import { Component } from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {AsyncPipe} from "@angular/common";
import {SuggestionDashboardComponent} from "./_components/suggestion-dashboard/suggestion-dashboard.component";
import {DownloadService} from "../_services/download.service";
import {InfoStat, QueueStat} from "../_models/stats";

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
export class DashboardComponent {

  running: InfoStat[] | [] = [];
  queued: QueueStat[] | [] = [];

  constructor(private navService: NavService,
              protected pageService: PageService,
              private downloadService: DownloadService,
  ) {
    this.navService.setNavVisibility(true);

    this.downloadService.running$.subscribe(running =>{
      if (running) {
        this.running = running;
      }
    });

    this.downloadService.queued$.subscribe(queued =>{
      if (queued) {
        this.queued = queued;
      }
    });
  }

}
