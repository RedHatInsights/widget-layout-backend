# Widget Layout Backend

Welcome to the Widget Layout Backend project!

## Quick Start

- Build: `make build`
- Run in development: `make dev`
- Run tests: `make test`
- Generate identity header for local requests: `make generate-identity`

## Local API Requests & Identity Header

Most endpoints require a valid `x-rh-identity` header. See [DEVELOPMENT_IDENTITY_HEADER.md](DEVELOPMENT_IDENTITY_HEADER.md) for instructions on generating and using this header for local development and testing.

## More Information

- See the Makefile for available commands.
- See the `cmd/dev/user-identity.go` script for identity header generation logic.

---

For more details, refer to the documentation files in this repository.
