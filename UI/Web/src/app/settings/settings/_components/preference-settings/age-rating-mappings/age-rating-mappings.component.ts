import {ChangeDetectorRef, Component, EventEmitter, Input, Output} from '@angular/core';
import {
  AgeRatingMap,
  ComicInfoAgeRating,
  ComicInfoAgeRatings,
  normalize,
  Preferences
} from "../../../../../_models/preferences";
import {Dialog} from "primeng/dialog";
import {TranslocoDirective} from "@jsverse/transloco";
import {FloatLabel} from "primeng/floatlabel";
import {FormsModule} from "@angular/forms";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {InputText} from "primeng/inputtext";
import {TitleCasePipe} from "@angular/common";
import {ToastService} from "../../../../../_services/toast.service";
import {Select} from "primeng/select";
import {VirtualScrollerModule} from "@iharbeck/ngx-virtual-scroller";

@Component({
  selector: 'app-age-rating-mappings',
  imports: [
    Dialog,
    TranslocoDirective,
    FloatLabel,
    FormsModule,
    IconField,
    InputIcon,
    InputText,
    TitleCasePipe,
    Select,
    VirtualScrollerModule
  ],
  templateUrl: './age-rating-mappings.component.html',
  styleUrl: './age-rating-mappings.component.css'
})
export class AgeRatingMappingsComponent {

  @Input({required: true}) preferences!: Preferences;
  @Output() preferencesChange: EventEmitter<Preferences> = new EventEmitter<Preferences>();

  @Input({required: true}) showDialog!: boolean;
  @Output() showDialogChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  filter: string = '';
  newAgeRating: string = '';
  toDisplay: AgeRatingMap[] = []

  constructor(
    private toastService: ToastService,
    private cdRef: ChangeDetectorRef,
  ) {
  }

  hide() {
    this.showDialog = false
    this.newAgeRating = '';
    this.filter = '';
    this.showDialogChange.emit(false);
  }

  removeAgeRatingMap(AgeRatingMap: AgeRatingMap) {
    if (!this.preferences) {
      return;
    }

    this.preferences.ageRatingMappings = this.preferences.ageRatingMappings
      .filter(ageRating => ageRating.tag.normalizedName !== AgeRatingMap.tag.normalizedName);
    this.filterToDisplay();
  }

  addAgeRatingMap() {
    if (!this.preferences) {
      return;
    }
    if (this.newAgeRating.length === 0) {
      return;
    }

    if (this.preferences.ageRatingMappings.find(ar => ar.tag.normalizedName === normalize(this.newAgeRating))) {
      this.newAgeRating = '';
      this.toastService.warningLoco("settings.preferences.toasts.age-rating-duplicate");
      return;
    }

    this.preferences.ageRatingMappings.push({
      tag:  {
        name: this.newAgeRating,
        normalizedName: normalize(this.newAgeRating),
      },
      comicInfoAgeRating: ComicInfoAgeRating.Unknown,
    });
    this.filterToDisplay();
    this.newAgeRating = ''
  }

  filterToDisplay() {
    if (!this.preferences) {
      return;
    }

    if (this.filter.length === 0) {
      this.toDisplay = this.preferences.ageRatingMappings;
      return;
    }

    this.toDisplay = this.preferences.ageRatingMappings
      .filter(ar => ar.tag.normalizedName.includes(normalize(this.filter)))
    this.cdRef.markForCheck();
  }

  protected readonly ComicInfoAgeRatings = ComicInfoAgeRatings;
}
