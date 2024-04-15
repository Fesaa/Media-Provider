import React from "react";
import NyaaForm, { CategoryRecord } from "./nyaaForm";

const cats: CategoryRecord[] = [
  { key: "all", value: "All categories" },
  { key: "anime", value: "Anime" },
  { key: "anime-amv", value: "Music Video" },
  { key: "anime-eng", value: "English Translated" },
  { key: "anime-non-eng", value: "Non-English Translated" },
  { key: "anime-raw", value: "Raw" },
  { key: "audio", value: "Audio" },
  { key: "audio-lossless", value: "Lossless" },
  { key: "audio-lossy", value: "Lossy" },
];

export default function AnimeForm() {
  return <NyaaForm baseDir="Anime" title="Anime" categories={cats} />;
}
