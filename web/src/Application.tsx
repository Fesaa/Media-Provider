import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import NavBar, { navigation } from "./components/navbar";
import axios from "axios";
import Torrent from "./components/torrentStats";
import { ChevronDoubleRightIcon } from "@heroicons/react/16/solid";

function Application() {
  const [info, setInfo] = useState({});

  async function updateInfo(repeat: boolean) {
    axios
      .get(`${BASE_URL}/api/stats`)
      .then((res) => setInfo(res.data))
      .catch((err) => console.error(err));

    if (repeat) {
      const wait = Object.keys(info).length > 0 ? 10000 : 1000;
      setTimeout(() => updateInfo(repeat), wait);
    }
  }

  useEffect(() => {
    updateInfo(true);
  }, []);

  return (
    <div>
      <NavBar current="Home" />
      <main className="bg-gray-50 dark:bg-gray-900">
        <section className="pt-5">
          <div className="mx-10 flex flex-col px-6 py-8 lg:py-0">
            <div className="m-5 flex flex-row flex-wrap gap-5">
              {Object.entries(info).map((i: any) => (
                <div key={i[1].Infohash}>
                  <Torrent
                    torrent={i[1]}
                    TKey={i[0]}
                    refreshFunc={updateInfo}
                  />
                </div>
              ))}
            </div>
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
