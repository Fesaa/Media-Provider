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
import {flyInOutAnimation} from "../_animations/fly-animation";
import {DirectoryBrowserComponent} from "../directory-browser/directory-browser.component";
import {NgIcon} from "@ng-icons/core";
import {FormInputComponent} from "../shared/form/form-input/form-input.component";
import {DialogService} from "../_services/dialog.service";
import {fadeOut} from "../_animations/fade-out";

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
export class PageComponent implements OnInit{

  searchForm: FormGroup | undefined;
  page: Page | undefined = undefined;

  modifiers: Modifier[] = [];

  searchResult: SearchInfo[] = [];
  currentPage: number = 1;
  showSearchForm: boolean = true;
  hideSearchForm: boolean = false;

  constructor(private navService: NavService,
              private pageService: PageService,
              private downloadService: DownloadService,
              private cdRef: ChangeDetectorRef,
              private fb: FormBuilder,
              private toastr: ToastrService,
              private dialogService: DialogService,
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

    if (page.custom_root_dir) {
      this.searchForm.addControl('customDir', this.fb.control(null));
    }

    for (const modifier of page.modifiers) {
      switch (modifier.type) {
        case ModifierType.DROPDOWN:
          const entries = Object.entries(modifier.values);
          if (entries.length > 0) {
            this.searchForm.addControl(modifier.key, this.fb.control(entries[0][0]));
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
      this.toastr.error(`Cannot search`, "Error");
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

    this.downloadService.search(req).subscribe(info => {
      if (info.length == 0) {
        this.toastr.error("No results found")
      } else {
        this.toastr.success(`Found ${info.length} items`,"Search completed")
      }
      this.searchResult = info;
      this.currentPage = 1;
      this.showSearchForm = false;
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

    const dir = await this.dialogService.openDirBrowser(this.page.custom_root_dir);
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

  protected readonly ModifierType = ModifierType;
  protected readonly Math = Math;
}
