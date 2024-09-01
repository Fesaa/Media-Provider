import { Component } from '@angular/core';
import {NavService} from "../_services/nav.service";

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [],
  templateUrl: './page.component.html',
  styleUrl: './page.component.css'
})
export class PageComponent {

  constructor(private navService: NavService) {
    this.navService.setNavVisibility(true);
  }

}
