#!/bin/bash

mkdir temp
mkdir temp/Anime
mkdir temp/LightNovels
mkdir temp/Series
mkdir temp/Manga
mkdir temp/Movies

EXAMPLE_CONFIG_FILE="config.yaml.example"
CONFIG_FILE="config.yaml"

if [ ! -e "$CONFIG_FILE"]; then
  echo "Copying example config"
  cp EXAMPLE_CONFIG_FILE CONFIG_FILE
else
  echo "Skipping config copy"
fi