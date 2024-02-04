FROM node:18 AS npm-stage

WORKDIR /app

COPY web/package.json web/package-lock.json ./
RUN npm install

COPY web ./

RUN npm run build:prod 
RUN npm run tailwind:prod


FROM golang:1.21 as go-stage

WORKDIR /app

COPY . ./

RUN go mod download
RUN go build -o /media-provider

FROM ubuntu:latest

WORKDIR /app

COPY --from=go-stage /media-provider /app/media-provider
COPY --from=go-stage /app/mount.sh /app/mount.sh
COPY --from=npm-stage /app/public/ /app/web/public
COPY --from=npm-stage /app/views/ /app/web/views


RUN apt-get update && apt-get install -y ca-certificates && apt install -y cifs-utils psmisc
RUN mkdir /app/mount
RUN chmod +x /app/mount.sh

EXPOSE 80

CMD ["sh", "-c", "./mount.sh && ./media-provider"]
