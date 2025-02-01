import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {ContentService} from "../_services/content.service";
import {DownloadMetadata, Modifier, ModifierType, Page, Provider} from "../_models/page";
import {FormBuilder, FormGroup, ReactiveFormsModule} from "@angular/forms";
import {SearchRequest} from "../_models/search";
import {DropdownModifierComponent} from "./_components/dropdown-modifier/dropdown-modifier.component";
import {MultiModifierComponent} from "./_components/multi-modifier/multi-modifier.component";
import {SearchInfo} from "../_models/Info";
import {SearchResultComponent} from "./_components/search-result/search-result.component";
import {PaginatorComponent} from "../paginator/paginator.component";
import {dropAnimation} from "../_animations/drop-animation";
import {bounceIn500ms} from "../_animations/bounce-in";
import {flyInOutAnimation} from "../_animations/fly-animation";
import {NgIcon} from "@ng-icons/core";
import {FormInputComponent} from "../shared/form/form-input/form-input.component";
import {DialogService} from "../_services/dialog.service";
import {fadeOut} from "../_animations/fade-out";
import {SubscriptionService} from "../_services/subscription.service";
import {ProviderNamePipe} from "../_pipes/provider-name.pipe";
import {MessageService} from "../_services/message.service";

@Component({
  selector: 'app-page',
  imports: [
    ReactiveFormsModule,
    DropdownModifierComponent,
    MultiModifierComponent,
    SearchResultComponent,
    PaginatorComponent,
    NgIcon,
    FormInputComponent
  ],
  templateUrl: './page.component.html',
  styleUrl: './page.component.css',
  animations: [dropAnimation, bounceIn500ms, flyInOutAnimation, fadeOut]
})
export class PageComponent implements OnInit {

  searchForm: FormGroup | undefined;
  page: Page | undefined = undefined;

  modifiers: Modifier[] = [];
  providers: Provider[] = [];
  metadata: Map<Provider, DownloadMetadata> = new Map();

  searchResult: SearchInfo[] = [];
  currentPage: number = 1;
  showSearchForm: boolean = true;
  hideSearchForm: boolean = false;
  protected readonly ModifierType = ModifierType;
  protected readonly Math = Math;

  constructor(private navService: NavService,
              private pageService: PageService,
              private downloadService: ContentService,
              private cdRef: ChangeDetectorRef,
              private fb: FormBuilder,
              private msgService: MessageService,
              private dialogService: DialogService,
              private subscriptionService: SubscriptionService,
              private providerNamePipe: ProviderNamePipe,
  ) {
    this.navService.setNavVisibility(true);
    this.downloadService.loadStats(false);
  }

  ngOnInit(): void {
    this.navService.pageIndex$.subscribe(index => {
      this.pageService.getPage(index).subscribe(page => {
        this.hideSearchForm = true;
        this.cdRef.detectChanges();
        this.searchResult = [];

        setTimeout(() => {
          this.page = page;
          this.modifiers = this.page.modifiers;
          this.buildForm(page);
          this.hideSearchForm = false;
          this.loadMetadata()
          this.cdRef.detectChanges()
        }, this.page === undefined ? 0 : 800)
      });
    })

    this.subscriptionService.providers().subscribe(providers => {
      this.providers = providers;
    })
  }

  getDownloadMetadata(provider: Provider) {
    return this.metadata.get(provider)
  }

  buildForm(page: Page) {
    this.searchForm = this.fb.group({
      query: [''],
      customDir: [null],
    });

    if (page.dirs.length > 0) {
      this.searchForm.addControl('dir', this.fb.control(page.dirs[0]));
    }

    if (page.custom_root_dir) {
      this.searchForm.addControl('customDir', this.fb.control(null));
    }

    for (const modifier of page.modifiers) {
      switch (modifier.type) {
        case ModifierType.DROPDOWN:
          if (modifier.values.length > 0) {
            this.searchForm.addControl(modifier.key, this.fb.control(modifier.values[0].key));
          }
          break;
        case ModifierType.MULTI:
          this.searchForm.addControl(modifier.key, this.fb.control([]));
          break;
      }
    }
  }

  search() {
    if (!this.searchForm || !this.searchForm.valid || !this.page) {
      this.msgService.error("Error", `Cannot search`);
      return;
    }
    const modifiers: { [key: string]: string[] } = {};
    for (const modifier of this.page.modifiers) {
      const val = this.searchForm.value[modifier.key];
      if (val) {
        modifiers[modifier.key] = Array.isArray(val) ? val : [val];
      }
    }

    const req: SearchRequest = {
      query: this.searchForm.value.query,
      provider: this.page?.providers,
      modifiers: modifiers,
    };

    this.downloadService.search(req).subscribe({
      next: info => {
        if (!info || info.length == 0) {
          this.msgService.error("No results found")
        } else {
          this.msgService.success("Search completed", `Found ${info.length} items`)
        }
        this.searchResult = info || [];
        this.currentPage = 1;
        this.showSearchForm = false;
      },
      error: error => {
        this.msgService.error("Search failed", error.error.message);
      }
    })
  }

  toggleSearchForm() {
    this.showSearchForm = !this.showSearchForm;
    this.cdRef.detectChanges();
  }

  async selectCustomDir() {
    if (!this.page) {
      return;
    }

    const dir = await this.dialogService.openDirBrowser(this.page.custom_root_dir, {create: true,});
    if (dir) {
      this.searchForm?.get('customDir')?.setValue(dir);
    }
  }

  clearCustomDir() {
    this.searchForm?.get('customDir')?.setValue(null);
  }

  toShowResults(): SearchInfo[] {
    return this.searchResult.slice((this.currentPage - 1) * 10, this.currentPage * 10);
  }

  onPageChange(page: number) {
    this.currentPage = page;
    this.cdRef.detectChanges();
  }

  private loadMetadata() {
    if (!this.page) {
      return;
    }

    for (const provider of this.page.providers) {
      this.pageService.metadata(provider).subscribe({
        next: metadata => {
          this.metadata.set(provider, metadata);
        },
        error: error => {
          this.msgService.error(`Failed to load download metadata for: ${this.providerNamePipe.transform(provider)}`, error.error.message)
        }
      })
    }
  }
}
