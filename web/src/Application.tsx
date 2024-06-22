import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import InfoLine from "./components/infoLine";
import { ChevronDoubleRightIcon } from "@heroicons/react/16/solid";
import NotificationHandler from "./notifications/handler";
import Header from "./components/navigation/header";
import {getStats, loadNavigation} from "./utils/http";
import {InfoStat, NavigationItem, QueueStat, SpeedData, Stats} from "./utils/types";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';
import QueueLine from "./components/queueLine";

declare const BASE_URL: string;

ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Legend
);

function Application() {
  const [info, setInfo] = useState<Stats>({queued: [], running: []});
  const [navigation, setNavigation] = useState<NavigationItem[]>([]);
  const [speedData, setSpeedDate] = useState<{[key: string]: SpeedData[]}>({});

  async function updateInfo(repeat: boolean): Promise<void> {
    let waitLong = true;
    await getStats().then(stats => {
      setInfo(stats);

      const curData = speedData;
      stats.running.forEach(data => {
        curData[data.id] = [...(curData[data.id] || []), data.speed];
      })
      setSpeedDate(curData);

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

  const comparator = (a: InfoStat, b: InfoStat) => {
    return a.name.localeCompare(b.name);
  }
  const comparatorQueue = (a: QueueStat, b: QueueStat) => {
    return a.name.localeCompare(b.name);
  }

  return (
    <div>
      <Header />
      <main className="bg-gray-50 dark:bg-gray-900">
        <NotificationHandler />
        <section className="pt-5">
          {info.running.length > 0 && <div className="flex flex-col flex-grow py-5 px-5 md:mx-20 space-y-2">
            <span className="font-bold text-xl">Downloading</span>
            {info.running.sort(comparator).map(stat => (
                <InfoLine
                    key={stat.id}
                    infoStat={stat}
                    speeds={speedData[stat.id] || []}
                    TKey={stat.id}
                    refreshFunc={updateInfo}
                />
            ))}
          </div>}
          {info.queued.length > 0 && <div className="flex flex-col flex-grow py-5 px-5 md:mx-20 space-y-2">
            <span className="font-bold text-xl">Queue</span>
            {info.queued.sort(comparatorQueue).map(stat => (
                <QueueLine stat={stat} refreshFunc={updateInfo} key={stat.id}/>
            ))}
          </div>}
          {info.running.length == 0 && (
              <div className="flex flex-col items-center justify-center">
                <h1 className="text-3xl font-bold text-gray-800 dark:text-gray-200">
                  No content found
                </h1>
                <p className="text-gray-500 dark:text-gray-400">
                  Add content to get started
                </p>
                <ul className="flex flex-col justify-start items-start space-y-2 mt-2">
                  {navigation
                      .filter((i) => i.href != `${BASE_URL}/`)
                      .map((nav) => (
                          <li
                              key={nav.name}
                              className="flex flex-row items-center justify-center text-center space-x-2"
                          >
                            <ChevronDoubleRightIcon className="w-4 h-4"/>
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
root.render(<Application/>);
