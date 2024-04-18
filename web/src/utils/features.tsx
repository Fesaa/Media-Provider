import axios from "axios";

export type NavigationItem = {
    name: string;
    href: string;
    current: boolean
}

function wrapper(s: string): string {
    return `${BASE_URL}${s}`;
}

function titleCaseWord(word: string) {
    if (!word) return word;
    return word[0].toUpperCase() + word.slice(1).toLowerCase();
}

export const defaultNavigation: NavigationItem[] = [
    { name: "Home", href: wrapper("/"), current: false },
    { name: "Anime", href: wrapper("/anime"), current: false },
    { name: "Movies", href: wrapper("/movies"), current: false },
    { name: "Lime", href: wrapper("/lime"), current: false },
];


export async function getNavigationItems(): Promise<NavigationItem[]> {
    return axios.get(`${BASE_URL}/api/features`)
        .catch(err => Promise.reject(err))
        .then(res => {
            if (res == null || res.data == null) {
                return Promise.reject("No data received")
            }

            const features: string[] = res.data
            const nav: NavigationItem[] = [defaultNavigation[0]];
            for (const feature of features) (
                nav.push({
                    name: titleCaseWord(feature),
                    href: feature,
                    current: false,
                })
            )

            return Promise.resolve(nav);
        })
}