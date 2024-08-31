# syntax=docker/dockerfile:1

FROM node:21-alpine as ui

WORKDIR /ui

COPY ui ./

RUN npm install
RUN npm run build 



FROM golang:1.23-bookworm as app

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o jvbe ./cmd/jvbe


FROM debian:bookworm

WORKDIR /app

COPY --from=ui /ui/public ./ui/public
COPY --from=ui /ui/views ./ui/views
COPY --from=app /app/jvbe ./

ENTRYPOINT [ "./jvbe", "-e" ]
