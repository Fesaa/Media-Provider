import {Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {SubscriptionService} from '../_services/subscription.service';
import {RefreshFrequency, Subscription} from "../_models/subscription";
import {Provider} from "../_models/page";
import {SubscriptionComponent} from "./components/subscription/subscription.component";
import {dropAnimation} from "../_animations/drop-animation";

@Component({
    selector: 'app-subscription-manager',
    imports: [
        SubscriptionComponent,
    ],
    templateUrl: './subscription-manager.component.html',
    styleUrl: './subscription-manager.component.css',
    animations: [dropAnimation]
})
export class SubscriptionManagerComponent implements OnInit {

  allowedProviders: Provider[] = [];
  subscriptions: Subscription[] = [];
  show: boolean = true;

  constructor(private navService: NavService,
              private subscriptionService: SubscriptionService,
              ) {
  }

  ngOnInit(): void {
    this.navService.setNavVisibility(true)
    this.subscriptionService.all().subscribe(s => {
      this.subscriptions = s ?? [];
    })
    this.subscriptionService.providers().subscribe(providers => {
      this.allowedProviders = providers;
    })
  }

  remove(id: number) {
    this.subscriptions = this.subscriptions.filter(s => s.ID !== id);
  }

  reload() {
    this.subscriptionService.all().subscribe(s => {
      this.subscriptions = s ?? [];
    })
  }

  newSubscriptions() {
    return this.subscriptions.filter(s => s.ID === 0);
  }

  toggle() {
    this.show = !this.show;
  }


  addNew() {
    this.subscriptions.push({
      ID: 0,
      info: {
        title: "",
        baseDir: "",
        lastCheck: new Date(),
        lastCheckSuccess: true,
      },
      contentId: "",
      provider: Provider.MANGADEX,
      refreshFrequency: RefreshFrequency.Week,
    })
  }
}
