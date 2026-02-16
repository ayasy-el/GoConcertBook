.PHONY: test build run run-worker token up down k6

test:
	go test ./...

build:
	go build ./...

run:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

token:
	go run ./cmd/token --role=user --sub=user-1 --secret=dev-secret

up:
	docker compose up -d --build

down:
	docker compose down -v

k6:
	k6 run ./k6/reserve.js
