PACKAGE := github.com/zondax/rosetta-filecoin-proxy/rosetta/services
REVISION := $(shell git rev-parse --short HEAD)
ROSETTASDKVER := $(shell go list -m all | grep github.com/coinbase/rosetta-sdk-go | awk '{print $$2}')
LOTUSVER := $(shell go list -m all | grep github.com/filecoin-project/lotus | awk '{print $$2}')
RETRYNUM := 10
APPNAME := rosetta-filecoin-proxy

build: 	build_ffi
	go build -ldflags "-X $(PACKAGE).GitRevision=$(REVISION) -X $(PACKAGE).RosettaSDKVersion=$(ROSETTASDKVER) \
 	-X $(PACKAGE).LotusVersion=$(LOTUSVER)" -o $(APPNAME)

build_CI: 	build_ffi
	go build -ldflags "-X $(PACKAGE).GitRevision=$(REVISION) -X $(PACKAGE).RosettaSDKVersion=$(ROSETTASDKVER) \
	-X $(PACKAGE).LotusVersion=$(LOTUSVER) -X $(PACKAGE).RetryConnectAttempts=$(RETRYNUM)" -o $(APPNAME)

clean:
	go clean

build_ffi:
	make -C extern/filecoin-ffi

install_lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.31.0

check-modtidy:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

lint:
	golangci-lint --version
	golangci-lint run -E gofmt -E gosec -E goconst -E gocritic
#   golangci-lint run -E stylecheck -E gosec -E goconst -E godox -E gocritic

test: build
	go test -race ./rosetta/services

test_integration: build
	go test -race ./rosetta/tests

gitclean:
	git clean -xfd
	git submodule foreach --recursive git clean -xfd
