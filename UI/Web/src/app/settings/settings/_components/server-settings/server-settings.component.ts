import {ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {CacheType, Config, LogHandler, LogLevel} from '../../../../_models/config';
import {ConfigService} from "../../../../_services/config.service";
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";
import {BoundNumberValidator, IntegerFormControl} from "../../../../_validators/BoundNumberValidator";
import {Clipboard} from "@angular/cdk/clipboard";
import {Tooltip} from "primeng/tooltip";
import {MessageService} from "../../../../_services/message.service";

@Component({
  selector: 'app-server-settings',
  imports: [
    ReactiveFormsModule,
    FormInputComponent,
    FormSelectComponent,
    Tooltip
  ],
  templateUrl: './server-settings.component.html',
  styleUrl: './server-settings.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ServerSettingsComponent implements OnInit {

  config: Config | undefined;
  settingsForm: FormGroup | undefined;

  showKey = false;
  protected readonly LogHandler = LogHandler;
  protected readonly Object = Object;
  protected readonly LogLevel = LogLevel;
  protected readonly CacheType = CacheType;

  constructor(private configService: ConfigService,
              private fb: FormBuilder,
              private cdRef: ChangeDetectorRef,
              private msgService: MessageService,
              private clipBoardService: Clipboard
  ) {
  }

  ngOnInit(): void {
    this.configService.getConfig().subscribe(config => {
      this.config = config;
      this.buildForm();
    })
  }

  hidden() {
    return "X".repeat(this.config!.api_key.length);
  }

  save() {
    if (!this.settingsForm) {
      return;
    }

    const errors = this.errors();
    if (errors > 0) {
      this.msgService.error('Cannot submit', `Found ${errors} errors in the form`);
      return;
    }

    if (!this.settingsForm.dirty) {
      this.msgService.warning('Not saving', 'No changes detected');
      return;
    }

    if (this.settingsForm.value.cache.type != CacheType.REDIS) {
      this.settingsForm.value.cache.redis = ""
    }

    this.configService.updateConfig(this.settingsForm.value).subscribe({
      next: () => {
        this.configService.getConfig().subscribe(config => {
          this.config = config;
          this.buildForm();
          this.msgService.success('Success', 'Settings saved');
        });
      },
      error: (error) => {
        this.msgService.error('Failed to save settings', error.error.message);
      }
    });
  }

  private buildForm() {
    if (!this.config) {
      return;
    }

    this.settingsForm = this.fb.group({
      password: this.fb.control(this.config.password),
      root_dir: this.fb.control(this.config.root_dir, Validators.required),
      base_url: this.fb.control(this.config.base_url),
      cache: this.fb.group({
        type: this.fb.control(this.config.cache.type, [Validators.required]),
        redis: this.fb.control(this.config.cache.redis),
      }),
      logging: this.fb.group({
        level: this.fb.control(this.config.logging.level, Validators.required),
        source: this.fb.control(this.config.logging.source, Validators.required),
        handler: this.fb.control(this.config.logging.handler, Validators.required),
        log_http: this.fb.control(this.config.logging.log_http, Validators.required),
      }),
      downloader: this.fb.group({
        max_torrents: new IntegerFormControl(this.config.downloader.max_torrents, [Validators.required, BoundNumberValidator(1, 10)]),
        max_mangadex_images: new IntegerFormControl(this.config.downloader.max_mangadex_images, [Validators.required, BoundNumberValidator(1, 5)]),
      })
    });
    this.cdRef.detectChanges();
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
