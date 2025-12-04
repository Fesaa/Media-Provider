FROM node:24 AS npm-stage

WORKDIR /app

COPY UI/Web/package.json UI/Web/package-lock.json ./
RUN npm install

COPY UI/Web ./

RUN npm run build


FROM golang:1.25.5 AS go-stage

WORKDIR /app

COPY API/go.mod API/go.sum ./
RUN go mod download

COPY ./API ./

ARG COMMIT_HASH
ARG BUILD_TIMESTAMP

RUN go build -o /media-provider -ldflags "-linkmode external -extldflags '-static' -X github.com/Fesaa/Media-Provider/internal/metadata.CommitHash=${COMMIT_HASH} -X github.com/Fesaa/Media-Provider/internal/metadata.BuildTimestamp=${BUILD_TIMESTAMP}"

FROM alpine:latest

WORKDIR /app

COPY --from=go-stage /media-provider /app/media-provider
COPY --from=npm-stage /app/dist/web/browser /app/public
COPY ./favicon.ico /app/public/favicon.ico
COPY ./API/I18N /app/I18N

RUN apk add --no-cache ca-certificates curl tzdata

ENV CONFIG_DIR="/mp"
ENV DOCKER="true"

HEALTHCHECK CMD curl --fail http://0.0.0.0:8080/health || exit 1

CMD ["./media-provider"]
