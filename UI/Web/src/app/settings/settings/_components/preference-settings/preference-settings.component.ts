import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {PreferencesService} from '../../../../_services/preferences.service';
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Preferences} from "../../../../_models/preferences";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from '../../../../_services/message.service';
import {ToggleSwitch} from "primeng/toggleswitch";
import {InputNumber} from "primeng/inputnumber";
import {Button} from "primeng/button";
import {DynastyGenresComponent} from "./dynasty-genres/dynasty-genres.component";
import {TagsBlacklistComponent} from "./tags-blacklist/tags-blacklist.component";

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
  ],
  templateUrl: './preference-settings.component.html',
  styleUrl: './preference-settings.component.css'
})
export class PreferenceSettingsComponent implements OnInit {

  preferences: Preferences | undefined;
  displayDynastyGenresDialog: boolean = false;
  displayBlackListTagDialog: boolean = false;

  constructor(private preferencesService: PreferencesService,
              private msgService: MessageService,
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
        this.msgService.success('Saved', 'Preferences have been updated',);
      },
      error: err => {
        this.msgService.error('Error', err.error.message);
      }
    })
  }

}
