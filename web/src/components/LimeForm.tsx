import React, { FormEvent, useState } from "react";
import axios from "axios";
import Torrent, { TorrentInfo } from "./torrent";

type LimeTorrent = {
  Name: string;
  Url: string;
  Seed: string;
  Leech: string;
  Size: string;
  Added: string;
};

export default function MoviesForm() {
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("ALL");
  const [dir, setDir] = useState("Movies");

  const [results, setResults] = useState<TorrentInfo[]>([]);

  function onSubmit(e: FormEvent) {
    console.log("Getting torrents");
    e.preventDefault();

    const data = JSON.stringify({
      provider: "limetorrents",
      query,
      category,
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

        const limeTorrents: LimeTorrent[] = res.data;
        const torrents: TorrentInfo[] = limeTorrents.map((torrent) => {
          return {
            Category: "",
            Name: torrent.Name,
            Description: "",
            Date: torrent.Added,
            Size: torrent.Size,
            Seeders: torrent.Seed,
            Leechers: torrent.Leech,
            Downloads: "",
            IsTrusted: "",
            IsRemake: "",
            Link: torrent.Url,
            GUID: "",
            CategoryID: "",
            InfoHash: "",
            CoverImage: "",
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
    <div className="justify-items-center">
      <section className="bg-gray-50 dark:bg-gray-900">
        <div className="mx-auto flex flex-col items-center justify-center px-6 py-8 md:h-screen lg:py-0">
          <div className="w-full rounded-lg bg-white shadow sm:max-w-md md:mt-0 xl:p-0 dark:border dark:border-gray-700 dark:bg-gray-800">
            <div className="space-y-4 p-6 sm:p-8 md:space-y-6">
              <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
                Search for content to download
              </h1>
              <form className="space-y-4 md:space-y-6" onSubmit={onSubmit}>
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
                      <option value={"TV"}> Likes </option>
                      <option value={"OTHER"}> Date Added </option>
                    </select>
                  </div>
                </div>

                <div className="flex flex-wrap justify-around">
                  <div>
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
                      onChange={(e) => setDir(e.target.value)}
                    >
                      <option value={"Movies"}> Movies </option>
                      <option value={"Anime"}> Anime </option>
                      <option value={"Series"}> Series </option>
                    </select>
                  </div>
                </div>

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
        <ul className="mx-auto flex flex-wrap gap-4">
          {results.map((t: TorrentInfo) => (
            <li key={t.Link} className="p-4">
              <Torrent torrent={t} baseDir={dir} url={true} />
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
