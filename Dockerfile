FROM golang:1.21.6-alpine as build

ADD . /build
WORKDIR /build

RUN cd src && go build -mod vendor -o /build/main .

FROM alpine:latest

ARG REVISION=unknown
ENV REVISION=${REVISION}

WORKDIR /app
RUN apk update && apk add tzdata
COPY --from=build /build/main ./

ENTRYPOINT [ "/app/main" ]
