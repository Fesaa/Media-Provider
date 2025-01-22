import { Component } from '@angular/core';
import {PageService} from "../../../../_services/page.service";
import {NavService} from "../../../../_services/nav.service";
import {Page} from "../../../../_models/page";
import {Steps} from "primeng/steps";
import {MenuItem} from "primeng/api";
import {PageWizardGeneralComponent} from "./_components/page-wizard-general/page-wizard-general.component";
import {PageWizardDirsComponent} from "./_components/page-wizard-dirs/page-wizard-dirs.component";
import {PageWizardModifiersComponent} from "./_components/page-wizard-modifiers/page-wizard-modifiers.component";

export enum PageWizardID {
  General = 'General',
  Dirs = 'Dirs',
  Modifiers = 'Modifiers',
}

@Component({
  selector: 'app-page-wizard',
  imports: [
    Steps,
    PageWizardGeneralComponent,
    PageWizardDirsComponent,
    PageWizardModifiersComponent
  ],
  templateUrl: './page-wizard.component.html',
  styleUrl: './page-wizard.component.css'
})
export class PageWizardComponent {

  page: Page = {
    ID: 0,
    title: "",
    custom_root_dir: "",
    dirs: [],
    providers: [],
    modifiers: [],
    sortValue: 1000, // TODO: GET THIS FIXED
  };

  index: number = 0;
  sections: {id: PageWizardID, label: string}[] = [
    {
      id: PageWizardID.General,
      label: "General"
    },
    {
      id: PageWizardID.Dirs,
      label: "Directories"
    },
    {
      id: PageWizardID.Modifiers,
      label: "Modifiers"
    },
  ];

  constructor(private pageService: PageService,
              private navService: NavService,
  ) {
    this.navService.setNavVisibility(true)
  }


  protected readonly PageWizardID = PageWizardID;
}
