FROM golang:1.20.3-alpine as builder

#RUN apk add --update -t build-deps go git mercurial libc-dev gcc libgcc \
#    && cd $APPPATH && go get -d && go build -o /nsq_exporter \
#    && apk del --purge build-deps && rm -rf $GOPATH

WORKDIR /work
COPY go.mod .
COPY go.sum .
COPY vendor/ vendor/

COPY  ./collector ./collector
COPY *.go .

RUN go build -o nsq_exporter

FROM alpine:3.17.3

EXPOSE 9117

COPY --from=builder /work/nsq_exporter /nsq_exporter

ENTRYPOINT ["/nsq_exporter"]
