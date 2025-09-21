# Installation Guide

Media-Provider is also provided to you as a container, on [Docker Hub](https://hub.docker.com/repository/docker/ameliaah/media-provider/).
You may run the container however you want, docker, podman, kubernetes...

### Docker compose
The easiest way for most people will be to use docker compose. Below is a minimal example
```yaml
services:
  media-provider:
    image: ameliaah/media-provider:latest
    ports:
      - "8080:8080"
    environment:
      - "TZ=Your/Timezone"
    volumes:
      - /path/to/config:/mp # Must be /mp unless changed with env vars
      - /path/to/media:/media # /media can be whatever you want 
```

### Env variables
Most of the config can, and should, be done via the UI. However, some more advanced settings (called features) are set
via environment variables. A full list with explanation can be found on [GitHub](https://github.com/Fesaa/Media-Provider/blob/main/API/config/features.go).