import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import MoviesForm from "./components/moviesForm";

function Search() {
  return (
    <div>
      <NavBar current="Search" />
      <MoviesForm />
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
