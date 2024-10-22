
export type Page = {
  id: number;
  title: string;
  providers: Provider[];
  modifiers: { [key: string]: Modifier };
  dirs: string[];
  custom_root_dir: string;
}

export type Modifier = {
  title: string;
  type: ModifierType;
  values: { [key: string]: string };
}

export enum Provider {
  SUKEBEI = 1,
  NYAA,
  YTS,
  LIMETORRENTS,
  SUBSPLEASE,
  MANGADEX,
  WEBTOON
}

export const providerNames = Object.keys(Provider).filter(key => isNaN(Number(key))) as string[];
export const providerValues = Object.values(Provider).filter(value => typeof value === 'number') as number[];

export enum ModifierType {
  DROPDOWN = 1,
  MULTI,
}
