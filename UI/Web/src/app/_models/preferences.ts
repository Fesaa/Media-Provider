export type Preferences = {
  subscriptionRefreshHour: number,
  logEmptyDownloads: boolean,
  coverFallbackMethod: CoverFallbackMethod,
  dynastyGenreTags: Tag[],
  blackListedTags: Tag[],
};

export enum CoverFallbackMethod {
  CoverFallbackFirst = 0,
  CoverFallbackLast = 1,
  CoverFallbackNone = 2,
}

export const CoverFallbackMethods = [
  {label: "First", value: CoverFallbackMethod.CoverFallbackFirst},
  {label: "Last", value: CoverFallbackMethod.CoverFallbackLast},
  {label: "None", value: CoverFallbackMethod.CoverFallbackNone},
]


export type Tag = {
  name: string,
  normalizedName: string,
}


export function normalize(s: string): string {
  return s.replace(/[^a-zA-Z0-9]/g, "").toLowerCase();
}
