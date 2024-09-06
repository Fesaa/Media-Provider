import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../../_services/nav.service";
import {NgIcon} from "@ng-icons/core";
import {ServerSettingsComponent} from "./_components/server-settings/server-settings.component";
import {PagesSettingsComponent} from "./_components/pages-settings/pages-settings.component";
import {dropAnimation} from "../../_animations/drop-animation";
import {ActivatedRoute, Router} from "@angular/router";

export enum SettingsID {

  Server = "server",
  Pages = "pages",

}

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [
    NgIcon,
    ServerSettingsComponent,
    PagesSettingsComponent
  ],
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.css',
  animations: [dropAnimation]
})
export class SettingsComponent implements OnInit{
  showMobileConfig = false;

  selected: SettingsID = SettingsID.Server;
  settings: {id: SettingsID, title: string, icon: string}[] = [
    {
      id: SettingsID.Server,
      title: 'Server',
      icon: 'heroServerStack',
    },
    {
      id: SettingsID.Pages,
      title: 'Pages',
      icon: 'heroDocument',
    }
  ]

  constructor(private navService: NavService,
              private cdRef: ChangeDetectorRef,
              private activatedRoute: ActivatedRoute,
              private router: Router
  ) {

    this.activatedRoute.fragment.subscribe(fragment => {
      if (fragment) {
        if (Object.values(SettingsID).find(id => id === fragment)) {
          this.selected = fragment as SettingsID;
        }
      }
    })
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
  }

  toggleMobile() {
    this.showMobileConfig = !this.showMobileConfig;
    this.cdRef.detectChanges();
  }

  setSettings(id: SettingsID) {
    this.selected = id;
    this.router.navigate([], {fragment: id});
    this.cdRef.detectChanges();
  }

  protected readonly SettingsID = SettingsID;
}
