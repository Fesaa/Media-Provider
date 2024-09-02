import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import {AccountService} from "./_services/account.service";
import {AsyncPipe} from "@angular/common";
import {NavHeaderComponent} from "./nav-header/nav-header.component";
import {Title} from "@angular/platform-browser";

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, AsyncPipe, NavHeaderComponent],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent {
  title = 'Media Provider';

  constructor(protected accountService: AccountService, private titleService: Title) {
    this.titleService.setTitle(this.title);
  }
}
