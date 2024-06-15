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

type InfoStat = {
    Provider: string;
    Completed: number;
    InfoHash: string;
    Name: string;
    Progress: number;
    Size: number;
    Speed: string;
};

interface SubDirsRequest {
    dir: string;
}

interface NewDirRequest {
    baseDir: string;
    newDir: string;
}

type Stats = { [key: string]: InfoStat };

export {SearchRequest, ContentInfo, NavigationItem, InfoStat, Stats, StopRequest, DownloadRequest, SubDirsRequest, NewDirRequest}