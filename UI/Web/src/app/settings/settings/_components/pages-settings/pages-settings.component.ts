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
export class PagesSettingsComponent implements OnInit {

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
  ) {
    this.configService.getConfig().subscribe();
    this.pageService.pages$.subscribe(pages => this.pages = pages);
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
    if (this.pageForm === undefined) {
      return;
    }

    if (this.pageForm.invalid) {
      this.displayErrors();
      return;
    }

    const page = this.pageForm.value as Page;
    if (this.selectedPageIndex === -1) {
      this.configService.addPage(page).subscribe({
        next: () => {
          this.toastR.success(`${page.title} added`, 'Success');
          this.pageService.refreshPages();
        },
        error: (err) => {
          this.toastR.error(`Failed to add page ${err.error.error}`, 'Error');
        }
      });
      return;
    }

    this.configService.updatePage(page, this.selectedPageIndex).subscribe({
      next: () => {
        this.toastR.success(`${page.title} updated`, 'Success');
        this.pageService.refreshPages();
      },
      error: (err) => {
        this.toastR.error(`Failed to update page ${err.error.error}`, 'Error');
      }
    });
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

  getDirs() {
    return this.pageForm?.controls['dirs'].value;
  }

  updateDir(index: number, e: Event) {
    const dir = (e.target as HTMLInputElement).value;
    this.updateInArray(this.pageForm?.controls['dirs'] as FormArray, dir, index);
  }

  async getNewDir(index: number) {
    const dir = await this.dialogService.openDirBrowser("");
    if (dir === undefined) {
      return;
    }
    this.updateInArray(this.pageForm?.controls['dirs'] as FormArray, dir, index);
  }

  async removeDir(index: number) {
    const dirs = this.pageForm?.controls['dirs'] as FormArray;
    const values = dirs.value;
    if (index >= values.length) {
      this.toastR.error('Invalid index', 'Error');
      return;
    }

    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${values[index]}?`)) {
      return;
    }

    if (dirs.value.length === 1) {
      this.toastR.error('You must have at least one directory', 'Error');
      return;
    }

    dirs.patchValue(values.filter((_: any, i: number) => i !== index));
    this.toastR.warning(`Removed directory ${values[index]}`, 'Success');
  }

  private updateInArray(formArray: FormArray, value: any, index: number) {
    const values = formArray.value;

    if (index >= values.length) {
      const find = values.find((v: any) => v === value);
      if (find !== undefined) {
        this.toastR.info('Directory already added', 'Nothing happened');
        return;
      }

      values.push(value);
      formArray.patchValue(values);
      this.toastR.success(`Added directory ${value}`, 'Success');
      return;
    }

    values[index] = value;
    formArray.patchValue(values);
  }

  toggleModifiers() {
    this.showModifiers = !this.showModifiers;
    this.cdRef.detectChanges();
  }

  getModifiers() {
    const modifiers: {[key: string]: Modifier} = {};
    if (this.pageForm === undefined) {
      return modifiers;
    }

    const form = this.pageForm.controls['modifiers'] as FormGroup;
    for (const [key, value] of Object.entries(form.controls)) {
      modifiers[key] = value.value;
    }

    return modifiers;
  }

  addModifier() {
    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    modifierGroup.addControl('modifier', this.fb.group({
      title: this.fb.control('', [Validators.required]),
      type: this.fb.control('string', [Validators.required]),
      values: this.fb.control({}),
    }));
  }

  updateModifierTitle(key: string, e: Event) {
    const title = (e.target as HTMLInputElement).value;

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    modifier.controls['title'].patchValue(title);
  }

  updateModifierKey(key: string, e: Event) {
    const newKey = (e.target as HTMLInputElement).value;

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    modifierGroup.removeControl(key);
    modifierGroup.addControl(newKey, modifier);
  }

  async removeModifier(key: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${key}?`)) {
      return;
    }

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    modifierGroup.removeControl(key);
    this.toastR.warning(`Removed modifier ${key}`, 'Success');
  }

  async removeModifierValue(key: string, valueKey: string) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove ${valueKey}?`)) {
      return;
    }

    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    const values = modifier.controls['values'];
    delete values.value[valueKey];
    values.patchValue(values.value);
    this.toastR.warning(`Removed value ${valueKey}`, 'Success');
  }

  addModifierValue(key: string) {
    const modifierGroup = this.pageForm?.controls['modifiers'] as FormGroup;
    const modifier = modifierGroup.controls[key] as FormGroup;
    const values = modifier.controls['values'];
    values.value['key'] = 'value';
  }


  protected readonly providerValues = providerValues;
  protected readonly providerNames = providerNames;
}
