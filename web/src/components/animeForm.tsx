import React, { FormEvent, useState } from "react";
import axios from "axios";
import TorrentTable from "./torrentTable";

export default function AnimeForm() {
  const provider = "nyaa";
  const [query, setQuery] = useState("");
  const [category, setCategory] = useState("all");
  const [sort_by, setSortBy] = useState("downloads");

  const [results, setResults] = useState([]);

  function onSubmit(e: FormEvent) {
    console.log("Getting torrents");
    e.preventDefault();

    const data = JSON.stringify({
      provider,
      query,
      category,
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
        setResults(res.data);
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
                Search Anime to download
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
                      <option value={"downloads"}> Downloads </option>
                      <option value={"date"}> Date </option>
                      <option value={"size"}> Size </option>
                      <option value={"seeders"}> Seeders </option>
                      <option value={"leechers"}> Leechers </option>
                      <option value={"comments"}> Comments </option>
                    </select>
                  </div>

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
                      <option value={"all"}>All categories</option>
                      <option value={"anime"}> Anime </option>
                      <option value={"anime-amv"}> - Music Video </option>
                      <option value={"anime-eng"}> - English Translated</option>
                      <option value={"anime-non-eng"}>
                        {" "}
                        - Non-English Translated
                      </option>
                      <option value={"anime-raw"}> - Raw </option>
                      <option value={"audio"}> Audio </option>
                      <option value={"audio-lossless"}> - Lossless </option>
                      <option value={"audio-lossy"}> - Lossy </option>
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
        <TorrentTable
          torrents={results}
          options={{ baseDir: "Anime", url: false }}
        />
      </section>
    </div>
  );
}
