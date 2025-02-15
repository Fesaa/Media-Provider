import {Component, EventEmitter, Input, Output} from '@angular/core';
import {Preferences} from "../../../../../_models/preferences";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {CdkFixedSizeVirtualScroll, CdkVirtualForOf, CdkVirtualScrollViewport} from "@angular/cdk/scrolling";
import {FormsModule} from "@angular/forms";
import {InputText} from "primeng/inputtext";

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
    InputText
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
  dynastyToDisplayGenres: string[] = [];

  hide() {
    this.showDialog = false;
    this.dynastyGenresNew = '';
    this.dynastyFilter = '';
    this.showDialogChange.emit(false);
  }

  removeGenre(genre: string) {
    if (!this.preferences) {
      return;
    }
    this.preferences.dynastyGenreTags = this.preferences.dynastyGenreTags.filter(g => g !== genre);
    this.dynastyGenresFiltered();
  }

  addGenre() {
    if (!this.preferences) {
      return;
    }
    if (this.dynastyGenresNew.length === 0) {
      return;
    }
    if (this.preferences.dynastyGenreTags.find(g => g === this.dynastyGenresNew)) {
      this.dynastyGenresNew = ''
      return;
    }
    this.preferences.dynastyGenreTags = [...this.preferences.dynastyGenreTags, this.dynastyGenresNew];
    this.dynastyGenresFiltered();
    this.dynastyGenresNew = ''
  }

  dynastyGenresFiltered() {
    if (!this.preferences) {
      return;
    }
    this.dynastyToDisplayGenres = this.preferences.dynastyGenreTags
      .filter(g => g.toLowerCase().includes(this.dynastyFilter.toLowerCase()));
  }

}
