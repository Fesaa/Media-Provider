export type SearchProps = {
  page: Page;
};

export interface Page {
  title: string;
  search: SearchConfig;
}

export interface SearchConfig {
  providers: string[];
  search_modifiers: { [key: string]: Modifier };
  root_dirs: string[];
  custom_root_dir: string;
}

export interface Modifier {
  title: string
  type: string
  values: Pair[]
}

export interface Pair {
  key: string;
  name: string;
}
