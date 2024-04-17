import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import MoviesForm from "./components/moviesForm";
import NotificationHandler from "./notifications/handler";

function Search() {
  return (
    <div>
      <NavBar current="Movies" />
      <main>
        <NotificationHandler />
        <MoviesForm />
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
