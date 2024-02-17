import React, { useState } from "react";
import axios from "axios";
import PopUpNotification from "./notification";

export type TorrentInfo = {
  Category: string;
  Name: string;
  Description: string;
  Date: string;
  Size: string;
  Seeders: string;
  Leechers: string;
  Downloads: string;
  IsTrusted: string;
  IsRemake: string;
  Link: string;
  GUID: string;
  CategoryID: string;
  InfoHash: string;
  CoverImage: string;
};

function shadowColour(torrent: TorrentInfo): String {
  const seeders = Math.max(Number(torrent.Seeders), 1);
  const leechers = Math.max(Number(torrent.Leechers), 1);

  const ratio = seeders / leechers;

  if (ratio < 1) {
    return "shadow-lg shadow-red-500";
  }

  if (seeders < 10) {
    return "shadow-lg shadow-red-300";
  }

  if (seeders < 50) {
    return "shadow-lg shadow-yellow-300";
  }

  return "shadow-lg shadow-green-500";
}

export default function Torrent(props: {
  torrent: TorrentInfo;
  baseDir: string;
}) {
  const [open, setOpen] = useState(false);

  function download(
    hash: string,
    baseDir: string,
  ): (e: React.MouseEvent<HTMLAnchorElement>) => void {
    return (e) => {
      e.preventDefault();

      axios
        .get(`/api/download/${hash}?base_dir=${baseDir}`)
        .then((res) => {
          if (res.status == 202) {
            setOpen(true);
          }
        })
        .then((e) => {
          console.log(e);
        });
    };
  }

  return (
    <div
      className={
        "max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow dark:border-gray-700 dark:bg-gray-800" +
        shadowColour(props.torrent)
      }
    >
      {props.torrent.CoverImage != "" && (
        <img
          className="mx-auto h-32 w-32 rounded-full"
          src={props.torrent.CoverImage}
          alt=""
        />
      )}
      <h5 className="mb-2 font-bold tracking-tight text-gray-900 dark:text-white">
        {props.torrent.Name.replace(/\./g, " ")}
      </h5>
      <div className="flex flex-wrap gap-2">
        <div className="center relative inline-block select-none whitespace-nowrap rounded-lg bg-blue-500 px-3.5 py-2 align-baseline font-sans text-xs font-bold uppercase leading-none text-white">
          <div className="mt-px">Size: {props.torrent.Size}</div>
        </div>
        <div className="center relative inline-block select-none whitespace-nowrap rounded-lg bg-red-500 px-3.5 py-2 align-baseline font-sans text-xs font-bold uppercase leading-none text-white">
          <div className="mt-px">Leachers: {props.torrent.Leechers}</div>
        </div>
        <div className="center relative inline-block select-none whitespace-nowrap rounded-lg bg-green-500 px-3.5 py-2 align-baseline font-sans text-xs font-bold uppercase leading-none text-white">
          <div className="mt-px">Seeders: {props.torrent.Seeders}</div>
        </div>
        {props.torrent.Downloads != "" && (
          <div className="center relative inline-block select-none whitespace-nowrap rounded-lg bg-pink-500 px-3.5 py-2 align-baseline font-sans text-xs font-bold uppercase leading-none text-white">
            <div className="mt-px">Downloads: {props.torrent.Downloads}</div>
          </div>
        )}
        <a
          onClick={download(props.torrent.InfoHash, props.baseDir)}
          className="inline-flex items-center justify-center rounded-lg bg-purple-700 p-5 px-3 py-2 text-center text-sm font-medium text-white hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
        >
          Download
        </a>
      </div>
      <PopUpNotification
        open={open}
        setOpen={setOpen}
        title="Succes!"
        desc="Your download has started"
      />
    </div>
  );
}
