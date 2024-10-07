import {Page} from "./page";

export type Config = {
  sync_id: number;
  password: string;
  root_dir: string;
  base_url: string;

  cache: Cache;
  logging: Logging;
  downloader: Downloader;

  pages: Page[];
}

export type Cache = {
  type: CacheType;
  redis?: string
}

export enum CacheType {
  MEMORY = "MEMORY",
  REDIS = "REDIS",
}

export type Logging = {
  level: LogLevel;
  source: boolean;
  handler: LogHandler;
  log_http: boolean;
}

export type Downloader = {
  max_torrents: number;
  max_mangadex_images: number;
}

export enum LogHandler {
  JSON = 'JSON',
  TEXT = 'TEXT',
}

export enum LogLevel {
  TRACE = 'DEBUG-4',
  DEBUG = 'DEBUG',
  INFO = 'INFO',
  WARN = 'WARN',
  ERROR = 'ERROR',
  FATAL = 'ERROR+4',
}

export type MovePageRequest = {
  oldIndex: number;
  newIndex: number;
}
