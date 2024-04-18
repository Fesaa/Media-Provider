import React, { FormEvent, useEffect, useState } from "react";
import axios from "axios";
import { LimeTorrent } from "../response/SearchResults";
import TorrentTable, { TorrentInfo } from "./torrentTable";
import DirFormComponent from "./io/form";
import NotificationHandler from "../notifications/handler";

export default function LimeForm() {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("ALL");
  const [dir, setDir] = useState("");
  const [dirs, setDirs] = useState([]);
  const [customDir, setCustomDir] = useState<string>("");

  const [results, setResults] = useState<TorrentInfo[]>([]);

  function onLoad() {
    const data = {
      dir: "",
    };
    axios.post(`${BASE_URL}/api/io/ls`, data)
      .catch(err => {
        NotificationHandler.addErrorNotificationByTitle("Could not load root dirs.");
        console.error(err)
      }).then(res => {
        if (res == null) {
          NotificationHandler.addErrorNotificationByTitle("No root dirs found");
          return;
        }
        if (res.data == null) {
          NotificationHandler.addErrorNotificationByTitle("No root dirs found");
          return;
        }

        setDirs(res.data);
        if (res.data.length > 0) {
          setDir(res.data[0]);
        }
      })
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (query == "") {
      NotificationHandler.addErrorNotificationByTitle("Query cannot be empty");
      return;
    }

    const data = JSON.stringify({
      provider: "limetorrents",
      query,
      category,
    });

    axios
      .post(`${BASE_URL}/api/search`, data, {
        headers: {
          "Content-Type": "application/json",
        },
      })
      .then((res) => {
        if (res.data == null || res.data.length == 0) {
          NotificationHandler.addErrorNotificationByTitle("No results found");
          return;
        }

        const limeTorrents: LimeTorrent[] = res.data;
        const torrents: TorrentInfo[] = limeTorrents.map((torrent) => {
          return {
            Name: torrent.Name,
            Description: "",
            Date: torrent.Added,
            Size: torrent.Size,
            Seeders: torrent.Seed,
            Leechers: torrent.Leach,
            Downloads: "N/A",
            Link: torrent.Url,
            InfoHash: torrent.Hash,
          };
        });

        setResults(torrents);
        document
          .getElementById("search-results")!
          .scrollIntoView({ behavior: "smooth" });
      })
      .catch((err) => console.error(err));
  }

  useEffect(onLoad, []);

  return (
    <div className="justify-items-center bg-gray-50 dark:bg-gray-900">
      <section>
        <div className="mx-auto flex flex-col items-center justify-center px-6 py-8 md:h-screen lg:py-0">
          <div className="w-full rounded-lg bg-white shadow sm:max-w-md md:mt-0 xl:p-0 dark:border dark:border-gray-700 dark:bg-gray-800">
            <div className="space-y-4 p-6 sm:p-8 md:space-y-6">
              <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
                Search for content to download
              </h1>
              <form
                className="space-y-4 md:space-y-6"
                onSubmit={onSubmit}
                noValidate={true}
              >
                <div>
                  <label
                    htmlFor="query"
                    className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
                  >
                    Query
                  </label>
                  <input
                    type="text"
                    name="query"
                    id="query"
                    className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
                    required
                    onChange={(e) => setQuery(e.target.value)}
                  />
                </div>
                <div className="flex flex-col md:flex-row grow justify-around">
                  <div className="flex flex-wrap justify-start">
                    <label
                      htmlFor="category"
                      className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
                    >
                      Category
                    </label>
                    <select
                      name="category"
                      id="category"
                      className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
                      onChange={(e) => setCategory(e.target.value)}
                    >
                      <option value={"ALL"}> All </option>
                      <option value={"ANIME"}> Anime </option>
                      <option value={"MOVIES"}> Movies </option>
                      <option value={"TV"}> Tv </option>
                      <option value={"OTHER"}> Other </option>
                    </select>
                  </div>

                  <div className="flex flex-wrap justify-start">
                    <label
                      htmlFor="dir"
                      className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
                    >
                      Directory
                    </label>
                    <select
                      name="dir"
                      id="dir"
                      className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
                      onChange={(e) => {
                        if (e.target.value != "") {
                          setDir(e.target.value);
                        }
                      }}
                    >
                      {dirs.map(dirName => {
                        return <option key={dirName} value={dirName}> {dirName} </option>
                      })}
                    </select>
                  </div>
                </div>

                <DirFormComponent setter={setCustomDir} base="" name="Root" />

                <button
                  type="submit"
                  className="focus:ring-primary-300 dark:bg-primary-600 dark:hover:bg-primary-700 dark:focus:ring-primary-800 w-full rounded-lg bg-blue-600 px-5 py-2.5 text-center text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-4"
                >
                  Search
                </button>
              </form>
            </div>
          </div>
        </div>
      </section>

      <section
        id="search-results"
        className="flex items-center justify-center justify-items-center"
      >
        <TorrentTable
          torrents={results}
          options={{
            baseDir: customDir.trim() === "" ? dir : customDir,
            url: false,
          }}
        />
      </section>
    </div>
  );
}
