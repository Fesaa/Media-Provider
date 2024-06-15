import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import NotificationHandler from "./notifications/handler";
import Header from "./components/navigation/header";
import { Page } from "./components/form/types";
import SearchForm from "./components/form/SearchForm";
import {getPage} from "./utils/http";

declare const BASE_URL: string;

function PageFunc() {
  const queryParameters = new URLSearchParams(window.location.search);
  const index = queryParameters.get("index");
  if (index == null || index == "") {
    window.location.href = `${BASE_URL}/404`
    return;
  }

  const [page, setPage] = useState<Page | null>(null);

  useEffect(() => {
    getPage(parseInt(index)).then(setPage)
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
