FROM golang:1.21-alpine as build

WORKDIR /app

COPY main.go go.mod go.sum ./
COPY vendor ./vendor
COPY bot ./bot
RUN go build -mod vendor -o main .

FROM alpine:latest

WORKDIR /app
RUN mkdir data
RUN apk update && apk add tzdata
COPY --from=build /app/main ./

ENTRYPOINT [ "/app/main" ]
