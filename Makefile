BUILD_DIR = build

GO       = go
GOX      = gox
GOX_ARGS = -output="$(BUILD_DIR)/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="linux/amd64 linux/386 linux/arm linux/arm64 darwin/amd64 freebsd/amd64 freebsd/386 windows/386 windows/amd64"

.PHONY: build
build:
	$(GO) build -o $(BUILD_DIR)/nsq_exporter .

.PHONY: deps-init deps-get
deps-init:
	-go mod init
	go mod vendor

deps-get: deps-init
	go get -u ./...
	go mod tidy
	go mod vendor

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
