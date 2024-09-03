import {Component, OnInit} from '@angular/core';
import { Config } from '../../_models/config';
import {NavService} from "../../_services/nav.service";
import {ConfigService} from "../../_services/config.service";

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [],
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.css'
})
export class SettingsComponent implements OnInit{

  syncID: number | undefined;

  config: Config | undefined;

  constructor(private navService: NavService, private configService: ConfigService) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
    this.configService.getConfig().subscribe(config => {
      this.config = config;
      this.syncID = config.sync_id;
    })
  }

}
