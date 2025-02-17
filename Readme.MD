# Media-Provider

[![Validate - Tests](https://github.com/Fesaa/Media-Provider/actions/workflows/tests.yml/badge.svg)](https://github.com/Fesaa/Media-Provider/actions/workflows/tests.yml) [![Validate - Linter](https://github.com/Fesaa/Media-Provider/actions/workflows/linting.yml/badge.svg)](https://github.com/Fesaa/Media-Provider/actions/workflows/linting.yml)

DOCS: https://fesaa.github.io/Media-Provider/

A web application for downloading content. Written in Go, and Angular. With a simple to navigate interface.
Set up once, no need to touch files again. 

![Preview](docs/assets/demo.gif)


Torrenting content with copy right is theft, bla bla, this is all educational, bla bla...

## Docker compose

Is it recommended to use the docker image, no other way is officially supported.
The image is available on [Docker Hub](https://hub.docker.com/r/ameliaah/media-provider). 
Here is a simple example;

```yaml
media-provider:
  image: ameliaah/media-provider:latest
  restart: "on-failure:3"
  ports:
    - "8080:8080"
  volumes:
    - ./mp-data:/mp
```

Or with docker run, 

`docker run -v ./mp-data:/mp -p 8080:8080 ameliaah/media-provider:latest`
