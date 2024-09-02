import { Component } from '@angular/core';
import {PageService} from "../../../_services/page.service";
import {AsyncPipe} from "@angular/common";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {dropAnimation} from "../../../_animations/drop-animation";

@Component({
  selector: 'app-suggestion-dashboard',
  standalone: true,
  imports: [
    AsyncPipe,
    RouterLink,
    NgIcon
  ],
  templateUrl: './suggestion-dashboard.component.html',
  styleUrl: './suggestion-dashboard.component.css',
  animations: [dropAnimation]
})
export class SuggestionDashboardComponent {

  constructor(protected pageService: PageService) { }

}
