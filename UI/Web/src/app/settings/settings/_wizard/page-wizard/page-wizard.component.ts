import { Component } from '@angular/core';
import {PageService} from "../../../../_services/page.service";
import {NavService} from "../../../../_services/nav.service";
import {Page} from "../../../../_models/page";
import {Steps} from "primeng/steps";
import {MenuItem} from "primeng/api";
import {PageWizardGeneralComponent} from "./_components/page-wizard-general/page-wizard-general.component";
import {PageWizardDirsComponent} from "./_components/page-wizard-dirs/page-wizard-dirs.component";
import {PageWizardModifiersComponent} from "./_components/page-wizard-modifiers/page-wizard-modifiers.component";
import {ActivatedRoute, Router} from "@angular/router";
import {Skeleton} from "primeng/skeleton";
import {Card} from "primeng/card";
import {ToastrService} from "ngx-toastr";

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
    PageWizardModifiersComponent,
    Skeleton,
    Card
  ],
  templateUrl: './page-wizard.component.html',
  styleUrl: './page-wizard.component.css'
})
export class PageWizardComponent {

  private readonly defaultPage: Page = {
    ID: 0,
    title: "",
    custom_root_dir: "",
    dirs: [],
    providers: [],
    modifiers: [],
    sortValue: 1000, // TODO: GET THIS FIXED
  };

  page: Page | undefined;

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
              private route: ActivatedRoute,
              private toastr: ToastrService,
              private router: Router,
  ) {
    this.navService.setNavVisibility(true)

    this.route.queryParams.subscribe(params => {
      const pageIdParams = params['pageId'];
      if (!pageIdParams) {
        this.page = this.defaultPage;
        return;
      }

      try {
        const pageId = parseInt(pageIdParams);

        this.pageService.getPage(pageId).subscribe({
          next: page => {
            this.page = page;
          },
          error: error => {
            if (error.status === 404) {
              this.toastr.error("Page not found");
              this.router.navigateByUrl("/home");
              return;
            }

            this.toastr.error("Failed to retrieve page\n" + error.error.message, "Error");
          }
        })

      } catch (e) {
        console.error(e);
        this.page = this.defaultPage;
      }
    })
  }


  protected readonly PageWizardID = PageWizardID;
}
