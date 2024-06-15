import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import InfoLine from "./components/torrentStats";
import { ChevronDoubleRightIcon } from "@heroicons/react/16/solid";
import NotificationHandler from "./notifications/handler";
import Header from "./components/navigation/header";
import {getStats, loadNavigation} from "./utils/http";
import {NavigationItem, Stats} from "./utils/types";

declare const BASE_URL: string;

function Application() {
  const [info, setInfo] = useState<Stats>({});
  const [navigation, setNavigation] = useState<NavigationItem[]>([]);

  async function updateInfo(repeat: boolean) {
    let waitLong = true;
    await getStats().then(stats => {
      setInfo(stats);
      if (Object.keys(stats).length > 0) {
        waitLong = false;
      }
    }).catch((err) => {
      NotificationHandler.addErrorNotificationByTitle("Unable to load stats: " + err.message);
    })
    if (repeat) {
      const wait = waitLong ? 10000 : 1000;
      setTimeout(() => updateInfo(repeat), wait);
    }
  }

  useEffect(() => {
    loadNavigation(null).then(setNavigation);
    updateInfo(true);
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
                  {Object.entries(info).map(([key, stat]) => (
                    <InfoLine
                      key={key}
                      infoStat={stat}
                      TKey={key}
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
                No content found
              </h1>
              <p className="text-gray-500 dark:text-gray-400">
                Add a content to get started
              </p>
              <ul className="flex flex-col justify-start items-start space-y-2 mt-2">
                {navigation
                  .filter((i) => i.href != `${BASE_URL}/`)
                  .map((nav) => (
                    <li
                      key={nav.name}
                      className="flex flex-row items-center justify-center text-center space-x-2"
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
