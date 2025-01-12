import {Component, OnInit} from '@angular/core';
import { PreferencesService } from '../../../../_services/preferences.service';
import {FormBuilder, FormGroup, FormsModule, ReactiveFormsModule, Validators} from "@angular/forms";
import {ToastrService} from "ngx-toastr";
import {Preferences} from "../../../../_models/preferences";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {Tooltip} from "primeng/tooltip";

@Component({
  selector: 'app-preference-settings',
  standalone: true,
  imports: [
    FormsModule,
    ReactiveFormsModule,
    FormInputComponent,
    Tooltip
  ],
  templateUrl: './preference-settings.component.html',
  styleUrl: './preference-settings.component.css'
})
export class PreferenceSettingsComponent implements OnInit {

  preferences: Preferences | undefined;
  preferencesForm: FormGroup | undefined;

  constructor(private preferencesService: PreferencesService,
              private fb: FormBuilder,
              private toastr: ToastrService
              ) {
  }

  ngOnInit(): void {
    this.preferencesService.get().subscribe(preferences => {
      this.preferences = preferences;
      this.buildForm()
    })
  }

  save() {
    if (!this.preferencesForm) {
      return;
    }
    if (!this.preferencesForm.dirty) {
      this.toastr.warning('No changes detected', 'Not saving');
      return;
    }

    if (!this.preferencesForm.valid) {
      this.toastr.warning('Please fill out all required fields correctly', 'Not saving');
      return;
    }

    const pref: Preferences = {
      ...this.preferences,
      subscriptionRefreshHour: +this.preferencesForm.get("subscriptionRefreshHour")?.value!
    }

    this.preferencesService.save(pref).subscribe({
      next: () => {
        this.preferences = this.preferencesForm!.value as Preferences;
        this.toastr.success('Saved changes detected', 'Saved');
      },
      error: err => {
        this.toastr.error(err.error.message, 'Error');
      }
    })
  }

  private buildForm() {
    if (!this.preferences) {
      return;
    }

    this.preferencesForm = this.fb.group({
      subscriptionRefreshHour: this.fb.control<number>(this.preferences.subscriptionRefreshHour,
        [Validators.required, Validators.min(0), Validators.max(23)])
    })
  }

}
