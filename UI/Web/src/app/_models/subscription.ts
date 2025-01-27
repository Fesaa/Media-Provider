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

export const RefreshFrequencies = [
  {label: "Day", value: RefreshFrequency.Day},
  {label: "Week", value: RefreshFrequency.Week},
  {label: "Month", value: RefreshFrequency.Month},
];
