import {Provider} from "./page";

export type SearchRequest = {
  provider: Provider;
  query: string;
  modifiers?: {[key: string]: string[]};
}

export type DownloadRequest = {
  provider: Provider;
  id: string;
  dir: string;
  title: string;
}

export type StopRequest = {
  provider: Provider;
  id: string;
  delete: boolean;
}
