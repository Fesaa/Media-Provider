import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {SearchInfo} from "../../../_models/Info";
import {RefreshFrequencies, RefreshFrequency, Subscription} from "../../../_models/subscription";
import {SubscriptionService} from "../../../_services/subscription.service";
import {FloatLabel} from "primeng/floatlabel";
import {InputText} from "primeng/inputtext";
import {FormsModule, ReactiveFormsModule} from "@angular/forms";
import {Button} from "primeng/button";
import {Select} from "primeng/select";
import {ToastService} from "../../../_services/toast.service";
import {TranslocoDirective} from "@jsverse/transloco";
import {TitleCasePipe} from "@angular/common";

@Component({
  selector: 'app-subscription-dialog',
  imports: [
    FloatLabel,
    InputText,
    ReactiveFormsModule,
    FormsModule,
    Button,
    Select,
    TranslocoDirective,
    TitleCasePipe
  ],
  templateUrl: './subscription-dialog.component.html',
  styleUrl: './subscription-dialog.component.css'
})
export class SubscriptionDialogComponent implements OnInit {

  @Input({required: true}) visible!: boolean;
  @Output() visibleChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  @Input({required: true}) downloadDir!: string;
  @Input({required: true}) searchResult!: SearchInfo;

  subscription!: Subscription;
  protected readonly RefreshFrequencies = RefreshFrequencies;

  constructor(
    private subscriptionService: SubscriptionService,
    private toastService: ToastService,
  ) {
  }

  ngOnInit(): void {
    this.subscription = {
      ID: 0,
      contentId: this.searchResult.InfoHash,
      provider: this.searchResult.Provider,
      info: {
        title: this.searchResult.Name,
        baseDir: this.downloadDir,
        lastCheckSuccess: true,
        lastCheck: new Date()
      },
      refreshFrequency: RefreshFrequency.Week
    };
  }

  close(): void {
    this.visibleChange.emit(false);
  }

  subscribe() {
    this.subscriptionService.new(this.subscription).subscribe({
      next: sub => {
        this.toastService.successLoco("page.subscription-dialog.toasts.success", {}, {name: sub.info.title})
      },
      error: err => {
        this.toastService.genericError(err.error.message);
      }
    }).add(() => {
      this.visibleChange.emit(false);
    })
  }
}
