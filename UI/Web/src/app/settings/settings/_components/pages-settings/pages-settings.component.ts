import { Component } from '@angular/core';
import {Page, Provider, providerNames, providerValues} from "../../../../_models/page";
import {PageService} from "../../../../_services/page.service";
import {ConfigService} from "../../../../_services/config.service";
import {RouterLink} from "@angular/router";
import {NgIcon} from "@ng-icons/core";
import {ToastrService} from "ngx-toastr";
import {dropAnimation} from "../../../../_animations/drop-animation";
import {DialogService} from "../../../../_services/dialog.service";
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";
import {KeyValuePipe, TitleCasePipe} from "@angular/common";

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
    TitleCasePipe
  ],
  templateUrl: './pages-settings.component.html',
  styleUrl: './pages-settings.component.css',
  animations: [dropAnimation],
})
export class PagesSettingsComponent {

  pages: Page[] = []

  cooldown = false;
  selectedPage: Page | null = null;

  pageForm: FormGroup | undefined;

  constructor(private configService: ConfigService,
              private toastR: ToastrService,
              private pageService: PageService,
              private dialogService: DialogService,
              private fb: FormBuilder
  ) {
    this.configService.getConfig().subscribe();
    this.pageService.pages$.subscribe(pages => this.pages = pages);
  }

  setSelectedPage(page?: Page) {
    if (page === undefined) {
      page = {
        dirs: [],
        title: '',
        modifiers: {},
        custom_root_dir: '',
        providers: [],
      }
    }
    this.pageForm = undefined;
    this.selectedPage = page;
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

  }

  submit() {
    console.log(this.pageForm?.value);
  }

  async remove(index: number) {
    if (!await this.dialogService.openDialog('Are you sure you want to remove this page?')) {
      return;
    }

   this.configService.removePage(index).subscribe({
      next: () => {
        const page = this.pages[index];
        this.toastR.success(`${page.title} removed`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

  moveUp(index: number) {
    this.configService.movePage(index, index - 1).subscribe({
      next: () => {
        const temp = this.pages[index];
        this.toastR.success(`${temp.title} moved up`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

  moveDown(index: number) {
    this.configService.movePage(index, index + 1).subscribe({
      next: () => {
        const temp = this.pages[index];
        this.toastR.success(`${temp.title} moved down`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(err.error.message, 'Error');
      }
    });
  }

  hasProvider(provider: Provider) {
    if (this.selectedPage === null) {
      return false;
    }
    return this.selectedPage.providers.includes(provider);
  }

  onProviderCheckboxChange(provider: number) {
    if (this.pageForm === undefined) {
      return;
    }
    const formArray = this.pageForm.controls['providers'];
    if (formArray.value.includes(provider)) {
      formArray.patchValue(formArray.value.filter((v: number) => v !== provider));
    } else {
      formArray.patchValue([...formArray.value, provider]);
    }
  }

  async updateCustomDir() {
    const newDir = await this.dialogService.openDirBrowser("");
    if (newDir === undefined) {
      return;
    }
    this.pageForm?.controls['custom_root_dir'].patchValue(newDir);
  }

  protected readonly Provider = Provider;
  protected readonly Object = Object;
  protected readonly providerValues = providerValues;
  protected readonly providerNames = providerNames;
}
