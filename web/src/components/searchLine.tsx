import {ChevronDownIcon, ChevronUpIcon} from "@heroicons/react/24/outline";
import {ArrowDownTrayIcon} from "@heroicons/react/16/solid";
import React, {ReactNode, useState} from "react";
import {ContentInfo, DownloadRequest} from "../utils/types";
import {startDownload} from "../utils/http";
import NotificationHandler from "../notifications/handler";

async function downloadTorrent(id: string, provider: string, baseDir: string, title: string): Promise<void> {
    const downloadRequest: DownloadRequest = {
        provider: provider,
        id: id,
        base_dir: baseDir,
        temp_title: title,
    };

    startDownload(downloadRequest)
        .then(() => {
            NotificationHandler.addSuccesNotificationByTitle("Content is downloading!");
        })
        .catch(err => {
            NotificationHandler.addErrorNotificationByTitle(err.message);
        })
}

export type SearchLineProps = {
    content: ContentInfo;
    baseDir: string;
}

function tag(key: string, value: string): { key: string, value: string } {
    return {key, value};
}

const colours = ["bg-blue-200", "bg-green-200", "bg-yellow-200", "bg-red-200", "bg-purple-200", "bg-pink-200", "bg-indigo-200", "bg-gray-200"];

function tagify(...tags: { key: string, value: string }[]): ReactNode {
    return (
        <div className="flex flex-row space-x-2 overflow-auto">
            {tags.filter(tag => tag.value != "").map((tag, i) => (
                <div key={i} className={`shadow rounded p-1 space-x-2 ${colours[i % colours.length]} whitespace-nowrap flex flex-row`}>
                    {tag.key != "" &&
                        <span className="font-bold whitespace-nowrap">
                            {tag.key}
                        </span>}
                     <span className="whitespace-nowrap">{tag.value}</span>
                </div>
            ))}
        </div>
    );
}

export default function SearchLine(props: SearchLineProps) {
    const [open, setOpen] = useState(false);
    const toggleOpen = () => setOpen(!open);

    const c = props.content;


    return (
        <div className="flex flex-grow flex-col bg-white border-2 border-solid border-gray-200 p-2 text-center mx-2 md:mx-10 rounded shadow">

            <div className={`flex flex-row justify-between items-center ${open && "pb-2 border-b-2 border-gray-200 border-solid"}`}>
                <div className="flex flex-row flex-grow space-x-2 mr-5">
                    {open
                        ? <ChevronUpIcon className="hidden md:block h-6 w-6" onClick={() => toggleOpen()}/>
                        : <ChevronDownIcon className="hidden md:block h-6 w-6" onClick={() => toggleOpen()}/>
                    }
                    <a href={c.RefUrl} target="_blank" className="break-all min-w-20 md:min-w-56 hover:cursor-pointer hover:underline">
                        {c.Name}
                    </a>
                    <div className="flex flex-row flex-grow space-x-2 justify-end">
                        <span className="hidden md:block">
                            ({c.Size})
                        </span>
                    </div>
                </div>

                <div className="flex flex-col justify-center">
                    <ArrowDownTrayIcon className="h-6 md:h-10 w-6 md:w-10 text-blue-500 hover:cursor-pointer"
                                       onClick={() => downloadTorrent(c.InfoHash, c.Provider, props.baseDir, c.Name)}
                    />
                    {open
                        ? <ChevronUpIcon className="md:hidden h-6 w-6" onClick={() => toggleOpen()} />
                        : <ChevronDownIcon className="md:hidden h-6 w-6" onClick={() => toggleOpen()} />
                    }
                </div>

            </div>

            {open && <div className="flex flex-col p-2">
                <div className="flex flex-row space-x-2 p-2">
                    {tagify(tag("", c.Size), tag("Downloads:", c.Downloads), tag("Seeds:", c.Seeders), tag("", c.Date))}
                </div>
                {c.Description && <div className="flex flex-col">
                    <div className="font-bold text-start">
                        Description
                    </div>
                    <div className="p-2 text-start bg-gray-100 rounded break-all">
                        {c.Description}
                    </div>
                </div>}
            </div>}
        </div>
    );
}