APPNAME := gmailai
VERSION := 0.1.0
GOARCH := amd64

.PHONY: all clean

all: clean test build

build:
	GOOS=linux GOARCH=$(GOARCH) go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(APPNAME)-linux-$(GOARCH) cmd/main.go
	GOOS=darwin GOARCH=$(GOARCH) go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(APPNAME)-macos-$(GOARCH) cmd/main.go
	GOOS=windows GOARCH=$(GOARCH) go build -ldflags "-s -w -X main.Version=$(VERSION)" -o bin/$(APPNAME)-windows-$(GOARCH).exe cmd/main.go

clean:
	rm -f bin/*

test:
	go test -v ./...