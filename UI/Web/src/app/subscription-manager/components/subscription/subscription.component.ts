import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {RefreshFrequency, Subscription} from "../../../_models/subscription";
import {SubscriptionService} from "../../../_services/subscription.service";
import {DatePipe, NgClass, NgForOf, NgIf, TitleCasePipe} from "@angular/common";
import {RefreshFrequencyPipe} from "../../../_pipes/refresh-frequency.pipe";
import {NgIcon} from "@ng-icons/core";
import {DirectoryBrowserComponent} from "../../../directory-browser/directory-browser.component";
import {SubscriptionExternalUrlPipe} from "../../../_pipes/subscription-external-url.pipe";
import {ToastrService} from "ngx-toastr";
import {DialogService} from "../../../_services/dialog.service";
import {FormsModule} from "@angular/forms";
import {Observable} from "rxjs";
import {Tooltip} from "primeng/tooltip";
import {Provider} from "../../../_models/page";

@Component({
    selector: 'app-subscription',
  imports: [
    NgClass,
    RefreshFrequencyPipe,
    NgIcon,
    SubscriptionExternalUrlPipe,
    FormsModule,
    NgForOf,
    Tooltip,
    TitleCasePipe,
    DatePipe
  ],
    templateUrl: './subscription.component.html',
    styleUrl: './subscription.component.css'
})
export class SubscriptionComponent implements OnInit {

  @Input({required: true}) subscription!: Subscription;
  @Input({required: true}) providers!: Provider[];
  @Output() onDelete = new EventEmitter<number>();
  @Output() onSave = new EventEmitter<void>();

  editMode: boolean = false;

  refreshFrequencies = Object.keys(RefreshFrequency)
    .filter((key) => isNaN(Number(key)))
    .map((key) => ({ label: key, value: RefreshFrequency[key as keyof typeof RefreshFrequency] }));

  providerOptions!: {label: string; value: Provider}[];
  constructor(private subscriptionService: SubscriptionService,
              private toastR: ToastrService,
              private dialogService: DialogService,
  ) {
  }

  toggleEditMode() {
    this.editMode = !this.editMode;
  }

  ngOnInit(): void {
    if (this.subscription.ID == 0) {
      this.editMode = true;
    }

    this.providerOptions = this.providers.map((provider) => {
      const key = Provider[provider];
      return { label: key, value: provider };
    });
  }

  runOnce() {
    if (this.subscription.ID == 0) {
      return
    }

    this.subscriptionService.runOnce(this.subscription.ID).subscribe({
      next: () => {
        this.toastR.success("Success")
      },
      error: (err) => {
        this.toastR.error("Failed to run once", err.message)
      }
      })
  }

  async delete() {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove your subscription on ${this.subscription.info.title}`)) {
      return;
    }

    if (this.subscription.ID == 0) {
      this.onDelete.emit(this.subscription.ID);
      return;
    }


    this.subscriptionService.delete(this.subscription.ID).subscribe({
      next: () => {
        this.toastR.success('Subscription deleted');
      },
      error: err => {
        this.toastR.error(err.message);
      },
      complete: () => {
        this.onDelete.emit(this.subscription.ID);
      }
    })
  }

  async openDirSelector() {
    const dir = await this.dialogService.openDirBrowser("");
    if (dir == undefined) {
      return;
    }

    this.subscription.info.baseDir = dir;
  }

  private valid(): boolean {
    if (this.subscription.contentId == "") {
      this.toastR.error("Subscription id must be set");
      return false;
    }

    if (this.subscription.info.title == "") {
      this.toastR.error("Subscription title must be set");
      return false;
    }

    if (this.subscription.info.baseDir == "") {
      this.toastR.error("Subscription baseDir must be set");
      return false;
    }

    return true;
  }

  saveSubscription() {
    if (!this.valid()) {
      return;
    }

    // monkey patch
    this.subscription.refreshFrequency = Number(this.subscription.refreshFrequency);
    this.subscription.provider = Number(this.subscription.provider);

    let obs: Observable<any>;
    if (this.subscription.ID == 0) {
      obs = this.subscriptionService.new(this.subscription);
    } else {
      obs = this.subscriptionService.update(this.subscription);
    }

    obs.subscribe({
      next: () => {
        this.toastR.success('Subscription updated the subscription');
      },
      error: err => {
        this.toastR.error(err.message);
      },
      complete: () => {
        this.onSave.emit();
      }
    })
  }
}
