export type TorrentInfo = {
  infoHash: string;
  name: string;
  size: number;
  progress: number;
  completed: number;
  speed: string;
};

export type StatsInfo = Record<string, TorrentInfo>;
