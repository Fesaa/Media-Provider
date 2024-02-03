import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import SearchForm from "./components/searchForm";

function Search() {
  return (
    <div>
      <NavBar current="Search" />
      <SearchForm />
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
