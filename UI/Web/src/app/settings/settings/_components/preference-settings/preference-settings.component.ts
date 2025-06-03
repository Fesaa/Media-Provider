import {Component, OnInit} from '@angular/core';
import {PreferencesService} from '../../../../_services/preferences.service';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {CoverFallbackMethods, Preferences} from "../../../../_models/preferences";
import {Tooltip} from "primeng/tooltip";
import {ToastService} from '../../../../_services/toast.service';
import {ToggleSwitch} from "primeng/toggleswitch";
import {InputNumber} from "primeng/inputnumber";
import {Button} from "primeng/button";
import {DynastyGenresComponent} from "./dynasty-genres/dynasty-genres.component";
import {TagsBlacklistComponent} from "./tags-blacklist/tags-blacklist.component";
import {TranslocoDirective} from "@jsverse/transloco";
import {Select} from "primeng/select";
import {AgeRatingMappingsComponent} from "./age-rating-mappings/age-rating-mappings.component";
import {WhiteListTagsComponent} from "./white-list-tags/white-list-tags.component";

@Component({
  selector: 'app-preference-settings',
  standalone: true,
  imports: [
    FormsModule,
    ReactiveFormsModule,
    Tooltip,
    ToggleSwitch,
    InputNumber,
    Button,
    DynastyGenresComponent,
    TagsBlacklistComponent,
    TranslocoDirective,
    Select,
    AgeRatingMappingsComponent,
    WhiteListTagsComponent,
  ],
  templateUrl: './preference-settings.component.html',
  styleUrl: './preference-settings.component.css'
})
export class PreferenceSettingsComponent implements OnInit {

  preferences: Preferences | undefined;
  displayDynastyGenresDialog: boolean = false;
  displayBlackListTagDialog: boolean = false;
  displayWhiteListTagDialog: boolean = false;
  displayAgeRatingMappingDialog: boolean = false;

  constructor(private preferencesService: PreferencesService,
              private toastService: ToastService,
  ) {
  }

  ngOnInit(): void {
    this.preferencesService.get().subscribe(preferences => {
      this.preferences = preferences;
    })
  }

  save() {
    if (!this.preferences) {
      return;
    }

    this.preferencesService.save(this.preferences).subscribe({
      next: () => {
        this.toastService.successLoco("settings.preferences.toasts.save.success");
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    })
  }

  protected readonly CoverFallbackMethods = CoverFallbackMethods;
}
