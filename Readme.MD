# Media-Provider
A simple web applications for downloading torrents. Written in Go, and React, with a very simple env config, meant to run as a docker container.

Torrenting content with copy right is theft, bla bla, this is all educational, bla bla...

## Config

All these should be set with a environment variable

General config

| Key | Description | Default |
| ---- | ----- | ---- |
| PORT | port | 80 |
| PASSWORD | login pass | admin |
| TORRENT_DIR | Root dir | temp |
| BASE_URL | service url from root | / |

Features

| Key | Description | Default |
| --- | ----- | ---- |
| anime | Nyaa anime search tab* | enabled |
| manga | Nyaa manga search tab* | enabled |
| movies | YTS search tab* | enabled |
| lime | Lime search tab | enabled |

\* Wants a dir with the same name from root. (anime => TORRENT_DIR/Anime). Currently not changeable

## Docker compose
This is the entry in my compose file. You can disable 
```yaml
  media-provider:
    image: media-provider
    restart: "on-failure:3"
    networks:
      - backend
    environment:
      - "TORRENT_DIR=/media"
      - "BASE_URL=/mp"
    volumes:
      - /Media:/media

```