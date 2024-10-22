FROM node:18 AS npm-stage

WORKDIR /app

COPY UI/Web/package.json UI/Web/package-lock.json ./
RUN npx update-browserslist-db@latest
RUN npm install

COPY UI/Web ./

RUN npm run build


FROM golang:1.22.2 AS go-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./api ./api
COPY ./auth ./auth
COPY ./comicinfo ./comicinfo
COPY ./config ./config
COPY ./limetorrents ./limetorrents
COPY ./log ./log
COPY ./mangadex ./mangadex
COPY ./payload ./payload
COPY ./providers ./providers
COPY ./subsplease ./subsplease
COPY ./utils ./utils
COPY ./webtoon ./webtoon
COPY ./wisewolf ./wisewolf
COPY ./yoitsu ./yoitsu
COPY ./yts ./yts
COPY ./*.go ./

RUN go build -o /media-provider -ldflags '-linkmode external -extldflags "-static"'

FROM alpine:latest

WORKDIR /app

COPY --from=go-stage /media-provider /app/media-provider
COPY --from=npm-stage /app/dist/web/browser /app/public

RUN apk add --no-cache ca-certificates curl

ENV CONFIG_DIR="/mp/"
ENV DOCKER="true"

HEALTHCHECK CMD curl --fail http://0.0.0.0:8080/health || exit 1

CMD ["./media-provider"]
