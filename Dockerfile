#Build image
FROM golang:latest AS builder

ENV APP_PATH=/go/src/hot-copy

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH

ADD . $APP_PATH
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o hot-copy


#End image
FROM alpine
LABEL maintainer="Brandon Butler bmbawb@gmail.com"

RUN mkdir -p /data && mkdir -p /inside

VOLUME /inside
VOLUME /data

COPY --from=builder /go/src/hot-copy/hot-copy /hot-copy

ENTRYPOINT ["/hot-copy"]