import React, { useState } from "react";
import axios from "axios";
import PopUpNotification from "./notification";

type TorrentStat = {
  Completed: number;
  InfoHash: string;
  Name: string;
  Progress: number;
  Size: number;
};

export default function Torrent(props: { TKey: string; torrent: TorrentStat }) {
  const [open, setOpen] = useState(false);

  function remove(
    hash: string,
  ): (e: React.MouseEvent<HTMLAnchorElement>) => void {
    return (e) => {
      e.preventDefault();

      axios
        .get(`/api/stop/${hash}`)
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
    <div className="max-w-sm rounded-lg border border-gray-200 bg-white p-6 shadow dark:border-gray-700 dark:bg-gray-800">
      <h5 className="mb-2 font-bold tracking-tight text-gray-900 dark:text-white">
        {props.torrent.Name.replace(/\./g, " ")}
      </h5>

      <div className="h-2.5 w-full rounded-full bg-gray-200 dark:bg-gray-700">
        <div
          className="h-2.5 rounded-full bg-blue-600"
          style={{ width: `${props.torrent.Completed}%` }}
        ></div>
      </div>

      <a
        onClick={remove(props.TKey)}
        className="inline-flex items-center justify-center rounded-lg bg-red-700 p-5 px-3 py-2 text-center text-sm font-medium text-white hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
      >
        Stop download
      </a>

      <PopUpNotification
        open={open}
        setOpen={setOpen}
        title="Succes!"
        desc="Your download has been cancelled."
      />
    </div>
  );
}
