import {Component, EventEmitter, Input, Output} from '@angular/core';
import {normalize, Preferences, Tag} from "../../../../../_models/preferences";
import {ToastService} from "../../../../../_services/toast.service";
import {Dialog} from "primeng/dialog";
import {TranslocoDirective} from "@jsverse/transloco";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {FormsModule} from "@angular/forms";
import {InputText} from "primeng/inputtext";
import {TitleCasePipe} from "@angular/common";
import {VirtualScrollerModule} from "@iharbeck/ngx-virtual-scroller";

@Component({
  selector: 'app-white-list-tags',
  imports: [
    Dialog,
    TranslocoDirective,
    FloatLabel,
    IconField,
    InputIcon,
    FormsModule,
    InputText,
    TitleCasePipe,
    VirtualScrollerModule
  ],
  templateUrl: './white-list-tags.component.html',
  styleUrl: './white-list-tags.component.css'
})
export class WhiteListTagsComponent {

  @Input({required: true}) preferences!: Preferences;
  @Output() preferencesChange: EventEmitter<Preferences> = new EventEmitter<Preferences>();

  @Input({required: true}) showDialog!: boolean;
  @Output() showDialogChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  newTag: string = '';
  filter: string = '';
  toDisplay: Tag[] = [];

  constructor(
    private toastService: ToastService,
  ) {
  }

  hide() {
    this.showDialog = false;
    this.newTag = '';
    this.filter = '';
    this.showDialogChange.emit(false);
  }

  removeTag(tag: Tag) {
    if (!this.preferences) {
      return;
    }
    this.preferences.whiteListedTags = this.preferences.whiteListedTags
      .filter(g => g.normalizedName !== tag.normalizedName);
    this.filterToDisplay();
  }

  addTag() {
    if (!this.preferences) {
      return;
    }
    if (this.newTag.length === 0) {
      return;
    }
    if (this.preferences.whiteListedTags.find(g => g.normalizedName === normalize(this.newTag))) {
      this.newTag = ''
      this.toastService.warningLoco("settings.preferences.toasts.blacklist-duplicate");
      return;
    }
    this.preferences.whiteListedTags.push({
      name: this.newTag,
      normalizedName: normalize(this.newTag),
    });
    this.filterToDisplay();
    this.newTag = ''
  }

  filterToDisplay() {
    if (!this.preferences) {
      return;
    }

    if (this.filter.length === 0) {
      this.toDisplay = this.preferences.whiteListedTags;
      return;
    }

    this.toDisplay = this.preferences.whiteListedTags
      .filter(g => g.normalizedName.includes(normalize(this.filter)));
  }

}
