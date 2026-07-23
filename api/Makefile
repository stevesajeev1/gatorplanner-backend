-include .env
export

.PHONY: lint fmt migrate-up migrate-down generate

lint:
	go tool golangci-lint run

fmt:
	go fmt ./...

migrate-up:
	go tool goose up

migrate-down:
	go tool goose down

generate:
	go tool sqlc generate