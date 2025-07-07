import {ChangeDetectionStrategy, ChangeDetectorRef, Component, effect, inject, OnInit} from '@angular/core';
import {CacheType} from '../../../../_models/config';
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";
import {Tooltip} from "primeng/tooltip";
import {ToastService} from "../../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {Button} from "primeng/button";
import {SettingsService} from "../../../../_services/settings.service";

@Component({
  selector: 'app-server-settings',
  imports: [
    ReactiveFormsModule,
    FormInputComponent,
    FormSelectComponent,
    Tooltip,
    TranslocoDirective,
    Button
  ],
  templateUrl: './server-settings.component.html',
  styleUrl: './server-settings.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServerSettingsComponent {

  private readonly settingsService = inject(SettingsService);

  config = this.settingsService.config;

  settingsForm: FormGroup | undefined;

  protected readonly Object = Object;
  protected readonly CacheType = CacheType;

  constructor(private fb: FormBuilder,
              protected cdRef: ChangeDetectorRef,
              private toastService: ToastService,
  ) {

    effect(() => {
      const config = this.settingsService.config();
      if (config == undefined) return

      this.settingsForm = this.fb.group({
        rootDir: this.fb.control(config.rootDir, Validators.required),
        baseUrl: this.fb.control(config.baseUrl),
        cache: this.fb.group({
          cacheType: this.fb.control(config.cacheType),
          redisAddr: this.fb.control(config.redisAddr),
        }),
        downloader: this.fb.group({
          maxConcurrentImages: this.fb.control(config.maxConcurrentImages),
          maxConcurrentTorrents: this.fb.control(config.maxConcurrentTorrents),
        }),
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

    if (this.settingsForm.value.cache.type != CacheType.REDIS) {
      this.settingsForm.value.cache.redis = ""
    }

    this.settingsService.updateConfig(this.settingsForm.value).subscribe({
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
}
