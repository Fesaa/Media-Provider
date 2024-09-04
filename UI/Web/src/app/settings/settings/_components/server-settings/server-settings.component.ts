import {ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {Config, LogHandler, LogLevel} from '../../../../_models/config';
import {ConfigService} from "../../../../_services/config.service";
import {FormBuilder, FormGroup, ReactiveFormsModule, Validators} from "@angular/forms";
import {FormInputComponent} from "../../../../shared/form/form-input/form-input.component";
import {FormSelectComponent} from "../../../../shared/form/form-select/form-select.component";

@Component({
  selector: 'app-server-settings',
  standalone: true,
  imports: [
    ReactiveFormsModule,
    FormInputComponent,
    FormSelectComponent
  ],
  templateUrl: './server-settings.component.html',
  styleUrl: './server-settings.component.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ServerSettingsComponent implements OnInit {

  config: Config | undefined;
  settingsForm: FormGroup | undefined;

  constructor(private configService: ConfigService,
              private fb: FormBuilder,
              private cdRef: ChangeDetectorRef
  ) {
  }

  ngOnInit(): void {
    this.configService.getConfig().subscribe(config => {
      this.config = config;
      this.buildForm();
    })
  }

  private buildForm() {
    if (!this.config) {
      return;
    }

    this.settingsForm = this.fb.group({
      port: this.fb.control(this.config.port, Validators.required),
      password: this.fb.control(this.config.password, [Validators.required, Validators.pattern('^[a-zA-Z0-9]*$')]),
      root_dir: this.fb.control(this.config.root_dir, Validators.required),
      base_url: this.fb.control(this.config.base_url, Validators.required),
      logging: this.fb.group({
        level: this.fb.control(this.config.logging.level, Validators.required),
        source: this.fb.control(this.config.logging.source, Validators.required),
        handler: this.fb.control(this.config.logging.handler, Validators.required),
        log_http: this.fb.control(this.config.logging.log_http, Validators.required),
      }),
      downloader: this.fb.group({
        max_torrents: this.fb.control<number>(this.config.downloader.max_torrents, Validators.required),
        max_mangadex_images: this.fb.control<number>(this.config.downloader.max_mangadex_images, Validators.required),
      })
    });
    this.cdRef.detectChanges();
  }

  save() {
    if (!this.settingsForm) {
      return;
    }
    this.configService.updateConfig(this.settingsForm.value).subscribe(() => {
      this.configService.getConfig().subscribe(config => {
        this.config = config;
        this.buildForm();
      });
  });
  }


  protected readonly LogHandler = LogHandler;
  protected readonly Object = Object;
  protected readonly LogLevel = LogLevel;
}
