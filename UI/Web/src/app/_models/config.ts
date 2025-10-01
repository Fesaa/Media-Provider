export type Config = {
  baseUrl: string;
  cacheType: CacheType;
  redisAddr: string;
  maxConcurrentTorrents: number;
  maxConcurrentImages: number;
  disableIpv6: boolean;
  rootDir: string;
  oidc: OidcConfig;
  subscriptionRefreshHour: number;
  metadata: Metadata;
}

export type OidcConfig = {
  authority: string;
  clientId: string;
  clientSecret: string;
  disablePasswordLogin: boolean;
  autoLogin: boolean;
}

export type Oidc = {
  disablePasswordLogin: boolean;
  autoLogin: boolean;
  enabled: boolean;
}

export enum CacheType {
  MEMORY = "MEMORY",
  REDIS = "REDIS",
}

export const CacheTypes = [{
  value: CacheType.MEMORY,
  key: 'memory',
}, {
  value: CacheType.REDIS,
  key: 'redis',
}];

export type Metadata = {
  version: string;
  firstInstalledVersion: string;
  installDate: Date;
  lastUpdateDate: Date;
}
