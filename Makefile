.PHONY: build dev generate test infra migrate-db help build-mcp dev-mcp test-mcp lint-mcp

help:
	@echo "Available commands:"
	@echo "  build                Build the widget-layout-backend binary"
	@echo "  dev                  Run the application in development mode"
	@echo "  generate             Run go generate on all packages"
	@echo "  generate-identity    Generate a user identity for development"
	@echo "  test                 Run all unit tests with coverage"
	@echo "  infra                Start local infrastructure with Docker Compose"
	@echo "  migrate-db           Run the database migration script"
	@echo "  build-mcp            Build the MCP sidecar Docker image"
	@echo "  dev-mcp              Run MCP sidecar in development mode"
	@echo "  test-mcp             Run MCP sidecar tests"
	@echo "  lint-mcp             Lint MCP sidecar TypeScript code"
	@echo "  help                 Show this help message"

build:
	go build -o bin/widget-layout-backend .

dev:
	go run .

generate:
	go generate ./...

test:
	go test -coverprofile=coverage.out ./... -v
	@go tool cover -html=coverage.out -o coverage.html

infra:
	docker-compose --env-file .env -f local/database-compose.yaml up

migrate-db:
	go run cmd/database/migrate.go

generate-identity:
	go run cmd/dev/user-identity.go

# MCP Sidecar targets
build-mcp:
	cd mcp && docker build -t widget-layout-mcp:latest .

dev-mcp:
	cd mcp && npm run dev

test-mcp:
	cd mcp && npm test

lint-mcp:
	cd mcp && npm run lint
