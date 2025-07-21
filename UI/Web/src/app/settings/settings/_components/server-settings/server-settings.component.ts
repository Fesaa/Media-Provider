import {ChangeDetectionStrategy, ChangeDetectorRef, Component, effect, inject, OnInit} from '@angular/core';
import {CacheType, CacheTypes, Config} from '../../../../_models/config';
import {FormBuilder, FormControl, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";
import {ToastService} from "../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {Button} from "primeng/button";
import {SettingsService} from "../../../../_services/settings.service";

@Component({
  selector: 'app-server-settings',
  imports: [
    ReactiveFormsModule,
    FormInputComponent,
    TranslocoDirective,
    Button,
    FormSelectComponent
  ],
  templateUrl: './server-settings.component.html',
  styleUrl: './server-settings.component.scss',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServerSettingsComponent {

  private readonly settingsService = inject(SettingsService);

  config = this.settingsService.config;

  settingsForm: FormGroup | undefined;

  constructor(private fb: FormBuilder,
              protected cdRef: ChangeDetectorRef,
              private toastService: ToastService,
  ) {

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
        oidc: this.fb.group({
          authority: this.fb.control(config.oidc.authority),
          clientId: this.fb.control(config.oidc.clientId),
          disablePasswordLogin: this.fb.control(config.oidc.disablePasswordLogin),
          autoLogin: this.fb.control(config.oidc.autoLogin),
        }),
      });
      this.cdRef.detectChanges();
    });
  }

  getFormControl(path: string): FormControl | null {
    if (!this.settingsForm) return null;

    const control = this.settingsForm.get(path);
    return control instanceof FormControl ? control : null;
  }

  save() {
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

    const dto: Config = this.settingsForm.getRawValue();
    dto.maxConcurrentImages = parseInt(String(dto.maxConcurrentImages))
    dto.maxConcurrentTorrents = parseInt(String(dto.maxConcurrentTorrents))

    if (dto.cacheType != CacheType.REDIS) {
      dto.redisAddr = ""
    }

    this.settingsService.updateConfig(dto).subscribe({
      next: () => {
        this.toastService.successLoco("settings.server.toasts.save.success");
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
}
