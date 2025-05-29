export type Page = {
  ID: number;
  sortValue: number;
  title: string;
  icon: string;
  providers: Provider[];
  modifiers: Modifier[];
  dirs: string[];
  custom_root_dir: string;
}

export type Modifier = {
  ID: number;
  title: string;
  key: string;
  type: ModifierType;
  values: ModifierValue[];
}

export type ModifierValue = {
  key: string;
  value: string;
  default: boolean;
}

export enum Provider {
  NYAA = 2,
  YTS,
  LIMETORRENTS,
  SUBSPLEASE,
  MANGADEX,
  WEBTOON,
  DYNASTY,
  BATO
}

export const Providers = [
  {
    label: "Nyaa",
    value: Provider.NYAA
  },
  {
    label: "YTS",
    value: Provider.YTS
  },
  {
    label: "LimeTorrents",
    value: Provider.LIMETORRENTS
  },
  {
    label: "SubsPlease",
    value: Provider.SUBSPLEASE
  },
  {
    label: "MangaDex",
    value: Provider.MANGADEX
  },
  {
    label: "Webtoon",
    value: Provider.WEBTOON
  },
  {
    label: "Dynasty",
    value: Provider.DYNASTY
  },
  {
    label: "Bato",
    value: Provider.BATO
  }
];


export const providerNames = Object.keys(Provider).filter(key => isNaN(Number(key))) as string[];
export const providerValues = Object.values(Provider).filter(value => typeof value === 'number') as number[];

export enum ModifierType {
  DROPDOWN = 1,
  MULTI,
}

export type DownloadMetadata = {
  definitions: DownloadMetadataDefinition[];
}

export type DownloadMetadataDefinition = {
  key: string;
  formType: DownloadMetadataFormType;
  defaultOption: string;
  options: MetadataOption[];
}

export type MetadataOption = {
  key: string;
  value: string;
}

export enum DownloadMetadataFormType {
  SWITCH,
  DROPDOWN,
  MULTI,
  TEXT
}
