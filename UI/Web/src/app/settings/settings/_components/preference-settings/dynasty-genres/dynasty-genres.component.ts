import {Component, EventEmitter, Input, Output} from '@angular/core';
import {normalize, Preferences, Tag} from "../../../../../_models/preferences";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {CdkFixedSizeVirtualScroll, CdkVirtualForOf, CdkVirtualScrollViewport} from "@angular/cdk/scrolling";
import {FormsModule} from "@angular/forms";
import {InputText} from "primeng/inputtext";
import {ToastService} from "../../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";

@Component({
  selector: 'app-dynasty-genres',
  imports: [
    Dialog,
    FloatLabel,
    IconField,
    InputIcon,
    CdkVirtualScrollViewport,
    CdkVirtualForOf,
    FormsModule,
    CdkFixedSizeVirtualScroll,
    InputText,
    TranslocoDirective,
    TitleCasePipe
  ],
  templateUrl: './dynasty-genres.component.html',
  styleUrl: './dynasty-genres.component.css'
})
export class DynastyGenresComponent {

  @Input({required: true}) preferences!: Preferences;
  @Output() preferencesChange: EventEmitter<Preferences> = new EventEmitter<Preferences>();

  @Input({required: true}) showDialog!: boolean;
  @Output() showDialogChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  dynastyGenresNew: string = '';
  dynastyFilter: string = '';
  dynastyToDisplayGenres: Tag[] = [];

  constructor(
    private toastService: ToastService,
  ) {
  }

  hide() {
    this.showDialog = false;
    this.dynastyGenresNew = '';
    this.dynastyFilter = '';
    this.showDialogChange.emit(false);
  }

  removeGenre(genre: Tag) {
    if (!this.preferences) {
      return;
    }
    this.preferences.dynastyGenreTags = this.preferences.dynastyGenreTags
      .filter(g => g.normalizedName !== genre.normalizedName);
    this.dynastyGenresFiltered();
  }

  addGenre() {
    if (!this.preferences) {
      return;
    }
    if (this.dynastyGenresNew.length === 0) {
      return;
    }
    if (this.preferences.dynastyGenreTags.find(g => g.normalizedName === normalize(this.dynastyGenresNew))) {
      this.dynastyGenresNew = ''
      this.toastService.warningLoco("settings.preferences.toasts.dynasty-duplicate");
      return;
    }
    this.preferences.dynastyGenreTags = [...this.preferences.dynastyGenreTags, {
      name: this.dynastyGenresNew,
      normalizedName: normalize(this.dynastyGenresNew)
    }];
    this.dynastyGenresFiltered();
    this.dynastyGenresNew = ''
  }

  dynastyGenresFiltered() {
    if (!this.preferences) {
      return;
    }
    this.dynastyToDisplayGenres = this.preferences.dynastyGenreTags
      .filter(g => g.normalizedName.includes(normalize(this.dynastyFilter)));
  }

}
