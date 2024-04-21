import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import axios from "axios";
import Torrent from "./components/torrentStats";
import { ChevronDoubleRightIcon } from "@heroicons/react/16/solid";
import NotificationHandler from "./notifications/handler";
import Header, { NavigationItem } from "./components/navigation/header";
import { Page } from "./components/form/types";

function Application() {
  const [info, setInfo] = useState({});
  const [navigation, setNavigation] = useState<NavigationItem[]>([]);

  async function loadNavigation() {
    const queryParameters = new URLSearchParams(window.location.search);
    const index = queryParameters.get("index");
    try {
      const res = await axios.get(`${BASE_URL}/api/pages`);
      if (res == null || res.data == null) {
        return;
      }

      const pages: Page[] = res.data;
      let nav = [
        {
          name: "Home",
          href: `${BASE_URL}/`,
          current: index == null,
        },
      ];
      nav.push(
        ...pages.map((page, i) => {
          return {
            name: page.title,
            href: `${BASE_URL}/page?index=${i}`,
            current: String(i) == index,
          };
        }),
      );
      setNavigation(nav);
    } catch (e) {
      console.error(e);
    }
  }

  async function updateInfo(repeat: boolean) {
    var waitLong = true;

    await axios
      .get(`${BASE_URL}/api/stats`)
      .then((res) => {
        if (res == null || res.status != 200) {
          NotificationHandler.addErrorNotificationByTitle(
            "Unable to load stats",
          );
          setInfo({});
          return;
        }

        setInfo(res.data);
        if (Object.keys(res.data).length > 0) {
          waitLong = false;
        }
      })
      .catch((err) => {
        console.log(err);
        NotificationHandler.addErrorNotificationByTitle(
          "Unable to load stats: " + err.message,
        );
      });

    if (repeat) {
      const wait = waitLong ? 10000 : 1000;
      setTimeout(() => updateInfo(repeat), wait);
    }
  }

  useEffect(() => {
    updateInfo(true);
    loadNavigation();
  }, []);

  return (
    <div>
      <Header />
      <main className="bg-gray-50 dark:bg-gray-900">
        <NotificationHandler />
        <section className="pt-5">
          <div className="flex flex-col justify-center items-center p-5 overflow-x-auto">
            {Object.keys(info).length > 0 && (
              <table className="bg-white border border-gray-300 m-2 md:m-10">
                <thead>
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b">
                      Name
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b hidden md:table-cell">
                      Size
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b">
                      Completed
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider border-b"></th>
                  </tr>
                </thead>
                <tbody>
                  {Object.entries(info).map((i: any) => (
                    <Torrent
                      key={i[0]}
                      torrent={i[1]}
                      TKey={i[0]}
                      refreshFunc={updateInfo}
                    />
                  ))}
                </tbody>
              </table>
            )}
          </div>
          {Object.keys(info).length == 0 && (
            <div className="flex flex-col items-center justify-center">
              <h1 className="text-3xl font-bold text-gray-800 dark:text-gray-200">
                No torrents found
              </h1>
              <p className="text-gray-500 dark:text-gray-400">
                Add a torrent to get started
              </p>
              <ul className="flex flex-col justify-start items-start space-y-2 mt-2">
                {navigation
                  .filter((i) => i.href != `${BASE_URL}/`)
                  .map((nav) => (
                    <li
                      key={nav.name}
                      className="flex flex-row items-center justify-center text-center"
                    >
                      <ChevronDoubleRightIcon className="w-4 h-4" />
                      <a
                        href={nav.href}
                        className="text-blue-500 dark:text-blue-400 hover:underline"
                      >
                        {nav.name}
                      </a>
                    </li>
                  ))}
              </ul>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Application />);
