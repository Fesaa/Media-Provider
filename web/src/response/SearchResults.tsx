export type LimeTorrent = {
  Name: string;
  Url: string;
  Hash: string;
  Size: string;
  Seed: string;
  Leach: string;
  Added: string;
};

export type YTSTorrent = {
  url: string;
  hash: string;
  quality: string;
  type: string;
  seeds: number;
  peers: number;
  size: string;
  dateUploaded: string;
  dateUploadedUnix: number;
};

export type YTSMovie = {
  id: number;
  url: string;
  imdb_code: string;
  title: string;
  title_english: string;
  title_long: string;
  slug: string;
  year: number;
  rating: number;
  genres: string[];
  summary: string;
  descriptionFull: string;
  lang: string;
  back_ground_image: string;
  small_cover_image: string;
  medium_cover_image: string;
  large_cover_image: string;
  state: string;
  torrents: YTSTorrent[];
};

export type NyaTorrent = {
  category: string;
  name: string;
  description: string;
  date: string;
  size: string;
  seeders: string;
  leechers: string;
  downloads: string;
  isTrusted: string;
  isRemake: string;
  comments: string;
  link: string;
  guid: string;
  categoryID: string;
  infoHash: string;
};
