-include .env
export

.PHONY: lint lint-fix fmt migrate-up migrate-down migrate-create generate

lint:
	go tool golangci-lint run

lint-fix:
	go tool golangci-lint run --fix

fmt:
	go fmt ./...

migrate-up:
	go tool goose up

migrate-down:
	go tool goose down

migrate-create:
	go tool goose create $(name) sql

generate:
	go tool sqlc generate