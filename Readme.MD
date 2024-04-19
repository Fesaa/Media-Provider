# Media-Provider
A simple web applications for downloading torrents. Written in Go, and React, with a very simple env config, meant to run as a docker container.

Torrenting content with copy right is theft, bla bla, this is all educational, bla bla...

## Config

Is done via `config.yaml`, the layout can be found in [types.go](config/types.go). And example can be found [here](config.yaml.example)

## Docker compose
This is the entry in my compose file.
```yaml
  media-provider:
    image: media-provider
    restart: "on-failure:3"
    networks:
      - backend
    volumes:
      - ./config.yaml:/app/config.yaml
      - /Media:/media
```
