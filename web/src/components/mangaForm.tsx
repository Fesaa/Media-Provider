import React from "react";
import NyaaForm, { CategoryRecord } from "./nyaaForm";

const cats: CategoryRecord[] = [
  { key: "literature", value: "Literature" },
  { key: "literature-eng", value: "English Literature" },
  { key: "literature-non-eng", value: "Non English Literature" },
  { key: "literature-raw", value: "Raw Literature" },
];

export default function MangaForm() {
  return <NyaaForm baseDir="Manga" title="Manga" categories={cats} />;
}
