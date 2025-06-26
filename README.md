# Widget Layout Backend

Welcome to the Widget Layout Backend project!

## Quick Start

- Build: `make build`
- Run in development: `make dev`
- Run tests: `make test`
- Generate identity header for local requests: `make generate-identity`

## Testing

We use a comprehensive testing system with unique ID generation to ensure reliable, conflict-free tests. Key features:

- **Unique ID Generator**: Prevents test conflicts with collision-free ID generation
- **Reserved Constants**: Special IDs for non-existent records and mock scenarios
- **Thread Safety**: Safe for concurrent test execution
- **Test Isolation**: Each test run uses a unique database

### Quick Test Commands

```bash
# Run all tests
make test

# Run specific package tests with verbose output
go test ./pkg/server -v

# Run tests multiple times to check for reliability
go test ./pkg/server -count=3
```

For comprehensive testing documentation, patterns, and best practices, see **[TESTING.md](TESTING.md)**.

## Local API Requests & Identity Header

Most endpoints require a valid `x-rh-identity` header. See [DEVELOPMENT_IDENTITY_HEADER.md](DEVELOPMENT_IDENTITY_HEADER.md) for instructions on generating and using this header for local development and testing.

## More Information

- See the Makefile for available commands.
- See the `cmd/dev/user-identity.go` script for identity header generation logic.
- See **[TESTING.md](TESTING.md)** for comprehensive testing documentation.

---

For more details, refer to the documentation files in this repository.
