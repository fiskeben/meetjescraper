meetjescraper: $(shell find . -name '*.go')
	go build

linux: $(shell find . -name '*.go')
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o meetjescraper-linux

test:
	go test
	
clean:
	rm meetjescraper

.PHONY: clean test
