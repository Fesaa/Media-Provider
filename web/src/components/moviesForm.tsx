import React, { FormEvent, useState } from "react";
import axios from "axios";
import TorrentTable, { TorrentInfo } from "./torrentTable";
import { YTSMovie, YTSTorrent } from "../response/SearchResults";
import DirFormComponent from "./io/form";

export default function MoviesForm() {
  const [query, setQuery] = useState("");
  const [sort_by, setSortBy] = useState("title");
  const [customDir, setCustomDir] = useState<string>("");

  const [results, setResults] = useState<TorrentInfo[]>([]);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (query == "") {
      return;
    }

    const data = JSON.stringify({
      provider: "yts",
      query,
      sort_by,
    });

    axios
      .post("/api/search", data, {
        headers: {
          "Content-Type": "application/json",
        },
      })
      .then((res) => {
        if (res.data == null) {
          return;
        }

        const ytsTorrents: YTSMovie[] = res.data;
        const torrents: TorrentInfo[] = ytsTorrents.map((movie) => {
          const torrent =
            movie.torrents.find((t) => t.quality === "1080p") ||
            movie.torrents[0];
          return {
            Category: movie.genres.join(", "),
            Name: movie.title,
            Description: movie.summary,
            Date: movie.year.toString(),
            Size: torrent.size,
            Seeders: torrent.seeds.toString(),
            Leechers: torrent.peers.toString(),
            Downloads: "",
            IsTrusted: "",
            IsRemake: "",
            Link: torrent.url,
            GUID: "",
            CategoryID: "",
            InfoHash: torrent.hash,
            CoverImage: movie.medium_cover_image,
          };
        });

        setResults(torrents);
        document
          .getElementById("search-results")!
          .scrollIntoView({ behavior: "smooth" });
      })
      .catch((err) => console.error(err));
  }

  return (
    <div className="justify-items-center bg-gray-50 dark:bg-gray-900">
      <section className="bg-gray-50 dark:bg-gray-900">
        <div className="mx-auto flex flex-col items-center justify-center px-6 py-8 md:h-screen lg:py-0">
          <div className="w-full rounded-lg bg-white shadow sm:max-w-md md:mt-0 xl:p-0 dark:border dark:border-gray-700 dark:bg-gray-800">
            <div className="space-y-4 p-6 sm:p-8 md:space-y-6">
              <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
                Search for movies to download
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

                <div className="flex flex-wrap justify-around">
                  <div>
                    <label
                      htmlFor="sortby"
                      className="mb-2 block text-sm font-medium text-gray-900 dark:text-white"
                    >
                      Sort By
                    </label>
                    <select
                      name="sort_by"
                      id="sortby"
                      className="focus:ring-primary-600 focus:border-primary-600 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-gray-900 sm:text-sm dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400 dark:focus:border-blue-500 dark:focus:ring-blue-500"
                      onChange={(e) => setSortBy(e.target.value)}
                    >
                      <option value={"title"}> Title </option>
                      <option value={"year"}> Year </option>
                      <option value={"rating"}> Rating </option>
                      <option value={"peers"}> Peers </option>
                      <option value={"seeds"}> Seeders </option>
                      <option value={"download_count"}> Downloads </option>
                      <option value={"like_count"}> Likes </option>
                      <option value={"date_added"}> Date Added </option>
                    </select>
                  </div>
                </div>

                <DirFormComponent
                  setter={setCustomDir}
                  base="Movies"
                  name="Movies"
                />

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
            baseDir: customDir.trim() === "" ? "Movies" : customDir,
            url: false,
          }}
        />
      </section>
    </div>
  );
}