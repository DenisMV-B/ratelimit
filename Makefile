.PHONY: build
build:
	go build -v .

.PHONY: test
test:
	go test -v -race -timeout 30s ./...

.PHONY: clean
clean:
	go clean


.DEFAULT_GOAL := build
