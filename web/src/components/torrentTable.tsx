import { ArrowDownTrayIcon } from "@heroicons/react/16/solid";
import axios from "axios";
import React from "react";
import NotificationHandler from "../notifications/handler";
import ErrorNotification from "../notifications/error";

export type TorrentInfo = {
  Name: string;
  Description: string;
  Date: string;
  Size: string;
  Seeders: string;
  Leechers: string;
  Downloads: string;
  Link: string;
  InfoHash: string;
};

export type DownloadOptions = {
  baseDir: string;
  url: boolean;
};

export default function TorrentTable(props: {
  torrents: TorrentInfo[];
  options: DownloadOptions;
}) {
  async function downloadTorrent(infoHash: string): Promise<void> {
    const requestBody = {
      info: infoHash,
      base_dir: props.options.baseDir,
      url: props.options.url,
    };

    axios.post(`${BASE_URL}/api/download`, requestBody)
      .catch(err => {
        console.error(err);
        NotificationHandler.addErrorNotificationByTitle("Error while downloading downloading!");
      })
      .then(res => {
        if (res == null) {
          return;
        }

        if (res.status == 202) {
          NotificationHandler.addSuccesNotificationByTitle("Torrent is downloading!",);
        } else {
          NotificationHandler.addErrorNotificationByTitle("Error while downloading!");
        }
      })
  }

  return (
    <div>
      <div>
        {props.torrents.length > 0 && (
          <table className="my-10 ml-10 mr-10 overflow-x-auto">
            <thead>
              <tr className="rounded-lg bg-blue-400">
                <th className="p-5">Name</th>
                <th className="p-5 hidden md:table-cell">Date</th>
                <th className="p-5 hidden md:table-cell">Size</th>
                <th className="p-5 hidden md:table-cell">Downloads</th>
                <th className="p-5 hidden md:table-cell">Seeds</th>
                <th className="p-5 hidden md:table-cell">Leeches</th>
                <th className="p-5"></th>
              </tr>
            </thead>
            <tbody>
              {props.torrents.map((torrent) => (
                <tr
                  className="odd:bg-white even:bg-amber-50"
                  key={torrent.InfoHash}
                >
                  <td className="p-2 text-sm">{torrent.Name}</td>
                  <td className="p-2 text-sm text-center hidden md:table-cell">
                    {torrent.Date}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell">
                    {torrent.Size}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell">
                    {torrent.Downloads}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell">
                    {torrent.Seeders}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell">
                    {torrent.Leechers}
                  </td>
                  <td>
                    <ArrowDownTrayIcon
                      type="button"
                      onClick={(e) => {
                        downloadTorrent(torrent.InfoHash);
                      }}
                      style={{ cursor: "pointer" }}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
