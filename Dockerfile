# syntax=docker/dockerfile:1 
FROM node:21-alpine as ui

WORKDIR /ui

COPY ui ./

RUN npm install
RUN npm run build 


FROM golang:1.21-alpine as app
# install gcc needed for cgo
RUN apk add gcc musl-dev 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o jvbe ./cmd/jvbe


FROM alpine:3.19

WORKDIR /app

COPY --from=ui /ui/public ./ui/public
COPY --from=ui /ui/views ./ui/views
COPY --from=app /app/jvbe ./

ENTRYPOINT [ "./jvbe" ]
