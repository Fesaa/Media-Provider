import {ChangeDetectorRef, Component, OnInit} from '@angular/core';
import {NavService} from "../_services/nav.service";
import {SubscriptionService} from '../_services/subscription.service';
import {RefreshFrequency, Subscription} from "../_models/subscription";
import {Provider} from "../_models/page";
import {SubscriptionComponent} from "./components/subscription/subscription.component";
import {dropAnimation} from "../_animations/drop-animation";
import {PaginatorComponent} from "../paginator/paginator.component";

@Component({
    selector: 'app-subscription-manager',
  imports: [
    SubscriptionComponent,
    PaginatorComponent,
  ],
    templateUrl: './subscription-manager.component.html',
    styleUrl: './subscription-manager.component.css',
    animations: [dropAnimation]
})
export class SubscriptionManagerComponent implements OnInit {

  allowedProviders: Provider[] = [];
  subscriptions: Subscription[] = [];
  newSubscription: Subscription | null = null;
  show: boolean = true;

  currentPage: number = 1;
  pageSize: number = 5;

  constructor(private navService: NavService,
              private subscriptionService: SubscriptionService,
              private cdRef: ChangeDetectorRef,
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
    })
  }

  onPageChange(page: number) {
    this.currentPage = page;
  }

  remove(id: number) {
    if (id === 0) {
      this.newSubscription = null;
    }

    this.subscriptions = this.subscriptions.filter(s => s.ID !== id);
  }

  reload() {
    this.subscriptionService.all().subscribe(s => {
      this.subscriptions = s ?? [];
    })
  }

  toDisplay() {
    if (!this.show) {
      return []
    }
    return this.subscriptions.slice((this.currentPage - 1) * this.pageSize, this.currentPage * this.pageSize)
  }

  toggle() {
    this.show = !this.show;
  }


  addNew() {
    this.newSubscription = {
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
    }
  }

  protected readonly Math = Math;
}
