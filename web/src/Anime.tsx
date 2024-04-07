import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import AnimeForm from "./components/animeForm";

function Search() {
  return (
    <div>
      <NavBar current="Search" />
      <main>
        <AnimeForm />
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
