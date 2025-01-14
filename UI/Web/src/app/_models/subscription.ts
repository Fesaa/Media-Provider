import {Provider} from "./page";

export type Subscription = {
  ID: number;
  provider: Provider;
  contentId: string;
  refreshFrequency: RefreshFrequency;
  info: SubscriptionInfo
}

export type SubscriptionInfo = {
  title: string;
  description?: string;
  baseDir: string;
  lastCheck: Date;
  lastCheckSuccess: boolean;
}

export enum RefreshFrequency {
  Day = 2,
  Week,
  Month,
}
