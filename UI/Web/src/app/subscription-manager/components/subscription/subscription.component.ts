import {Component, EventEmitter, Input, OnInit, Output} from '@angular/core';
import {RefreshFrequency, Subscription} from "../../../_models/subscription";
import {SubscriptionService} from "../../../_services/subscription.service";
import {DatePipe, NgClass, NgForOf, NgIf} from "@angular/common";
import {RefreshFrequencyPipe} from "../../../_pipes/refresh-frequency.pipe";
import {NgIcon} from "@ng-icons/core";
import {DirectoryBrowserComponent} from "../../../directory-browser/directory-browser.component";
import {SubscriptionExternalUrlPipe} from "../../../_pipes/subscription-external-url.pipe";
import {ToastrService} from "ngx-toastr";
import {DialogService} from "../../../_services/dialog.service";
import {FormsModule} from "@angular/forms";
import {Observable} from "rxjs";

@Component({
    selector: 'app-subscription',
    imports: [
        NgClass,
        RefreshFrequencyPipe,
        NgIcon,
        SubscriptionExternalUrlPipe,
        FormsModule,
        NgForOf
    ],
    templateUrl: './subscription.component.html',
    styleUrl: './subscription.component.css'
})
export class SubscriptionComponent implements OnInit {

  @Input({required: true}) subscription!: Subscription;
  @Output() onDelete = new EventEmitter<number>();
  @Output() onSave = new EventEmitter<void>();

  editMode: boolean = false;

  refreshFrequencies = Object.keys(RefreshFrequency)
    .filter((key) => isNaN(Number(key))) // Exclude numeric keys
    .map((key) => ({ label: key, value: RefreshFrequency[key as keyof typeof RefreshFrequency] }));

  constructor(private subscriptionService: SubscriptionService,
              private toastR: ToastrService,
              private dialogService: DialogService,
  ) {
  }

  toggleEditMode() {
    this.editMode = !this.editMode;
  }

  ngOnInit(): void {
    if (this.subscription.id == -1) {
      this.editMode = true;
    }
  }

  async delete() {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove your subscription on ${this.subscription.info.title}`)) {
      return;
    }

    if (this.subscription.id == -1) {
      this.onDelete.emit(this.subscription.id);
      return;
    }


    this.subscriptionService.delete(this.subscription.id).subscribe({
      next: () => {
        this.toastR.success('Subscription deleted');
      },
      error: err => {
        this.toastR.error(err.message);
      },
      complete: () => {
        this.onDelete.emit(this.subscription.id);
      }
    })
  }

  protected readonly RefreshFrequencyPipe = RefreshFrequencyPipe;
  protected readonly RefreshFrequency = RefreshFrequency;
  protected readonly Object = Object;

  async openDirSelector() {
    const dir = await this.dialogService.openDirBrowser("");
    if (dir == undefined) {
      return;
    }

    this.subscription.info.baseDir = dir;
  }

  saveSubscription() {
    let obs: Observable<Subscription>;
    if (this.subscription.id == -1) {
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
