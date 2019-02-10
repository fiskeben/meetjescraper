meetjescraper: $(shell find . -name '*.go')
	go test
	go build

test:
	go test

clean:
	rm meetjescraper

.PHONY: clean test
