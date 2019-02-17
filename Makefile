version := $(shell git describe --tags --always)

meetjescraper-darwin-amd64: $(shell find . -name '*.go')
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(version)" -o meetjescraper-darwin-amd64

meetjescraper-linux-amd64: $(shell find . -name '*.go')
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(version)" -o meetjescraper-linux-amd64

all: meetjescraper-darwin-amd64 meetjescraper-linux-amd64

release:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(version)" -o meetjescraper-darwin-amd64-$(version)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(version)" -o meetjescraper-linux-amd64-$(version)

test:
	go test
	
clean:
	rm meetjescraper-darwin-amd64
	rm meetjescraper-linux-amd64

.PHONY: clean test all
