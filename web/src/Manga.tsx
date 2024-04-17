import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import MangaForm from "./components/mangaForm";
import NotificationHandler from "./notifications/handler";

function Search() {
  return (
    <div>
      <NavBar current="Search" />
      <main>
        <NotificationHandler />
        <MangaForm />
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
