BUILD_DIR = ../build
BINARY_NAME=nsq_exporter

GO       = go
GOX      = gox
GOX_ARGS = -output="$(BUILD_DIR)/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="linux/amd64 linux/386 linux/arm linux/arm64 darwin/amd64 freebsd/amd64 freebsd/386 windows/386 windows/amd64"

.PHONY: build
build:
	$(GO) build -o $(BUILD_DIR)/nsq_exporter .

debug:
	GOOS=linux GOARCH=amd64 go build -gcflags="all=-N -l" -o ${BINARY_NAME}_dlv main.go

dlv:
	dlv --listen=:10499 --headless=true --api-version=2 --accept-multiclient --log exec ./${BINARY_NAME}_dlv -- -nsqd.addr=http://10.16.48.65:4151/stats

.PHONY: deps-init deps-get
deps-init:
	@go get -u github.com/kardianos/govendor
	$(GOPATH)/bin/govendor init

deps-get: deps-init
	@$(GOPATH)/bin/govendor get github.com/lovoo/nsq_exporter

.PHONY: clean
clean:
	rm -R $(BUILD_DIR)/* || true

.PHONY: test
test:
	$(GO) test ./...

.PHONY: release-build
release-build:
	@go get -u github.com/mitchellh/gox
	@$(GOX) $(GOX_ARGS) github.com/lovoo/nsq_exporter
