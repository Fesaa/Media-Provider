import {Component, EventEmitter, Input, Output} from '@angular/core';
import {RefreshFrequencies, RefreshFrequency, Subscription} from "../../../_models/subscription";
import {SubscriptionService} from "../../../_services/subscription.service";
import {Dialog} from "primeng/dialog";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule} from "@angular/forms";
import {IconField} from "primeng/iconfield";
import {InputIcon} from "primeng/inputicon";
import {SubscriptionExternalUrlPipe} from "../../../_pipes/subscription-external-url.pipe";
import {Select} from "primeng/select";
import {
  DownloadMetadata,
  DownloadMetadataDefinition,
  DownloadMetadataFormType,
  Provider,
  Providers
} from "../../../_models/page";
import {ToastService} from "../../../_services/toast.service";
import {Button} from "primeng/button";
import {DirectorySelectorComponent} from "../../../shared/_component/directory-selector/directory-selector.component";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";
import {PageService} from "../../../_services/page.service";
import {
  ProviderMetadataOptionsComponent
} from "../../../shared/_component/provider-metadata-options/provider-metadata-options.component";
import {Tab, TabList, TabPanel, TabPanels, Tabs} from "primeng/tabs";

enum TabId {
  General = "general",
  Metadata = "metadata",
}

@Component({
  selector: 'app-subscription-edit-dialog',
  imports: [
    Dialog,
    FloatLabel,
    InputText,
    FormsModule,
    IconField,
    InputIcon,
    Select,
    Button,
    DirectorySelectorComponent,
    TranslocoDirective,
    TitleCasePipe,
    ProviderMetadataOptionsComponent,
    Tabs,
    TabList,
    Tab,
    TabPanels,
    TabPanel,
  ],
  templateUrl: './subscription-edit-dialog.component.html',
  styleUrl: './subscription-edit-dialog.component.css'
})
export class SubscriptionEditDialogComponent {

  @Input({required: true}) visible!: boolean;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();
  @Input({required: true}) sub!: Subscription;
  @Output() update: EventEmitter<Subscription> = new EventEmitter<Subscription>();
  @Input({required: true}) providers!: Provider[];

  metadata!: DownloadMetadata | undefined;

  copy: Subscription = {
    ID: 0,
    contentId: '',
    provider: Provider.NYAA,
    refreshFrequency: RefreshFrequency.Day,
    info: {
      title: '',
      lastCheckSuccess: true,
      lastCheck: new Date(),
      nextExecution: new Date(),
      description: '',
      baseDir: ''
    },
    metadata: {
      extra: {},
      startImmediately: true,
    }
  };

  dirBrowser = false;
  filteredProviders!: {label: string, value: Provider}[];

  constructor(
    private subscriptionService: SubscriptionService,
    private externalUrlPipe: SubscriptionExternalUrlPipe,
    private toastService: ToastService,
    private pageService: PageService,
  ) {
  }

  refresh() {

    if (!this.metadata) {
      this.pageService.metadata(this.sub.provider).subscribe(meta => {
        this.metadata = meta;
      })
    }

    this.filteredProviders = Providers.filter(p => this.providers.includes(p.value))
    this.copy = {
      ID: this.sub.ID,
      provider: this.sub.provider,
      refreshFrequency: this.sub.refreshFrequency,
      contentId: this.sub.contentId,
      info: {
        title: this.sub.info.title,
        baseDir: this.sub.info.baseDir,
        description: this.sub.info.description,
        lastCheck: this.sub.info.lastCheck,
        lastCheckSuccess: this.sub.info.lastCheckSuccess,
        nextExecution: this.sub.info.nextExecution,
      },
      metadata: {
        startImmediately: this.sub.metadata.startImmediately,
        extra: this.sub.metadata.extra || {},
      },
    }
  }

  close() {
    this.visibleChange.emit(false);
  }

  updateDir(dir: string | undefined) {
    if (!dir) {
      return;
    }

    this.copy.info.baseDir = dir;
  }

  edit() {
    this.subscriptionService.update(this.copy).subscribe({
      next: () => {
        this.toastService.successLoco("subscriptions.toasts.update.success", {name: this.copy.info.title});
        this.sub = this.copy
        this.update.emit(this.copy)
      },
      error: err => {
        this.toastService.errorLoco("subscriptions.toasts.update.error", {name: this.copy.info.title}, {msg: err.error.message});
      }
    }).add(() => this.close())
  }

  openExternal() {
    window.open(this.externalUrlPipe.transform(this.sub.contentId, this.sub.provider), '_blank');
  }

  protected readonly RefreshFrequencies = RefreshFrequencies;
  protected readonly Tab = Tab;
  protected readonly TabId = TabId;
}
