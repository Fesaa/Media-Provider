import {ChangeDetectorRef, Component, inject, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {DownloadMetadata, Modifier, ModifierType, Page, Provider} from "../_models/page";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {SearchRequest} from "../_models/search";
import {SearchInfo} from "../_models/Info";
import {SearchResultComponent} from "./_components/search-result/search-result.component";
import {dropAnimation} from "../_animations/drop-animation";
import {bounceIn500ms} from "../_animations/bounce-in";
import {flyInOutAnimation} from "../_animations/fly-animation";
import {fadeOut} from "../_animations/fade-out";
import {SubscriptionService} from "../_services/subscription.service";
import {ProviderNamePipe} from "../_pipes/provider-name.pipe";
import {ToastService} from "../_services/toast.service";
import {ContentService} from "../_services/content.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {ModalService} from "../_services/modal.service";

@Component({
  selector: 'app-page',
  imports: [
    ReactiveFormsModule,
    SearchResultComponent,
    FormsModule,
    TranslocoDirective,
  ],
  templateUrl: './page.component.html',
  styleUrl: './page.component.scss',
  animations: [dropAnimation, bounceIn500ms, flyInOutAnimation, fadeOut]
})
export class PageComponent implements OnInit {

  private readonly modalService = inject(ModalService);

  page: Page | undefined = undefined;
  providers: Provider[] = [];
  metadata: Map<Provider, DownloadMetadata> = new Map();
  searchRequest!: SearchRequest;
  dirs: {dir: string, custom: string} = {dir: '', custom: ''};

  loading = false;

  searchResult: SearchInfo[] = [];
  currentPage: number = 0;
  showSearchForm: boolean = true;
  hideSearchForm: boolean = false;
  protected readonly ModifierType = ModifierType;
  protected readonly Math = Math;

  constructor(private navService: NavService,
              private pageService: PageService,
              private contentService: ContentService,
              private cdRef: ChangeDetectorRef,
              private toastService: ToastService,
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
              const defaultValue = mod.values.find((v) => v.default);
              this.searchRequest.modifiers[mod.key] = defaultValue ? [defaultValue.key] : [];
              break;
            case ModifierType.MULTI:
              const defaultValues = mod.values
                .filter(v => v.default)
                .map(v => v.key);
              this.searchRequest.modifiers[mod.key] = defaultValues;
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
        this.currentPage = 0;
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

    // TODO: Dir selector
    /*const dir = await this.dialogService.openDirBrowser(this.page.custom_root_dir, {create: true,});
    if (dir) {
      this.dirs.custom = dir;
    }*/
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
    return this.searchResult.slice((this.currentPage) * 10, (this.currentPage+1) * 10);
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
