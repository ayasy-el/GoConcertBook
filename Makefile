.PHONY: test build run

test:
	go test ./...

build:
	go build ./...

run:
	go run ./cmd/api
