import React from "react";
import { createRoot } from "react-dom/client";

declare const BASE_URL: string;

function Status404() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-200">
      <div className="text-center">
        <h1 className="text-4xl font-medium">404</h1>
        <p className="m-6 text-xl font-medium">
          Sorry, the page you're looking for can't be found.
        </p>
        <a
          href={`${BASE_URL}/`}
          className="rounded bg-blue-500 px-4 py-2 text-white hover:bg-blue-600"
        >
          Go Home
        </a>
      </div>
    </div>
  );
}

const container = document.getElementById("page");
const root = createRoot(container!);
root.render(<Status404 />);
