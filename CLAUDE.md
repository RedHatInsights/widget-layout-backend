@AGENTS.md

## Build & Test Commands

```bash
make build              # Build the binary
make dev                # Run in development mode
make test               # Run all tests with coverage
make generate           # Regenerate api/generated.go from OpenAPI spec
make generate-identity  # Generate x-rh-identity header for local dev
make infra              # Start local PostgreSQL via Docker Compose
make migrate-db         # Run database migrations
```

## Additional Commands

```bash
go test ./pkg/server -v          # Run server tests with verbose output
go test ./pkg/server -race       # Race condition detection
go test ./pkg/server -count=3    # Reliability check (run 3x)
go vet ./...                     # Static analysis
gofmt -w .                       # Format all Go files
go mod tidy                      # Clean up dependencies
```

## Pre-commit Checks

Before committing, always run:

1. `make generate` if you modified `spec/openapi.yaml`
2. `make test` to verify all tests pass
3. `go vet ./...` to catch common issues
