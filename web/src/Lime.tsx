import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import LimeForm from "./components/LimeForm";

function Search() {
  return (
    <div>
      <NavBar current="Lime" />
      <main>
        <LimeForm />
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
