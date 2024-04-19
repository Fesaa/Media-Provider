import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import NotificationHandler from "./notifications/handler";
import Header from "./components/navigation/header";
import { Page } from "./components/form/types";
import SearchForm from "./components/form/SearchForm";
import axios from "axios";

function PageFunc() {
  const queryParameters = new URLSearchParams(window.location.search);
  const index = queryParameters.get("index");
  if (index == null || index == "") {
    // TODO Redirect to 404
    console.error("No index provided");
    return;
  }

  const [page, setPage] = useState<Page | null>(null);

  useEffect(() => {
    axios
      .get(`${BASE_URL}/api/pages/${index}`)
      .catch((error) => {
        console.error(error);
        // TODO Redirect to 404
      })
      .then((res) => {
        if (res == null || res.data == null) {
          // TODO Redirect to 404
          return;
        }

        setPage(res.data);
      });
  }, []);

  return (
    <div>
      <Header />
      <main>
        <NotificationHandler />
        {page && <SearchForm page={page} />}
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<PageFunc />);
