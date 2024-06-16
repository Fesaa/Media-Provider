interface SearchRequest {
    provider: string[];
    query: string;
    modifiers: { [key: string]: string[] };
}

interface DownloadRequest {
    provider: string;
    id: string;
    base_dir: string;
}

interface StopRequest {
    provider: string;
    id: string;
    delete_files: boolean;
}

 interface ContentInfo {
    Name: string;
    Description: string;
    Date: string;
    Size: string;
    Seeders: string;
    Leechers: string;
    Downloads: string;
    Link: string;
    InfoHash: string;
    ImageUrl: string;
    RefUrl: string;
    Provider: string;
}

type NavigationItem = {
    name: string;
    href: string;
    current: boolean;
};

interface InfoStat {
    provider: string;
    id: string;
    name: string;
    size: string;
    progress: number;
    speed_type: string;
    speed: SpeedData;
    download_dir: string;
}

interface SpeedData {
    time: number;
    speed: number;
}

interface SubDirsRequest {
    dir: string;
    files: boolean;
}

interface NewDirRequest {
    baseDir: string;
    newDir: string;
}

interface DirEntry {
    name: string;
    dir: boolean;
}

interface QueueStat {
    provider: string;
    id: string;
    name: string;
}

interface Stats {
    running: InfoStat[]
    queued: QueueStat[]
}

export {SearchRequest, ContentInfo, NavigationItem, InfoStat, Stats, StopRequest, DownloadRequest, SubDirsRequest, NewDirRequest, SpeedData, DirEntry, QueueStat}