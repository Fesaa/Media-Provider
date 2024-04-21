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
          <table className="bg-white border border-gray-300 m-2 md:m-10 ">
            <thead>
              <tr className="">
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b">Name</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">Date</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">Size</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">Downloads</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">Seeds</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">Leeches</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b"></th>
              </tr>
            </thead>
            <tbody>
              {props.torrents.map((torrent) => (
                <tr
                  className="even:bg-white border-gray-300 hover:bg-gray-300 border"
                  key={torrent.InfoHash}
                >
                  <td className="p-2 text-sm border">{torrent.Name.replace(".", " ")}</td>
                  <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Date}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Size}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Downloads}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Seeders}
                  </td>
                  <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Leechers}
                  </td>
                  <td className="border">
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
