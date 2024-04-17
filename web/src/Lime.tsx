import React from "react";
import { createRoot } from "react-dom/client";
import NavBar from "./components/navbar";
import LimeForm from "./components/LimeForm";
import NotificationHandler from "./notifications/handler";

function Search() {
  return (
    <div>
      <NavBar current="Lime" />
      <main>
        <NotificationHandler />
        <LimeForm />
      </main>
    </div>
  );
}

const container = document.getElementById("application");
const root = createRoot(container!);
root.render(<Search />);
