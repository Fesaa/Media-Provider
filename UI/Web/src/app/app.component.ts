import {Component, ViewContainerRef} from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {AccountService} from "./_services/account.service";
import {AsyncPipe} from "@angular/common";
import {NavHeaderComponent} from "./nav-header/nav-header.component";
import {Title} from "@angular/platform-browser";
import {DialogService} from "./_services/dialog.service";

@Component({
    selector: 'app-root',
    imports: [RouterOutlet, NavHeaderComponent],
    templateUrl: './app.component.html',
    styleUrl: './app.component.css'
})
export class AppComponent {
  title = 'Media Provider';

  constructor(protected accountService: AccountService, private titleService: Title, private vcr: ViewContainerRef, private ds: DialogService) {
    this.titleService.setTitle(this.title);
    this.ds.viewContainerRef = this.vcr;
  }
}
