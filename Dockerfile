FROM alpine:latest

EXPOSE 9117

ENV  GOPATH /go
ENV APPPATH $GOPATH/src/github.com/lovoo/nsq_exporter
COPY . $APPPATH
RUN apk add --update -t build-deps go git mercurial libc-dev gcc libgcc \
    && cd $APPPATH && go get -d && go build -o /nsq_exporter \
    && apk del --purge build-deps && rm -rf $GOPATH

ENTRYPOINT ["/nsq_exporter"]
