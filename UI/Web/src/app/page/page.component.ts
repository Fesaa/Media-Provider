import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {DownloadMetadata, Modifier, ModifierType, Page, Provider} from "../_models/page";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {SearchRequest} from "../_models/search";
import {SearchInfo} from "../_models/Info";
import {SearchResultComponent} from "./_components/search-result/search-result.component";
import {PaginatorComponent} from "../paginator/paginator.component";
import {dropAnimation} from "../_animations/drop-animation";
import {bounceIn500ms} from "../_animations/bounce-in";
import {flyInOutAnimation} from "../_animations/fly-animation";
import {DialogService} from "../_services/dialog.service";
import {fadeOut} from "../_animations/fade-out";
import {SubscriptionService} from "../_services/subscription.service";
import {ProviderNamePipe} from "../_pipes/provider-name.pipe";
import {ToastService} from "../_services/toast.service";
import {IconField} from "primeng/iconfield";
import {InputText} from "primeng/inputtext";
import {InputIcon} from "primeng/inputicon";
import {Select} from "primeng/select";
import {MultiSelect} from "primeng/multiselect";
import {FloatLabel} from "primeng/floatlabel";
import {ContentService} from "../_services/content.service";
import {TranslocoDirective} from "@jsverse/transloco";

@Component({
  selector: 'app-page',
  imports: [
    ReactiveFormsModule,
    SearchResultComponent,
    PaginatorComponent,
    IconField,
    InputText,
    InputIcon,
    Select,
    FormsModule,
    MultiSelect,
    FloatLabel,
    TranslocoDirective
  ],
  templateUrl: './page.component.html',
  styleUrl: './page.component.css',
  animations: [dropAnimation, bounceIn500ms, flyInOutAnimation, fadeOut]
})
export class PageComponent implements OnInit {

  page: Page | undefined = undefined;
  providers: Provider[] = [];
  metadata: Map<Provider, DownloadMetadata> = new Map();
  searchRequest!: SearchRequest;
  dirs: {dir: string, custom: string} = {dir: '', custom: ''};

  loading = false;

  searchResult: SearchInfo[] = [];
  currentPage: number = 1;
  showSearchForm: boolean = true;
  hideSearchForm: boolean = false;
  protected readonly ModifierType = ModifierType;
  protected readonly Math = Math;

  constructor(private navService: NavService,
              private pageService: PageService,
              private contentService: ContentService,
              private cdRef: ChangeDetectorRef,
              private toastService: ToastService,
              private dialogService: DialogService,
              private subscriptionService: SubscriptionService,
              private providerNamePipe: ProviderNamePipe,
  ) {
    this.navService.setNavVisibility(true);
  }

  ngOnInit(): void {
    this.navService.pageIndex$.subscribe(index => {
      this.pageService.getPage(index).subscribe(page => {
        this.hideSearchForm = true;
        this.cdRef.detectChanges();
        this.setup(page);


        setTimeout(() => {
          this.page = page;
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

  setup(page: Page) {
    this.searchResult = [];

    this.searchRequest = {
      query: '',
      provider: page.providers,
    }

    if (page.modifiers.length > 0) {
      this.searchRequest.modifiers = {}
      for (const mod of page.modifiers) {
        if (mod.values.length > 0) {
          switch (mod.type) {
            case ModifierType.DROPDOWN:
              this.searchRequest.modifiers[mod.key] = [mod.values[0].key]
              break;
            case ModifierType.MULTI:
              this.searchRequest.modifiers[mod.key] = []
              break;
          }
        }
      }
    }

    this.dirs.dir = page.dirs[0]
  }

  updateDropdownModifier(mod: Modifier, value: string) {
    this.searchRequest.modifiers![mod.key] = [value];
  }

  getDownloadMetadata(provider: Provider) {
    return this.metadata.get(provider)
  }

  search() {
    if (this.loading) {
      return;
    }

    this.loading = true;
    this.contentService.search(this.searchRequest).subscribe({
      next: info => {
        if (!info || info.length == 0) {
          this.toastService.errorLoco("page.toasts.no-results")
        } else {
          this.toastService.successLoco("page.toasts.search-success", {}, {amount: info.length});
        }
        this.searchResult = info || [];
        this.currentPage = 1;
        this.showSearchForm = false;
      },
      error: error => {
        this.toastService.genericError(error.error.message);
      }
    }).add(() => this.loading = false)
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
      this.dirs.custom = dir;
    }
  }

  clearCustomDir() {
    this.dirs.custom = '';
  }

  getDir() {
    if (this.dirs.custom && this.dirs.custom !== '') {
      return this.dirs.custom;
    }
    return this.dirs.dir;
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
          this.toastService.errorLoco("page.toasts.metadata-failed",
            {provider: this.providerNamePipe.transform(provider)}, {msg: error.error.message});
        }
      })
    }
  }
}
