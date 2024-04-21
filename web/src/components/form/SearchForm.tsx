import React, { FormEvent, useEffect, useState } from "react";
import { SearchProps } from "./types";
import TorrentTable, { TorrentInfo } from "../torrentTable";
import NotificationHandler from "../../notifications/handler";
import axios from "axios";
import DirFormComponent from "../io/form";

interface SearchRequest {
  provider: string;
  query: string;
  category?: string;
  sort_by?: string;
  filter?: string;
}

export default function SearchForm(props: SearchProps) {
  const title = props.page.title;
  const searchProvider = props.page.search.provider;
  const sortBys = props.page.search.sorts;
  const categories = props.page.search.categories;
  const dirs = props.page.search.root_dirs;
  const customRootDir = props.page.search.custom_root_dir;

  const [query, setQuery] = useState<string>("");
  const [requestDir, setRequestDir] = useState<string>("");
  const [customRequestDir, setCustomRequestDir] = useState<string>("");
  const [category, setCategory] = useState<string>("");
  const [sortBy, setSortBy] = useState<string>("");
  const [torrents, setTorrents] = useState<TorrentInfo[]>([]);

  async function searchTorrents() {
    if (query == "") {
      NotificationHandler.addErrorNotificationByTitle(
        "Search query cannot be empty",
      );
      return;
    }

    const searchReq: SearchRequest = {
      provider: searchProvider,
      query: query,
      category: category,
      sort_by: sortBy,
    };

    axios
      .post(`${BASE_URL}/api/search`, searchReq)
      .catch((err) => {
        NotificationHandler.addErrorNotificationByTitle(err.message);
        return null;
      })
      .then((res) => {
        if (res == null || res.data == null) {
          NotificationHandler.addErrorNotificationByTitle("No results found");
          return;
        }

        setTorrents(res.data);
        document
          .getElementById("search-results")!
          .scrollIntoView({ behavior: "smooth" });
      });
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    searchTorrents();
  }

  useEffect(() => {
    if (dirs.length > 0) {
      setRequestDir(dirs[0]);
    } else {
      setRequestDir(customRootDir);
    }

    if (categories && categories.length > 0) {
      setCategory(categories[0].key);
    }

    if (sortBys && sortBys.length > 0) {
      setSortBy(sortBys[0].key);
    }
  }, []);

  return (
    <div className="justify-items-center bg-gray-50 dark:bg-gray-900 h-screen">
      <section className="md:p-5">
        <div className="flex flex-row justify-center px-6 py-8 lg:py-0">
          <div className="w-full rounded-lg bg-white shadow sm:max-w-md md:mt-0 xl:p-0 dark:border dark:border-gray-700 dark:bg-gray-800">
            <div className="space-y-4 p-6 sm:p-8 md:space-y-6">
              <h1 className="text-xl font-bold leading-tight tracking-tight text-gray-900 md:text-2xl dark:text-white">
                {title}
              </h1>
              <form
                className="space-y-4 md:space-y-6"
                onSubmit={(e) => onSubmit(e)}
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

                <div className="flex flex-wrap flex-col justify-around">
                  {sortBys && sortBys.length > 0 && (
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
                        {sortBys.map((sortBy) => {
                          return (
                            <option key={sortBy.key} value={sortBy.key}>
                              {sortBy.name}
                            </option>
                          );
                        })}
                      </select>
                    </div>
                  )}

                  {categories && categories.length > 0 && (
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
                        {categories.map((category) => {
                          return (
                            <option key={category.key} value={category.key}>
                              {category.name}
                            </option>
                          );
                        })}
                      </select>
                    </div>
                  )}

                  {dirs && dirs.length > 1 && (
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
                            setRequestDir(e.target.value);
                          }
                        }}
                      >
                        {dirs.map((dirName) => {
                          return (
                            <option key={dirName} value={dirName}>
                              {dirName}
                            </option>
                          );
                        })}
                      </select>
                    </div>
                  )}
                </div>

                <DirFormComponent
                  setter={setCustomRequestDir}
                  base={customRootDir}
                  name={customRootDir}
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
          torrents={torrents}
          options={{
            baseDir: customRequestDir != "" ? customRequestDir : requestDir,
            url: false,
          }}
        />
      </section>
    </div>
  );
}
