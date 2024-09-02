import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {PageService} from "../_services/page.service";
import {DownloadService} from "../_services/download.service";
import {Modifier, ModifierType, Page} from "../_models/page";
import {FormBuilder, FormGroup, ReactiveFormsModule} from "@angular/forms";
import {SearchRequest} from "../_models/search";
import {KeyValuePipe} from "@angular/common";
import {DropdownModifierComponent} from "./_components/dropdown-modifier/dropdown-modifier.component";
import {MultiModifierComponent} from "./_components/multi-modifier/multi-modifier.component";
import {SearchInfo} from "../_models/Info";
import {SearchResultComponent} from "./_components/search-result/search-result.component";
import {PaginatorComponent} from "../paginator/paginator.component";
import {dropAnimation} from "../_animations/drop-animation";
import {bounceIn500ms} from "../_animations/bounce-in";
import {ToastrService} from "ngx-toastr";

@Component({
  selector: 'app-page',
  standalone: true,
  imports: [
    KeyValuePipe,
    ReactiveFormsModule,
    DropdownModifierComponent,
    MultiModifierComponent,
    SearchResultComponent,
    PaginatorComponent
  ],
  templateUrl: './page.component.html',
  styleUrl: './page.component.css',
  animations: [dropAnimation, bounceIn500ms]
})
export class PageComponent implements OnInit{

  searchForm: FormGroup | undefined;
  page: Page | undefined = undefined;

  modifiers: Map<string, Modifier> = new Map<string, Modifier>();

  searchResult: SearchInfo[] = [];
  currentPage: number = 1;
  showSearchForm: boolean = true;
  hideSearchForm: boolean = false;

  constructor(private navService: NavService,
              private pageService: PageService,
              private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
              private fb: FormBuilder,
              private toastr: ToastrService
  ) {
    this.navService.setNavVisibility(true);
    this.downloadService.loadStats(false);
  }

  ngOnInit(): void {
    this.navService.pageIndex$.subscribe(index => {
      this.pageService.getPage(index).subscribe(page => {
        this.hideSearchForm = true;
        this.cdRef.detectChanges();

        // TODO: Don't wait if the search form is hidden, and add out animation for the search results?
        setTimeout(() => {
          this.page = page;
          this.modifiers = new Map(Object.entries(this.page.modifiers));
          this.searchResult = [];
          this.buildForm(page);
          this.hideSearchForm = false;
          this.cdRef.detectChanges()
        }, this.page === undefined ? 0 : 800)
      });
    })
  }

  buildForm(page: Page) {
    this.searchForm = this.fb.group({
      query: [''],
      customDir: [null],
    });

    if (page.dirs.length > 0) {
      this.searchForm.addControl('dir', this.fb.control(page.dirs[0]));
    }

    for (const [key, modifier] of Object.entries(page.modifiers)) {
      switch (modifier.type) {
        case ModifierType.DROPDOWN:
          const entries = Object.entries(modifier.values);
          if (entries.length > 0) {
            this.searchForm.addControl(key, this.fb.control(entries[0][0]));
          }
          break;
        case ModifierType.MULTI:
          this.searchForm.addControl(key, this.fb.control([]));
          break;
      }
    }
  }

  search() {
    if (!this.searchForm || !this.searchForm.valid || !this.page) {
      return;
    }
    const modifiers: { [key: string]: string[] } = {};
    for (const [key, modifier] of Object.entries(this.page.modifiers)) {
      const val = this.searchForm.value[key];
      if (val) {
        modifiers[key] = Array.isArray(val) ? val : [val];
      }
    }

    const req: SearchRequest = {
      query: this.searchForm.value.query,
      provider: this.page?.provider,
      modifiers: modifiers,
    };

    this.downloadService.search(req).subscribe(info => {
      if (info.length == 0) {
        this.toastr.error("No results found")
      } else {
        this.toastr.success(`Found ${info.length} items`,"Search completed")
      }
      this.searchResult = info;
      this.showSearchForm = false;
    })
  }

  toggleSearchForm() {
    this.showSearchForm = !this.showSearchForm;
    this.cdRef.detectChanges();
  }

  toShowResults(): SearchInfo[] {
    return this.searchResult.slice((this.currentPage - 1) * 10, this.currentPage * 10);
  }

  onPageChange(page: number) {
    this.currentPage = page;
    this.cdRef.detectChanges();
  }

  protected readonly ModifierType = ModifierType;
}
