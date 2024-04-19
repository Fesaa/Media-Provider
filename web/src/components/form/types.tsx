export type SearchProps = {
  page: Page;
};

export interface Page {
  title: string;
  search: SearchConfig;
}

export type SearchProvider = "nyaa" | "yts" | "lime";

export interface SearchConfig {
  provider: SearchProvider;
  categories: Category[];
  sorts: SortBy[];
  root_dirs: string[];
  custom_root_dir: string;
}

export interface Category {
  key: string;
  name: string;
}

export interface SortBy {
  key: string;
  name: string;
}
