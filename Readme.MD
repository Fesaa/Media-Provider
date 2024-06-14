# Media-Provider
A simple web applications for downloading torrents. Written in Go, and React, with a very simple env config, meant to run as a docker container.

Torrenting content with copy right is theft, bla bla, this is all educational, bla bla...

## Support

The project currently supports the following providers (websites), to fetch content from. 

- [Lime torrents](limetorrents/Readme.MD)
- [Nyaa](nyaaReadme.MD)
- [YTS](yts/Readme.MD)
- [SubsPlease](subsplease/Readme.MD)
- [Mangadex](mangadex/Readme.MD)

## Config

Is done via `config.yaml`, and must be placed at `/app/config.yaml` when using Docker. Media-Provider works by provider `pages` in the config which will each give you a tab on the website.

Each page has a `title` used as name. And a search config. You can find an example [here](config.yaml.example), and the structs [here](config/impl.go) & [here](config/types.go) if you want to do more.

### Search modifiers
Search modifiers are essentials to how you want to search for content. You can provide filters, etc... per page. You can find information per provider in their respective Readme's. 

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
