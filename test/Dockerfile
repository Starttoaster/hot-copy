FROM golang:alpine

ENV APP_PATH=/go/src/hot-copy
ENV HC_PASSWORD="testing"
ENV PUID="1000"
ENV PGID="1000"

RUN mkdir -p $APP_PATH
WORKDIR $APP_PATH
ADD . $APP_PATH
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o hot-copy \
    && mkdir -p /data && mkdir -p /enc-data

CMD go version && go test -cover && go test -v