import {Component} from '@angular/core';
import {PageService} from "../../../../_services/page.service";
import {NavService} from "../../../../_services/nav.service";
import {Page} from "../../../../_models/page";
import {Steps} from "primeng/steps";
import {PageWizardGeneralComponent} from "./_components/page-wizard-general/page-wizard-general.component";
import {PageWizardDirsComponent} from "./_components/page-wizard-dirs/page-wizard-dirs.component";
import {PageWizardModifiersComponent} from "./_components/page-wizard-modifiers/page-wizard-modifiers.component";
import {ActivatedRoute, NavigationExtras, Router} from "@angular/router";
import {Skeleton} from "primeng/skeleton";
import {Card} from "primeng/card";
import {PageWizardSaveComponent} from "./_components/page-wizard-save/page-wizard-save.component";
import {ToastService} from "../../../../_services/toast.service";

export enum PageWizardID {
  General = 'General',
  Dirs = 'Dirs',
  Modifiers = 'Modifiers',
  Save = 'Save',
}

@Component({
  selector: 'app-page-wizard',
  imports: [
    Steps,
    PageWizardGeneralComponent,
    PageWizardDirsComponent,
    PageWizardModifiersComponent,
    Skeleton,
    Card,
    PageWizardSaveComponent
  ],
  templateUrl: './page-wizard.component.html',
  styleUrl: './page-wizard.component.css'
})
export class PageWizardComponent {

  page: Page | undefined;
  index: number = 0;
  sections: { id: PageWizardID, label: string }[] = [
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
    {
      id: PageWizardID.Save,
      label: "Save"
    }
  ];
  protected readonly PageWizardID = PageWizardID;
  private readonly defaultPage: Page = {
    ID: 0,
    title: "",
    custom_root_dir: "",
    icon: "",
    dirs: [],
    providers: [],
    modifiers: [],
    sortValue: PageService.DEFAULT_PAGE_SORT,
  }

  constructor(private pageService: PageService,
              private navService: NavService,
              private route: ActivatedRoute,
              private toastService: ToastService,
              private router: Router,
  ) {
    this.navService.setNavVisibility(true)

    this.route.queryParams.subscribe(params => {
      const pageIdParams = params['pageId'];
      if (!pageIdParams) {
        this.page = this.defaultPage;
        return;
      }

      let pageId;
      try {
        pageId = parseInt(pageIdParams);
      } catch (e) {
        console.error(e);
        this.router.navigateByUrl("/home")
        return;
      }

      this.pageService.getPage(pageId).subscribe({
        next: page => {
          this.page = page;
        },
        error: error => {
          if (error.status === 404) {
            this.toastService.warningLoco("page.toasts.not-found");
            this.router.navigateByUrl("/home");
            return;
          }

          this.toastService.genericError(error.error.message);
        }
      })
    })
    this.route.fragment.subscribe(fragment => {
      const section = this.sections.filter(section => section.id == fragment)
      if (section && section.length > 0) {
        this.navigateToPage(this.sections.indexOf(section[0]))
      } else {
        this.navigateToPage(0)
      }
    })
  }

  navigateToPage(index: number) {
    this.index = index;

    const sectionId = this.sections[this.index].id;

    const extras: NavigationExtras = {
      fragment: sectionId
    };

    if (this.page && this.page.ID !== 0) {
      extras.queryParams = {
        pageId: this.page.ID
      }
    }

    this.router.navigate([], extras)
  }
}
