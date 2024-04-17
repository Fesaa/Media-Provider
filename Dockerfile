FROM node:18 AS npm-stage

WORKDIR /app

COPY web/package.json web/package-lock.json ./
RUN npm install

COPY web ./

RUN npm run build:prod
RUN npm run tailwind:prod


FROM golang:1.21 as go-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./api ./api
COPY ./impl ./impl
COPY ./limetorrents ./limetorrents
COPY ./middelware ./middelware
COPY ./models ./models
COPY ./mount ./mount
COPY ./utils ./utils
COPY ./yts ./yts
COPY ./frontend.go ./frontend.go
COPY ./main.go ./

RUN go build -o /media-provider -ldflags '-linkmode external -extldflags "-static"'

FROM alpine:latest

WORKDIR /app

COPY --from=go-stage /media-provider /app/media-provider
COPY --from=npm-stage /app/public/ /app/web/public
COPY --from=npm-stage /app/views/ /app/web/views


RUN apk add --no-cache ca-certificates cifs-utils psmisc

EXPOSE 80

CMD ["./media-provider"]
