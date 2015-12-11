# NSQ Exporter

[![GoDoc](https://godoc.org/github.com/lovoo/nsq_exporter?status.svg)](https://godoc.org/github.com/lovoo/nsq_exporter)

NSQ exporter for prometheus.io, written in go.

## Usage

    docker run -d --name nsq_exporter -l nsqd:nsqd -p 9117:9117 lovoo/nsq_exporter:latest -nsq.addr=http://nsqd:4151 -collectors=nsqstats

## Building

    go install github.com/lovoo/nsq_exporter

## TODO

* collect all nsqd instances over nsqlookupd

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request
