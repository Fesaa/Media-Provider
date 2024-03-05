import { ArrowDownTrayIcon } from "@heroicons/react/16/solid";
import axios from "axios";
import React, { useState } from "react";

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
  const [title, setTitle] = useState("");
  const [success, setSuccess] = useState(false);
  const [errorTitle, setErrorTitle] = useState("");
  const [error, setError] = useState(false);

  async function downloadTorrent(infoHash: string): Promise<void> {
    const requestBody = {
      info: infoHash,
      base_dir: props.options.baseDir,
      url: props.options.url,
    };

    try {
      const response = await axios.post("/api/download", requestBody);
      if (response.status == 202) {
        setTitle("Torrent started downloading!");
        setSuccess(true);
      } else {
        setErrorTitle("Error downloading torrent: " + response.statusText);
        setError(true);
      }
    } catch (err) {
      console.error("Error downloading torrent", err);
    }
  }

  return (
    <div>
      <div>
        <button
          type="button"
          onClick={(e) => setSuccess(false)}
          className="fixed right-4 top-4 z-50 rounded-md bg-green-500 px-4 py-2 text-white transition hover:bg-green-600"
          style={success ? { display: "block" } : { display: "none" }}
        >
          <div className="flex items-center space-x-2">
            <span className="text-3xl">
              <i className="bx bx-check"></i>
            </span>
            <p className="font-bold">{title}</p>
          </div>
        </button>

        <button
          type="button"
          onClick={(e) => setError(false)}
          className="fixed right-4 top-4 z-50 rounded-md bg-red-500 px-4 py-2 text-white transition hover:bg-red-600"
          style={error ? { display: "block" } : { display: "none" }}
        >
          <div className="flex items-center space-x-2">
            <span className="text-3xl">
              <i className="bx bx-check"></i>
            </span>
            <p className="font-bold">{errorTitle}</p>
          </div>
        </button>

        {props.torrents.length > 0 && (
          <table className="my-10 ml-10 mr-10">
            <thead>
              <tr className="rounded-lg bg-blue-400">
                <th className="p-5">Name</th>
                <th className="p-5">Date</th>
                <th className="p-5">Size</th>
                <th className="p-5">Downloads</th>
                <th className="p-5">Seeds</th>
                <th className="p-5">Leeches</th>
                <th className="p-5"></th>
              </tr>
            </thead>
            <tbody>
              {props.torrents.map((torrent) => (
                <tr
                  className="odd:bg-white even:bg-amber-50"
                  key={torrent.InfoHash}
                >
                  <td className="p-2">{torrent.Name}</td>
                  <td className="p-2 text-center">{torrent.Date}</td>
                  <td className="p-2 text-center">{torrent.Size}</td>
                  <td className="p-2 text-center">{torrent.Downloads}</td>
                  <td className="p-2 text-center">{torrent.Seeders}</td>
                  <td className="p-2 text-center">{torrent.Leechers}</td>
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
