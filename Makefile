PACKAGE=github.com/nimda7/cert_checker
BIN=cert_checker
BUILD_VERSION=$(shell git describe --always --dirty --tags | tr '-' '.' )
BUILD_DATE=$(shell date --iso-8601=minutes)
BUILD_HASH=$(shell git rev-parse HEAD)
BUILD_MACHINE=$(shell echo $$HOSTNAME)
BUILD_USER=$(shell whoami)

BUILD_FLAGS=-ldflags "\
	-X '$(PACKAGE)/cmd.BuildVersion=$(BUILD_VERSION)' \
	-X '$(PACKAGE)/cmd.BuildDate=$(BUILD_DATE)' \
	-X '$(PACKAGE)/cmd.BuildHash=$(BUILD_HASH)' \
"

.PHONY: build test

all: build

run:
	go run main.go

test:
	go test ./...

build: darwin linux windows

darwin:
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BIN)-darwin main.go

linux:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BIN)-linux main.go

windows:
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/$(BIN)-windows main.go
