import {ChangeDetectorRef, Component, HostListener, OnInit} from '@angular/core';
import {Modifier, Page, Provider, providerNames, providerValues} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {ConfigService} from "../../../../_services/config.service";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {ToastrService} from "ngx-toastr";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {DialogService} from "../../../../_services/dialog.service";
import {FormArray, FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {ModifierSettingsComponent} from "../modifier-settings/modifier-settings.component";
import {DirectorySettingsComponent} from "../directory-settings/directory-settings.component";
import {ProviderSettingsComponent} from "../provider-settings/provider-settings.component";
import {hasPermission, Perm, User} from "../../../../_models/user";
import {AccountService} from "../../../../_services/account.service";

@Component({
    selector: 'app-pages-settings',
    imports: [
        RouterLink,
        NgIcon,
        ReactiveFormsModule,
        FormInputComponent,
        ModifierSettingsComponent,
        DirectorySettingsComponent,
        ProviderSettingsComponent
    ],
    templateUrl: './pages-settings.component.html',
    styleUrl: './pages-settings.component.css',
    animations: [dropAnimation]
})
export class PagesSettingsComponent implements OnInit {

  user: User | null = null;
  pages: Page[] = []

  cooldown = false;
  selectedPageIndex = -1;
  selectedPage: Page | null = null;

  pageForm: FormGroup | undefined;

  showModifiers = false;
  isMobile = false;

  constructor(private configService: ConfigService,
              private toastR: ToastrService,
              private pageService: PageService,
              private dialogService: DialogService,
              private fb: FormBuilder,
              private cdRef: ChangeDetectorRef,
              private accountService: AccountService,
  ) {
    this.configService.getConfig().subscribe();
    this.pageService.pages$.subscribe(pages => this.pages = pages);
    this.accountService.currentUser$.subscribe(user => {
      if (user) {
        this.user = user;
      }
    });
  }

  @HostListener('window:resize', ['$event'])
  onResize() {
    this.isMobile = window.innerWidth < 768;
  }

  ngOnInit(): void {
    this.isMobile = window.innerWidth < 768;
  }

  setSelectedPage(index: number | undefined, page?: Page | null) {
    if (page === null) {
      this.selectedPage = null;
      this.selectedPageIndex = -1;
      this.buildForm();
      return;
    }
    if (page === undefined) {
      page = {
        ID: 0,
        sort_value: 0,
        dirs: [],
        title: '',
        modifiers: [],
        custom_root_dir: '',
        providers: [],
      }
    }
    this.pageForm = undefined;
    this.selectedPage = page;
    this.selectedPageIndex = index === undefined ? -1 : index;
    this.buildForm();
    this.cooldown = true;
    setTimeout(() => this.cooldown = false, 700);
  }

  buildForm() {
    if (this.selectedPage === null) {
      return;
    }

    this.pageForm = this.fb.group({
      title: this.fb.control(this.selectedPage.title, [Validators.required, Validators.minLength(3), Validators.maxLength(25)]),
      providers: this.fb.control(this.selectedPage.providers, [Validators.required]),
      dirs: this.fb.control(this.selectedPage.dirs, [Validators.required]),
      custom_root_dir: this.fb.control(this.selectedPage.custom_root_dir),
      modifiers: this.fb.control(this.selectedPage.modifiers),
    });
  }

  submit() {
    if (this.pageForm === undefined || this.selectedPage === null) {
      return;
    }

    if (this.pageForm.invalid) {
      this.displayErrors();
      return;
    }

    const page = this.pageForm.value as Page;
    page.ID = this.selectedPage.ID;
    page.sort_value = this.selectedPage.sort_value;
    // Filter some stuff out
    page.modifiers = page.modifiers
      .filter(m => m.key !== "");

    let obs;
    if (page.ID === 0) {
      obs = this.pageService.new(page);
    } else {
      obs = this.pageService.update(page)
    }

    obs.subscribe({
      next: () => {
        this.toastR.success(`${page.title} upserted`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(`Failed to upsert page ${err.error.error}`, 'Error');
      }
    });
    return;
  }

  private displayErrors() {
    let count = 0;
    Object.keys(this.pageForm!.controls).forEach(key => {
      const controlErrors = this.pageForm!.get(key)?.errors;
      if (controlErrors) {
        console.log(controlErrors);
        count += Object.keys(controlErrors).length;
      }
    });

    this.toastR.error(`Found ${count} errors in the form`, 'Cannot submit');
  }

  async remove(page: Page) {
    if (!await this.dialogService.openDialog('Are you sure you want to remove this page?')) {
      return;
    }

   this.pageService.removePage(page.ID).subscribe({
      next: () => {
        this.toastR.success(`${page.title} removed`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.error, 'Error');
      }
    });
  }

  moveUp(index: number) {
    const page1 = this.pages[index];
    const page2 = this.pages[index-1];
    this.swap(page1, page2);
  }

  moveDown(index: number) {
    const page1 = this.pages[index];
    const page2 = this.pages[index+1];
    this.swap(page1, page2);
  }

  swap(page1: Page, page2: Page) {
    this.pageService.swapPages(page1.ID, page2.ID).subscribe({
      next: () => {
        this.toastR.success(`Swapped ${page1.title} and ${page2.title}`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.error, 'Error');
      }
    });
  }


  protected readonly hasPermission = hasPermission;
  protected readonly Perm = Perm;
}
