import {Provider} from "./page";


export type QueueStat = {
  provider: Provider;
  id: string;
  name: string;
}

export type InfoStat = {
  provider: Provider;
  id: string;
  contentStatus: ContentStatus;
  name: string;
  ref_url: string;
  size: string;
  downloading: boolean;
  progress: number;
  estimated?: number;
  speed_type: SpeedType;
  speed: SpeedData;
  download_dir: string;
}

export function ContentStatusWeight(contentStatus: ContentStatus): number {
  switch (contentStatus) {
    case ContentStatus.Downloading:
      return 100;
    case ContentStatus.Loading:
      return 90;
    case ContentStatus.Waiting:
      return 80;
    case ContentStatus.Queued:
      return 0;
  }
  return -1;
}

export enum ContentStatus {
  Downloading = "downloading",
  Waiting = "waiting",
  Loading = "loading",
  Queued = "queued",
}

export enum SpeedType {
  BYTES,
  VOLUMES,
  IMAGES,
}

export type SpeedData = {
  time: number;
  speed: number;
}

export type StatsResponse = {
  running: InfoStat[];
  queued: QueueStat[];
}
