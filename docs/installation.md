# Installation

Docker is the only support way to install and use Media provider. You can build the binary yourself if needed.

## Docker compose
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
    - /path/to/media:/my-media 
```

Or with docker run,

`docker run -v ./mp-data:/mp -p 8080:8080 ameliaah/media-provider:latest`
