import {Provider} from "./page";
import {DownloadRequestMetadata} from "./search";

export type Subscription = {
  ID: number;
  provider: Provider;
  contentId: string;
  refreshFrequency: RefreshFrequency;
  info: SubscriptionInfo;
  metadata: DownloadRequestMetadata;
}

export type SubscriptionInfo = {
  title: string;
  description?: string;
  baseDir: string;
  lastCheck: Date;
  lastCheckSuccess: boolean;
  nextExecution: Date;
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
