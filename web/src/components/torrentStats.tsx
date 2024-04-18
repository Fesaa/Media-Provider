import React from "react";
import axios from "axios";
import NotificationHandler from "../notifications/handler";
import { TrashIcon } from "@heroicons/react/16/solid";

type TorrentStat = {
  Completed: number;
  InfoHash: string;
  Name: string;
  Progress: number;
  Size: number;
  Speed: string;
};

function bytesToSize(bytes: number): string {
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  if (bytes === 0) return '0 Byte';
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return parseFloat((bytes / Math.pow(1024, i)).toFixed(2)) + ' ' + sizes[i];
}

export default function Torrent(props: {
  TKey: string;
  torrent: TorrentStat;
  refreshFunc: (repeat: boolean) => void;
}) {
  const torrent = props.torrent;

  async function remove(hash: string) {
    axios.get(`${BASE_URL}/api/stop/${hash}`)
      .catch(e => {
        console.log(e);
        NotificationHandler.addErrorNotificationByTitle("Error stopping download");
      })
      .then(res => {
        if (res == null) {
          NotificationHandler.addErrorNotificationByTitle("Error stopping download");
          return
        }
        if (res.status == 202) {
          NotificationHandler.addSuccesNotificationByTitle("Download stopped");
        } else {
          NotificationHandler.addErrorNotificationByTitle("Error stopping download");
        }
      })
  }

  return (
    <tr
      className="even:bg-white border-gray-300 hover:bg-gray-300"
      key={torrent.InfoHash}
    >
      <td className="p-2 text-sm">
        <div className="">
          {torrent.Name}
        </div>
      </td>
      <td className="p-2 text-sm text-center hidden md:table-cell">{bytesToSize(props.torrent.Size)}</td>
      <td className="p-2 text-sm text-center">
        {props.torrent.Completed}%
        <div className="h-2.5 w-full rounded-full bg-gray-200 dark:bg-gray-700  md:block">
          <div
            className="h-2.5 rounded-full bg-blue-600"
            style={{ width: `${props.torrent.Completed}%` }}
          ></div>
        </div>
      </td>
      <td className="p-2 flex flex-row md:flex-row justify-center">
        <TrashIcon
          className="h-8 w-8"
          type="button"
          onClick={e => remove(torrent.InfoHash)}
          style={{ cursor: "pointer" }}
        />
      </td>
    </tr>
  );
}
