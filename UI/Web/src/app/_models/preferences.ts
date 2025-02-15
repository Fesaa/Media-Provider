export type Preferences = {
  subscriptionRefreshHour: number,
  logEmptyDownloads: boolean,
  dynastyGenreTags: Tag[],
  blackListedTags: Tag[],
};


export type Tag = {
  name: string,
  normalizedName: string,
}


export function normalize(s: string): string {
  return s.replace(/[^a-zA-Z0-9]/g, "").toLowerCase();
}
