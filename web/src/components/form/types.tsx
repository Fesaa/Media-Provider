export type SearchProps = {
  page: Page;
};

export interface Page {
  title: string;
  providers: string[];
  modifiers: { [key: string]: Modifier };
  dirs: string[];
  custom_root_dir: string;
}


export interface Modifier {
  title: string
  type: string
  values: { [key: string]: string }
}
