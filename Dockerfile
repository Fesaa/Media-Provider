FROM node:18 AS npm-stage

WORKDIR /app

COPY UI/Web/package.json UI/Web/package-lock.json ./
RUN npm install

COPY UI/Web ./

RUN npm run build


FROM golang:1.24.1 AS go-stage

WORKDIR /app

RUN apt update && apt install -y libwebp-dev

COPY go.mod go.sum ./
RUN go mod download

COPY ./api ./api
COPY ./comicinfo ./comicinfo
COPY ./config ./config
COPY ./db ./db
COPY ./http ./http
COPY ./metadata ./metadata
COPY ./providers ./providers
COPY ./services ./services
COPY ./utils ./utils
COPY ./*.go ./

ARG COMMIT_HASH
ARG BUILD_TIMESTAMP

RUN go build -o /media-provider -ldflags "-linkmode external -extldflags '-static' -X github.com/Fesaa/Media-Provider/metadata.CommitHash=${COMMIT_HASH} -X github.com/Fesaa/Media-Provider/metadata.BuildTimestamp=${BUILD_TIMESTAMP}"

FROM alpine:latest

WORKDIR /app

COPY --from=go-stage /media-provider /app/media-provider
COPY --from=npm-stage /app/dist/web/browser /app/public
COPY ./I18N /app/I18N

RUN apk add --no-cache ca-certificates curl tzdata libwebp

ENV CONFIG_DIR="/mp"
ENV DOCKER="true"

HEALTHCHECK CMD curl --fail http://0.0.0.0:8080/health || exit 1

CMD ["./media-provider"]
