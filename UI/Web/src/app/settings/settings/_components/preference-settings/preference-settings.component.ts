import {Component, DestroyRef, inject, OnInit, signal} from '@angular/core';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {PreferencesService} from '../../../../_services/preferences.service';
import {CoverFallbackMethods, normalize, Preferences, Tag} from '../../../../_models/preferences';
import {ToastService} from '../../../../_services/toast.service';
import {TranslocoDirective} from '@jsverse/transloco';
import {debounceTime, distinctUntilChanged, map, switchMap} from 'rxjs';
import {SettingsItemComponent} from "../../../../shared/form/settings-item/settings-item.component";
import {TagBadgeComponent} from "../../../../shared/_component/tag-badge/tag-badge.component";
import {takeUntilDestroyed} from "@angular/core/rxjs-interop";

@Component({
  selector: 'app-preference-settings',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    TranslocoDirective,
    SettingsItemComponent,
    TagBadgeComponent
  ],
  templateUrl: './preference-settings.component.html',
  styleUrl: './preference-settings.component.scss'
})
export class PreferenceSettingsComponent implements OnInit {

  private readonly destroyRef = inject(DestroyRef);
  private readonly preferencesService = inject(PreferencesService);
  private readonly toastService = inject(ToastService);
  private readonly fb = inject(FormBuilder);

  preferences = signal<Preferences | undefined>(undefined);

  preferencesForm!: FormGroup;

  protected readonly CoverFallbackMethods = CoverFallbackMethods;

  ngOnInit(): void {
    this.preferencesService.get().subscribe((preferences: Preferences) => {
      this.preferences.set(preferences);

      this.preferencesForm = this.fb.group({
        subscriptionRefreshHour: [preferences.subscriptionRefreshHour, [Validators.required, Validators.min(0), Validators.max(23)]],
        logEmptyDownloads: [preferences.logEmptyDownloads],
        convertToWebp: [preferences.convertToWebp],
        coverFallbackMethod: [preferences.coverFallbackMethod],
        blackList: new FormControl(preferences.blackListedTags.map(t => t.name).join(',')),
        whiteList: new FormControl(preferences.whiteListedTags.map(t => t.name).join(',')),
        tagToGenre: new FormControl(preferences.dynastyGenreTags.map(t => t.name).join(',')),
      });

      this.preferencesForm.valueChanges
        .pipe(
          takeUntilDestroyed(this.destroyRef),
          debounceTime(300),
          distinctUntilChanged(),
          map(() => this.packData()),
          switchMap(data => this.preferencesService.save(data)),
        )
        .subscribe({
          error: err => this.toastService.genericError(err.error.message)
        });
    });
  }

  packData() {
    const preferences = this.preferences();
    const formValue = this.preferencesForm.value;

    return {
      ...preferences,
      ...formValue,
      coverFallbackMethod: parseInt(formValue.coverFallbackMethod),
      blackListedTags: (formValue.blackList as string)
        .split(',').map((item: string) => item.trim())
        .filter((t: string) => t.length > 0)
        .map(this.toTag),
      whiteListedTags: (formValue.whiteList as string)
        .split(',').map((item: string) => item.trim())
        .filter((t: string) => t.length > 0)
        .map(this.toTag),
      dynastyGenreTags: (formValue.tagToGenre as string)
        .split(',').map((item: string) => item.trim())
        .filter((t: string) => t.length > 0)
        .map(this.toTag),
    };
  }

  toTag(s: string): Tag {
    return {
      name: s,
      normalizedName: normalize(s),
    }
  }

  breakString(s: string) {
    if (s) {
      return s.split(',').filter(s => s.length > 0);
    }

    return [];
  }
}
