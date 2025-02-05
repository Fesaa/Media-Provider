import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {SubscriptionService} from '../_services/subscription.service';
import {RefreshFrequency, Subscription} from "../_models/subscription";
import {Provider} from "../_models/page";
import {dropAnimation} from "../_animations/drop-animation";
import {TableModule} from "primeng/table";
import {SubscriptionExternalUrlPipe} from "../_pipes/subscription-external-url.pipe";
import {DatePipe} from "@angular/common";
import {Tooltip} from "primeng/tooltip";
import {RefreshFrequencyPipe} from "../_pipes/refresh-frequency.pipe";
import {Button} from "primeng/button";
import { DialogService } from '../_services/dialog.service';
import {MessageService} from "../_services/message.service";
import {Tag} from "primeng/tag";
import {
  SubscriptionEditDialogComponent
} from "./components/subscription-edit-dialog/subscription-edit-dialog.component";

@Component({
  selector: 'app-subscription-manager',
  imports: [
    TableModule,
    SubscriptionExternalUrlPipe,
    DatePipe,
    Tooltip,
    RefreshFrequencyPipe,
    Button,
    Tag,
    SubscriptionEditDialogComponent,
  ],
  templateUrl: './subscription-manager.component.html',
  styleUrl: './subscription-manager.component.css',
  animations: [dropAnimation]
})
export class SubscriptionManagerComponent implements OnInit {

  allowedProviders: Provider[] = [];
  subscriptions: Subscription[] = [];
  displayEditSubscription: { [key: string]: boolean } = {};

  size = 10

  constructor(private navService: NavService,
              private subscriptionService: SubscriptionService,
              private cdRef: ChangeDetectorRef,
              private dialogService: DialogService,
              private msgService: MessageService,
  ) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
    this.subscriptionService.all().subscribe(s => {
      this.subscriptions = s ?? [];
      this.cdRef.detectChanges();
    })
    this.subscriptionService.providers().subscribe(providers => {
      this.allowedProviders = providers;
      this.cdRef.detectChanges();
    })
  }

  edit(sub: Subscription) {
    this.displayEditSubscription = {} // Close others
    this.displayEditSubscription[sub.ID] = true;
  }

  update(sub: Subscription) {
    this.subscriptions = this.subscriptions.map(s => {
      if (s.ID !== sub.ID) {
        return s;
      }

      return sub;
    })
  }

  runOnce(sub: Subscription) {
    if (sub.ID == 0) {
      return
    }

    this.subscriptionService.runOnce(sub.ID).subscribe({
      next: () => {
        this.msgService.success("Success", `${sub.info.title} has been added to the download queue`)
      },
      error: (err) => {
        this.msgService.error("Failed to run once", err.error.message)
      }
    })
  }

  async delete(sub: Subscription) {
    if (!await this.dialogService.openDialog(`Are you sure you want to remove your subscription on ${sub.info.title}`)) {
      return;
    }


    this.subscriptionService.delete(sub.ID).subscribe({
      next: () => {
        this.subscriptions = this.subscriptions.filter(s => s.ID !== sub.ID)
        this.msgService.success('Deleted', `Subscription ${sub.info.title} removed.`);
      },
      error: err => {
        this.msgService.error("Failed to delete the subscription", err.error.message);
      }
    })
  }

  getSeverity(sub: Subscription): "success" | "secondary" | "info" | "warn" | "danger" | "contrast" | undefined {
    switch (sub.refreshFrequency) {
      case RefreshFrequency.Day:
        return "info"
      case RefreshFrequency.Week:
        return "warn"
      case RefreshFrequency.Month:
        return "danger"
    }
  }

}
