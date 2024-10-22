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
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";
import {KeyValuePipe, TitleCasePipe} from "@angular/common";
import {ModifierSettingsComponent} from "../modifier-settings/modifier-settings.component";
import {DirectorySettingsComponent} from "../directory-settings/directory-settings.component";
import {ProviderSettingsComponent} from "../provider-settings/provider-settings.component";
import {hasPermission, Perm, User} from "../../../../_models/user";
import {AccountService} from "../../../../_services/account.service";

@Component({
  selector: 'app-pages-settings',
  standalone: true,
  imports: [
    RouterLink,
    NgIcon,
    ReactiveFormsModule,
    FormInputComponent,
    FormSelectComponent,
    KeyValuePipe,
    TitleCasePipe,
    ModifierSettingsComponent,
    DirectorySettingsComponent,
    ProviderSettingsComponent
  ],
  templateUrl: './pages-settings.component.html',
  styleUrl: './pages-settings.component.css',
  animations: [dropAnimation],
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
        id: 0,
        dirs: [],
        title: '',
        modifiers: {},
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
    });

    const modifiers = this.fb.group({});
    for (const [key, value] of Object.entries(this.selectedPage.modifiers)) {
      modifiers.addControl(key, this.fb.group({
        title: this.fb.control(value.title, [Validators.required]),
        type: this.fb.control(value.type, [Validators.required]),
        values: this.fb.control(value.values, [Validators.required]),
      }));
    }

    this.pageForm.addControl('modifiers', modifiers);
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
    page.id = this.selectedPage.id;
    this.pageService.upsertPage(page).subscribe({
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

   this.pageService.removePage(page.id).subscribe({
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
    this.pageService.swapPages(page1.id, page2.id).subscribe({
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
