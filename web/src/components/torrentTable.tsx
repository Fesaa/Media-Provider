import {
  ArrowDownTrayIcon,
  InformationCircleIcon,
} from "@heroicons/react/16/solid";
import axios from "axios";
import React, { useState } from "react";
import NotificationHandler from "../notifications/handler";

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
  ImageUrl: string;
  RefUrl: string;
  Provider: string;
};

export type DownloadOptions = {
  baseDir: string;
  url: boolean;
};

export default function TorrentTable(props: {
  torrents: TorrentInfo[];
  options: DownloadOptions;
}) {
  const [open, setOpen] = useState<boolean>(false);
  const [imageUrl, setImageUrl] = useState<string>("");

  async function downloadTorrent(infoHash: string, provider: string): Promise<void> {
    const requestBody = {
      provider: provider,
      info: infoHash,
      base_dir: props.options.baseDir,
    };

    axios
      .post(`${BASE_URL}/api/download`, requestBody)
      .catch((err) => {
        console.error(err);
        NotificationHandler.addErrorNotificationByTitle(
          "Error while downloading downloading!",
        );
      })
      .then((res) => {
        if (res == null) {
          return;
        }

        if (res.status == 202) {
          NotificationHandler.addSuccesNotificationByTitle(
            "Torrent is downloading!",
          );
        } else {
          NotificationHandler.addErrorNotificationByTitle(
            "Error while downloading!",
          );
        }
      });
  }

  function hasAtLeast(f: (torrent: TorrentInfo) => string): boolean {
    return props.torrents.map(f).some(str => str !== null && str !== undefined && str.trim() !== '');
  }


  const showDate = hasAtLeast(torrent => torrent.Date);
  const showSize = hasAtLeast(torrent => torrent.Size);
  const showDownloads = hasAtLeast(torrent => torrent.Downloads);
  const showSeeders = hasAtLeast(torrent => torrent.Seeders);
  const showLeechers = hasAtLeast(torrent => torrent.Leechers);

  return (
    <div>
      {open && <div></div>}
      <div>
        {props.torrents.length > 0 && (
          <table className="bg-white border border-gray-300 m-2 md:m-10 ">
            <thead>
              <tr className="">
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b">
                  Name
                </th>
                {showDate &&  <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                  Date
                </th>}
                {showSize && <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                  Size
                </th>}
                {showDownloads && <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                  Downloads
                </th>}
                {showSeeders && <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                  Seeds
                </th>}
                {showLeechers && <th
                    className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                  Leeches
                </th>}
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b"></th>
              </tr>
            </thead>
            <tbody>
              {props.torrents.map((torrent) => (
                <tr
                  className="even:bg-white border-gray-300 hover:bg-gray-300 border"
                  key={torrent.InfoHash}
                >
                  <td className="p-2 text-sm border">
                    <div className="flex flex-row text-center space-x-2">
                      <a
                        href={torrent.RefUrl}
                        target="_blank"
                        className="hover:cursor-pointer hover:underline"
                      >
                        {torrent.Name.replace(".", " ")}
                      </a>
                      {torrent.ImageUrl && (
                        <InformationCircleIcon
                          className="w-4 h-4"
                          onClick={() => {
                            setImageUrl(torrent.ImageUrl);
                            setOpen(true);
                          }}
                        />
                      )}
                    </div>
                  </td>
                  {showDate && <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Date}
                  </td>}
                  {showSize && <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Size}
                  </td>}
                  {showDownloads && <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Downloads}
                  </td>}
                  {showSeeders && <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Seeders}
                  </td>}
                  {showLeechers && <td className="p-2 text-sm text-center hidden md:table-cell border">
                    {torrent.Leechers}
                  </td>}
                  <td className="border">
                    <ArrowDownTrayIcon
                      type="button"
                      onClick={(e) => {
                        downloadTorrent(torrent.InfoHash, torrent.Provider);
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
