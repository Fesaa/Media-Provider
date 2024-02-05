import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import axios from "axios";
import Torrent from "./components/torrentStats";
import { ArrowPathRoundedSquareIcon } from "@heroicons/react/24/outline";

function Application() {
  const [info, setInfo] = useState({});

  async function updateInfo() {
    axios
      .get("/api/stats")
      .then((res) => setInfo(res.data))
      .catch((err) => console.error(err));
  }

  useEffect(() => {
    updateInfo();
  }, []);

  return (
    <div className="bg-gray-50 dark:bg-gray-900">
      <NavBar current="Home" />
      <section className="pt-5">
        <div className="mx-10 flex flex-col px-6 py-8 md:h-screen lg:py-0">
          <div
            onClick={updateInfo}
            className="mx-auto flex flex-row gap-4 rounded bg-blue-200 p-4 dark:bg-white"
          >
            <ArrowPathRoundedSquareIcon
              className="h-6 w-6 text-green-600"
              aria-hidden="true"
            />
            Update torrent info
          </div>

          <div className="m-5 flex flex-row flex-wrap gap-5">
            {Object.entries(info).map((i: any) => (
              <Torrent key={i[1].Infohash} torrent={i[1]} TKey={i[0]} />
            ))}
          </div>
        </div>
      </section>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Application />);
