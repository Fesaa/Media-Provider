import {Page} from "./page";

export type Config = {
  baseUrl: string;
  cacheType: CacheType;
  redisAddr: string;
  maxConcurrentTorrents: number;
  maxConcurrentImages: number;
  disableIpv6: boolean;
  rootDir: string;
  oidc: Oidc
}

export type Oidc = {
  authority: string;
  clientId: string;
  disablePasswordLogin: boolean;
  autoLogin: boolean;
}

export enum CacheType {
  MEMORY = "MEMORY",
  REDIS = "REDIS",
}
