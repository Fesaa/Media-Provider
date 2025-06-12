import {Component, EventEmitter, Input, Output} from '@angular/core';
import {normalize, Preferences, Tag} from "../../../../../_models/preferences";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputText} from "primeng/inputtext";
import {InputIcon} from "primeng/inputicon";
import {FormsModule} from "@angular/forms";
import {ToastService} from "../../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";
import {VirtualScrollerModule} from "@iharbeck/ngx-virtual-scroller";

@Component({
  selector: 'app-tags-blacklist',
  imports: [
    Dialog,
    FloatLabel,
    IconField,
    InputText,
    InputIcon,
    FormsModule,
    TranslocoDirective,
    TitleCasePipe,
    VirtualScrollerModule
  ],
  templateUrl: './tags-blacklist.component.html',
  styleUrl: './tags-blacklist.component.css'
})
export class TagsBlacklistComponent {

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
    this.preferences.blackListedTags = this.preferences.blackListedTags
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
    if (this.preferences.blackListedTags.find(g => g.normalizedName === normalize(this.newTag))) {
      this.newTag = ''
      this.toastService.warningLoco("settings.preferences.toasts.blacklist-duplicate");
      return;
    }
    this.preferences.blackListedTags.push({
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
      this.toDisplay = this.preferences.blackListedTags;
      return;
    }

    this.toDisplay = this.preferences.blackListedTags
      .filter(g => g.normalizedName.includes(normalize(this.filter)));
  }

}
