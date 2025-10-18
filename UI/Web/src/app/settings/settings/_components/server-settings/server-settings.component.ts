import {ChangeDetectionStrategy, ChangeDetectorRef, Component, DestroyRef, effect, inject} from '@angular/core';
import {CacheType, CacheTypes, Config} from '../../../../_models/config';
import {
  FormBuilder,
  FormControl,
  FormGroup,
  NonNullableFormBuilder,
  ReactiveFormsModule,
  Validators
} from "@angular/forms";
import {ToastService} from "../../../../_services/toast.service";
import {translate, TranslocoDirective} from "@jsverse/transloco";
import {SettingsService} from "../../../../_services/settings.service";
import {SettingsItemComponent} from "../../../../shared/form/settings-item/settings-item.component";
import {SettingsSwitchComponent} from "../../../../shared/form/settings-switch/settings-switch.component";
import {DefaultValuePipe} from "../../../../_pipes/default-value.pipe";
import {takeUntilDestroyed} from "@angular/core/rxjs-interop";
import {debounceTime, distinctUntilChanged, filter, map, switchMap, tap} from "rxjs";

@Component({
  selector: 'app-server-settings',
  imports: [
    ReactiveFormsModule,
    TranslocoDirective,
    SettingsItemComponent,
    SettingsSwitchComponent,
    DefaultValuePipe
  ],
  templateUrl: './server-settings.component.html',
  styleUrl: './server-settings.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServerSettingsComponent {

  private readonly settingsService = inject(SettingsService);
  private readonly fb = inject(NonNullableFormBuilder);
  protected readonly cdRef = inject(ChangeDetectorRef);
  private readonly toastService = inject(ToastService);
  private readonly destroyRef = inject(DestroyRef);

  config = this.settingsService.config;

  settingsForm: FormGroup<{
    rootDir: FormControl<string>
    baseUrl: FormControl<string>
    cacheType: FormControl<CacheType>
    redisAddr: FormControl<string>
    maxConcurrentImages: FormControl<number>
    maxConcurrentTorrents: FormControl<number>
    disableIpv6: FormControl<boolean>
    oidc: FormGroup<{
      authority: FormControl<string>;
      clientId: FormControl<string>;
      clientSecret: FormControl<string>;
      disablePasswordLogin: FormControl<boolean>;
      autoLogin:FormControl <boolean>;
    }>
    subscriptionRefreshHour: FormControl<number>;
  }> | undefined;

  constructor() {
    effect(() => {
      const config = this.settingsService.config();
      if (config == undefined) return

      this.settingsForm = this.fb.group({
        rootDir: this.fb.control(config.rootDir, [Validators.required]),
        baseUrl: this.fb.control(config.baseUrl),
        cacheType: this.fb.control(config.cacheType, [Validators.required]),
        redisAddr: this.fb.control(config.redisAddr),
        maxConcurrentImages: this.fb.control(config.maxConcurrentImages, [Validators.required, Validators.min(1), Validators.max(5)]),
        maxConcurrentTorrents: this.fb.control(config.maxConcurrentTorrents, [Validators.required, Validators.min(1), Validators.max(10)]),
        disableIpv6: this.fb.control(config.disableIpv6),
        oidc: this.fb.group({
          authority: this.fb.control(config.oidc.authority),
          clientId: this.fb.control(config.oidc.clientId),
          disablePasswordLogin: this.fb.control(config.oidc.disablePasswordLogin),
          autoLogin: this.fb.control(config.oidc.autoLogin),
          clientSecret: this.fb.control(config.oidc.clientSecret),
        }),
        subscriptionRefreshHour: this.fb.control(config.subscriptionRefreshHour),
      });
      this.cdRef.detectChanges();

      this.settingsForm.valueChanges.pipe(
        takeUntilDestroyed(this.destroyRef),
        distinctUntilChanged(),
        debounceTime(400),
        map(() => this.settingsForm?.getRawValue()),
        filter((dto) => { // Don't auto save when critical OIDC info changes
          if (!dto) return false;

          const oidc = this.config()?.oidc;
          if (!oidc) return false;

          return dto.oidc.authority === oidc.authority && dto.oidc.clientSecret === oidc.clientSecret;
        }),
        tap(() => this.save(false))
      ).subscribe();
    });
  }

  getFormControl(path: string): FormControl | null {
    if (!this.settingsForm) return null;

    const control = this.settingsForm.get(path);
    return control instanceof FormControl ? control : null;
  }

  save(toastOnSuccess: boolean = true) {
    if (!this.settingsForm) {
      return;
    }

    const errors = this.errors();
    if (errors > 0) {
      this.toastService.errorLoco("settings.server.toasts.cant-submit", {}, {amount: errors});
      return;
    }

    if (!this.settingsForm.dirty) {
      this.toastService.warningLoco("shared.toasts.no-changes")
      return;
    }

    const metadata = this.config()?.metadata;
    if (!metadata) return;

    const dto: Config = {
      metadata,
      ...this.settingsForm.getRawValue(),
    };
    dto.maxConcurrentImages = parseInt(String(dto.maxConcurrentImages))
    dto.maxConcurrentTorrents = parseInt(String(dto.maxConcurrentTorrents))

    if (dto.cacheType != CacheType.REDIS) {
      dto.redisAddr = ""
    }

    this.settingsService.updateConfig(dto).subscribe({
      next: () => {
        if (toastOnSuccess) {
          this.toastService.successLoco("settings.server.toasts.save.success");
        }
      },
      error: (error) => {
        console.error(error);
        this.toastService.genericError(error.error.message);
      }
    });
  }

  private errors() {
    let count = 0;
    Object.keys(this.settingsForm!.controls).forEach(key => {
      const controlErrors = this.settingsForm!.get(key)?.errors;
      if (controlErrors) {
        console.log(controlErrors);
        count += Object.keys(controlErrors).length;
      }
    });

    return count
  }

  protected readonly CacheType = CacheType;
  protected readonly CacheTypes = CacheTypes;
  protected readonly translate = translate;
}
