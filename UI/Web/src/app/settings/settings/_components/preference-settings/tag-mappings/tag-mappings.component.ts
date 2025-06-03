import {ChangeDetectorRef, Component, EventEmitter, Input, Output} from '@angular/core';
import {normalize, Preferences, TagMap} from "../../../../../_models/preferences";
import {ToastService} from "../../../../../_services/toast.service";
import {Dialog} from "primeng/dialog";
import {TranslocoDirective} from "@jsverse/transloco";
import {FloatLabel} from "primeng/floatlabel";
import {InputIcon} from "primeng/inputicon";
import {IconField} from "primeng/iconfield";
import {InputText} from "primeng/inputtext";
import {FormsModule} from "@angular/forms";
import {TitleCasePipe} from "@angular/common";
import {CdkFixedSizeVirtualScroll, CdkVirtualForOf, CdkVirtualScrollViewport} from "@angular/cdk/scrolling";

@Component({
  selector: 'app-tag-mappings',
  imports: [
    Dialog,
    TranslocoDirective,
    FloatLabel,
    InputIcon,
    IconField,
    InputText,
    FormsModule,
    TitleCasePipe,
    CdkFixedSizeVirtualScroll,
    CdkVirtualForOf,
    CdkVirtualScrollViewport
  ],
  templateUrl: './tag-mappings.component.html',
  styleUrl: './tag-mappings.component.css'
})
export class TagMappingsComponent {

  @Input({required: true}) preferences!: Preferences;
  @Output() preferencesChange: EventEmitter<Preferences> = new EventEmitter<Preferences>();

  @Input({required: true}) showDialog!: boolean;
  @Output() showDialogChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  filter: string = '';
  newTag: string = '';
  toDisplay: TagMap[] = []

  constructor(
    private toastService: ToastService,
    private cdRef: ChangeDetectorRef,
  ) {
  }

  hide() {
    this.showDialog = false
    this.newTag = '';
    this.filter = '';
    this.showDialogChange.emit(false);
  }

  removeTagMap(tagMap: TagMap) {
    if (!this.preferences) {
      return;
    }

    this.preferences.tagMappings = this.preferences.tagMappings
      .filter(ageRating => ageRating.origin.normalizedName !== tagMap.origin.normalizedName);
    this.filterToDisplay();
  }

  addTagMap() {
    if (!this.preferences) {
      return;
    }
    if (this.newTag.length === 0) {
      return;

    }

    if (this.preferences.tagMappings.find(ar => ar.origin.normalizedName === normalize(this.newTag))) {
      this.newTag = '';
      this.toastService.warningLoco("settings.preferences.toasts.age-rating-duplicate");
      return;
    }

    this.preferences.tagMappings.push({
      origin:  {
        name: this.newTag,
        normalizedName: normalize(this.newTag),
      },
      dest: {
        name: '',
        normalizedName: '',
      }
    });
    this.filterToDisplay();
    this.newTag = ''
  }

  filterToDisplay() {
    if (!this.preferences) {
      return;
    }

    if (this.filter.length === 0) {
      this.toDisplay = this.preferences.tagMappings;
      return;
    }

    this.toDisplay = this.preferences.tagMappings
      .filter(ar => ar.origin.normalizedName.includes(normalize(this.filter)))
    this.cdRef.markForCheck();
  }

}
