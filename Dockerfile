FROM        quay.io/prometheus/busybox:latest
MAINTAINER  Timon Wong <timon86.wang@gmail.com>

COPY        nsq_exporter /bin/nsq_exporter

EXPOSE      9118
ENTRYPOINT  [ "/bin/nsq_exporter" ]
