import {Page} from "../components/form/types";
import axios from "axios";
import {
    ContentInfo, DirEntry,
    DownloadRequest,
    InfoStat,
    NavigationItem, NewDirRequest,
    SearchRequest,
    Stats,
    StopRequest,
    SubDirsRequest
} from "./types";

declare const BASE_URL: string;

async function getPage(index: number): Promise<Page | null> {
    return axios.get(`${BASE_URL}/api/pages/${index}`)
        .then((res) => {
            if (!res || !res.data) {
                return null
            }
            return res.data as Page
        })
        .catch((err) => {
            console.debug(err)
            return null
        })
}

async function searchContent(searchRequest: SearchRequest): Promise<ContentInfo[]> {
    return axios.post(`${BASE_URL}/api/search`, searchRequest)
        .then((res) => {
            if (!res || !res.data) {
                throw new Error("No results found")
            }

            const info: ContentInfo[] = res.data as ContentInfo[]
            if (info.length == 0) {
                throw new Error("No results found")
            }
            return info
        })
}

async function loadPages(): Promise<Page[]> {
    return axios.get(`${BASE_URL}/api/pages`)
        .then((res) => {
            if (!res || !res.data) {
                throw new Error("No pages found")
            }

            const pages: Page[] = res.data as Page[]
            if (pages.length == 0) {
                throw new Error("No pages found")
            }
            return pages
        })
}

async function loadNavigation(index: number | null): Promise<NavigationItem[]> {
    return loadPages().then(pages => {
        let nav = [
            {
                name: "Home",
                href: `${BASE_URL}/`,
                current: index == null,
            },
        ];
        nav.push(
            ...pages.map((page, i) => {
                return {
                    name: page.title,
                    href: `${BASE_URL}/page?index=${i}`,
                    current: i == index,
                };
            }),
        );
        return nav
    })
}

async function getStats(): Promise<Stats> {
    return axios.get(`${BASE_URL}/api/stats`)
        .then((res) => {
            if (!res || res.status != 200) {
                throw new Error("Unable to load stats")
            }
            if (!res.data) {
                return {}
            }
            return res.data as { [key: string]: InfoStat }
        })
}

async function stopDownload(request: StopRequest): Promise<void> {
    return axios.post(`${BASE_URL}/api/stop/`, request)
        .then((res) => {
            if (!res || res.status != 202) {
                throw new Error("Error stopping download")
            }
        }).catch(err => {
            console.log(err)
            throw new Error("Error stopping download")
        })
}

async function startDownload(downloadRequest: DownloadRequest): Promise<void> {
    return axios.post(`${BASE_URL}/api/download`, downloadRequest)
        .then((res) => {
            if (!res || res.status != 202) {
                throw new Error("Error starting download")
            }
        })
        .catch(err => {
            console.log(err)
            throw new Error("Error starting download")
        })
}

async function getSubDirs(req: SubDirsRequest): Promise<DirEntry[]> {
    return axios.post(`${BASE_URL}/api/io/ls`, req)
        .then(res => {
            if (!res) {
                throw new Error("Error loading subdirectories")
            }
            if (!res.data) {
                return []
            }
            return res.data as DirEntry[]
        })
}

async function createNewDir(req: NewDirRequest): Promise<void> {
    return axios.post(`${BASE_URL}/api/io/create`, req)
        .then(res => {
            if (!res || res.status != 201) {
                throw new Error("Error creating new directory")
            }
        })
        .catch(err => {
            console.log(err)
            throw new Error("Error creating new directory")
        })
}

export {getPage, searchContent, loadNavigation, loadPages, getStats, stopDownload, startDownload, getSubDirs, createNewDir}