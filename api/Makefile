-include .env
export

.PHONY: lint lint-fix lint-proto fmt fmt-proto migrate-up migrate-down migrate-create generate-sql generate-proto test

lint:
	go tool golangci-lint run

lint-fix:
	go tool golangci-lint run --fix

lint-proto:
	go tool buf lint

fmt:
	go fmt ./...

fmt-proto:
	go tool buf format -w

migrate-up:
	go tool goose up

migrate-down:
	go tool goose down

migrate-create:
	go tool goose create $(name) sql

generate-sql:
	go tool sqlc generate

generate-proto:
	go tool buf generate

test:
	go test ./...