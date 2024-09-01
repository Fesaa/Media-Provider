
export type Page = {
  title: string;
  provider: Provider[];
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
}

export enum ModifierType {
  DROPDOWN = 1,
  MULTI,
}
