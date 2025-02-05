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
import {Provider, Providers} from "../../../_models/page";
import {MessageService} from "../../../_services/message.service";
import {Button} from "primeng/button";
import {DirectorySelectorComponent} from "../../../shared/_component/directory-selector/directory-selector.component";

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

  copy: Subscription = {
    ID: 0,
    contentId: '',
    provider: Provider.NYAA,
    refreshFrequency: RefreshFrequency.Day,
    info: {
      title: '',
      lastCheckSuccess: true,
      lastCheck: new Date(),
      description: '',
      baseDir: ''
    }
  };

  dirBrowser = false;
  filteredProviders!: {label: string, value: Provider}[];

  constructor(
    private subscriptionService: SubscriptionService,
    private externalUrlPipe: SubscriptionExternalUrlPipe,
    private msgService: MessageService,
  ) {
  }

  refresh() {
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
      }
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
        this.msgService.success("Updated", `${this.copy.info.title} has been updated`)
        this.sub = this.copy
        this.update.emit(this.copy)
      },
      error: err => {
        this.msgService.error("Failed", `An error occurred while trying to update ${this.copy.info.title}:\n ${err.error.message}`)
      }
    }).add(() => this.close())
  }

  openExternal() {
    window.open(this.externalUrlPipe.transform(this.sub.contentId, this.sub.provider), '_blank');
  }

  protected readonly RefreshFrequencies = RefreshFrequencies;
}
