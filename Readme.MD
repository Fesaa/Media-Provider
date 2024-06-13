# Media-Provider
A simple web applications for downloading torrents. Written in Go, and React, with a very simple env config, meant to run as a docker container.

Torrenting content with copy right is theft, bla bla, this is all educational, bla bla...

## Support

The project currently supports the following providers (websites), to fetch torrents from.

- Lime torrents
- Nyaa
- YTS
- SubsPlease

## Config

Is done via `config.yaml`, the layout can be found in [types.go](config/types.go) & [impl.go](config/impl.go). And example can be found [here](config.yaml.example)

## Docker compose
This is the entry in my compose file.
```yaml
  media-provider:
    image: ameliaah/media-provider:latest-x64
    restart: "on-failure:3"
    healthcheck:
      test: curl --fail http://0.0.0.0/mp/api/health || exit 1
      interval: 60s
      retries: 2
      start_period: 10s
      timeout: 10s
    networks:
      - backend
    volumes:
      - /Media:/media
      - ./mp.yaml:/app/config.yaml
```
