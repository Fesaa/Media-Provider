import {Component, EventEmitter, Input, Output} from '@angular/core';
import {normalize, Preferences, Tag} from "../../../../../_models/preferences";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputText} from "primeng/inputtext";
import {InputIcon} from "primeng/inputicon";
import {FormsModule} from "@angular/forms";
import {CdkFixedSizeVirtualScroll, CdkVirtualForOf, CdkVirtualScrollViewport} from "@angular/cdk/scrolling";
import {MessageService} from "../../../../../_services/message.service";

@Component({
  selector: 'app-tags-blacklist',
  imports: [
    Dialog,
    FloatLabel,
    IconField,
    InputText,
    InputIcon,
    FormsModule,
    CdkVirtualScrollViewport,
    CdkFixedSizeVirtualScroll,
    CdkVirtualForOf
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
    private msgService: MessageService,
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
      this.msgService.warning("Tag already present", "Tags are normalized, may not find the exact tag in the list")
      return;
    }
    this.preferences.blackListedTags = [...this.preferences.blackListedTags, {
      name: this.newTag,
      normalizedName: normalize(this.newTag),
    }];
    this.filterToDisplay();
    this.newTag = ''
  }

  filterToDisplay() {
    if (!this.preferences) {
      return;
    }
    this.toDisplay = this.preferences.blackListedTags
      .filter(g => g.normalizedName.includes(normalize(this.filter)));
  }

}
