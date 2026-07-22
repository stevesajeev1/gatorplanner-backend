.PHONY: dev lint fmt

dev:
	docker compose up

lint:
	go -C api tool golangci-lint run

fmt:
	go -C api fmt ./...