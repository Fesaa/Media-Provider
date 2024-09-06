import {Provider} from "./page";


export type QueueStat = {
  provider: Provider;
  id: string;
  name: string;
}

export type InfoStat = {
  provider: Provider;
  id: string;
  name: string;
  size: string;
  downloading: boolean;
  progress: number;
  speed_type: SpeedType;
  speed: SpeedData;
  download_dir: string;
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
