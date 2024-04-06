import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import axios from "axios";
import Torrent from "./components/torrentStats";

function Application() {
  const [info, setInfo] = useState({});

  async function updateInfo(repeat: boolean) {
    axios
      .get("/api/stats")
      .then((res) => setInfo(res.data))
      .catch((err) => console.error(err));

    if (repeat) {
      const wait = Object.keys(info).length > 0 ? 10000 : 1000;
      setTimeout(updateInfo, wait);
    }
  }

  useEffect(() => {
    updateInfo(true);
  }, []);

  return (
    <div className="bg-gray-50 dark:bg-gray-900">
      <NavBar current="Home" />
      <section className="pt-5">
        <div className="mx-10 flex flex-col px-6 py-8 md:h-screen lg:py-0">
          <div className="m-5 flex flex-row flex-wrap gap-5">
            {Object.entries(info).map((i: any) => (
              <div key={i[1].Infohash}>
                <Torrent torrent={i[1]} TKey={i[0]} refreshFunc={updateInfo} />
              </div>
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
