.PHONY: build dev generate

help:
	@echo "Available commands:"
	@echo "  build     Build the widget-layout-backend binary"
	@echo "  dev       Run the application in development mode"
	@echo "  generate  Run go generate on all packages"
	@echo "  test  		 Run all unit tests with coverage"
	@echo "  help      Show this help message"

build:
	go build -o bin/widget-layout-backend .

dev:
	go run .

generate:
	go generate ./...

test:
	go test -coverprofile=coverage.out ./... $(ARGS)
	@go tool cover -html=coverage.out -o coverage.html