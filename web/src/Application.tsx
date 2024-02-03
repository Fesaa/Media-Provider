import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";

function Application() {
  return (
    <div>
      <NavBar current="Home" />
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Application />);
