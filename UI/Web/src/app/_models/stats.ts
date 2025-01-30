import {Provider} from "./page";

export type InfoStat = {
  provider: Provider;
  id: string;
  contentState: ContentState;
  name: string;
  ref_url: string;
  size: string;
  downloading: boolean;
  progress: number;
  estimated?: number;
  speed_type: SpeedType;
  speed: number;
  download_dir: string;
}

export enum ContentState {
  Downloading = 0,
  Ready = 1,
  Waiting = 2,
  Loading = 3,
  Queued = 4,
}

export enum SpeedType {
  BYTES,
  VOLUMES,
  IMAGES,
}

export type StatsResponse = {
  running: InfoStat[];
}
