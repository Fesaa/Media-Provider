import React from "react";
import NotificationHandler from "../notifications/handler";
import {QueueStat, StopRequest} from "../utils/types";
import {stopDownload} from "../utils/http";
import {TrashIcon} from "@heroicons/react/16/solid";


export default function QueueLine(props: {
    stat: QueueStat;
    refreshFunc: (repeat: boolean) => void;
}) {
    const i = props.stat;

    async function remove(id: string) {
        const stopRequest: StopRequest = {
            provider: i.provider,
            id: id,
            delete_files: true,
        }
        stopDownload(stopRequest).then(() => {
            NotificationHandler.addSuccesNotificationByTitle("Successfully removed content from queue");
        }).catch((err) => {
            NotificationHandler.addErrorNotificationByTitle(err.message);
        })
    }

    return (
        <div
            className="flex flex-grow flex-row bg-white border-2 border-solid border-gray-200 p-2 justify-between"
            key={i.id}>
            <span className="break-all min-w-20 md:min-w-56">
                {i.name || i.id}
            </span>
            <div className="flex flex-col justify-center">
                <TrashIcon className="h-8 md:h-10 w-8 md:w-10 text-red-500 hover:cursor-pointer"
                           onClick={() => remove(i.id)}/>
            </div>
        </div>
    );
}
