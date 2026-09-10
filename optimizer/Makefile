.PHONY: lint lint-fix fmt generate

lint:
	uv run ruff check .

lint-fix:
	uv run ruff check . --fix

fmt:
	uv run ruff format .

generate:
	uv run python -m grpc_tools.protoc -I proto --python_betterproto2_out=src/generated proto/*.proto --python_betterproto2_opt=client_generation=none --python_betterproto2_opt=server_generation=async