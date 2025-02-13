import {Component, OnInit} from '@angular/core';
import {PreferencesService} from '../../../../_services/preferences.service';
import {FormBuilder, FormGroup, FormsModule, ReactiveFormsModule, Validators} from "@angular/forms";
import {Preferences} from "../../../../_models/preferences";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from '../../../../_services/message.service';
import {ToggleSwitch} from "primeng/toggleswitch";
import {InputText} from "primeng/inputtext";
import {InputNumber} from "primeng/inputnumber";
import {Button} from "primeng/button";

@Component({
  selector: 'app-preference-settings',
  standalone: true,
  imports: [
    FormsModule,
    ReactiveFormsModule,
    Tooltip,
    ToggleSwitch,
    InputNumber,
    Button
  ],
  templateUrl: './preference-settings.component.html',
  styleUrl: './preference-settings.component.css'
})
export class PreferenceSettingsComponent implements OnInit {

  preferences: Preferences | undefined;

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
