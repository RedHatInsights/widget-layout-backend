# Widget Layout Backend

Welcome to the Widget Layout Backend project!

## Quick Start

### Main Application

- Build: `make build`
- Run in development: `make dev`
- Run tests: `make test`
- Generate identity header for local requests: `make generate-identity`

### MCP Sidecar

- Build Docker image: `make build-mcp`
- Run in development: `make dev-mcp`
- Run tests: `make test-mcp`
- Lint TypeScript code: `make lint-mcp`

## Testing

We use a comprehensive testing system with unique ID generation to ensure reliable, conflict-free tests. Key features:

- **Focused Test Files**: We prefer multiple smaller, focused test files over large monolithic ones for better maintainability and organization
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

For comprehensive testing documentation, patterns, and best practices, see **[docs/TESTING.md](docs/TESTING.md)**.

## Local API Requests & Identity Header

Most endpoints require a valid `x-rh-identity` header. See [docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md) for instructions on generating and using this header for local development and testing.

## MCP (Model Context Protocol) Integration

This service includes a TypeScript-based MCP sidecar container that enables AI agents to interact with widget dashboard data via a standardized JSON-RPC 2.0 protocol. The sidecar runs alongside the main Go application and provides read-only access to dashboard templates, base templates, and widget mappings.

**Key Features:**
- 6 read-only tools for AI agents
- Standard JSON-RPC 2.0 protocol
- Authentication via x-rh-identity header
- Prometheus metrics and structured logging
- Full test coverage

See **[docs/MCP.md](docs/MCP.md)** for complete MCP documentation and usage examples.

See **[mcp/README.md](mcp/README.md)** for development setup and technical details.

## Documentation

All project documentation is organized in the `docs/` folder:

- **[docs/API.md](docs/API.md)** - Complete API documentation with endpoints, examples, and schemas
- **[docs/TESTING.md](docs/TESTING.md)** - Comprehensive testing guide with patterns and best practices
- **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)** - Widget mapping and base template configuration guide
- **[docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md)** - Identity header generation for local development
- **[docs/MCP.md](docs/MCP.md)** - Model Context Protocol integration for AI agents
- **[docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md)** - Guidelines for AI-assisted development
- **[mcp/README.md](mcp/README.md)** - MCP sidecar development guide

## More Information

- See **[docs/API.md](docs/API.md)** for complete API documentation with endpoints, examples, and schemas.
- See the Makefile for available commands.
- See the `cmd/dev/user-identity.go` script for identity header generation logic.
- See **[docs/TESTING.md](docs/TESTING.md)** for comprehensive testing documentation.
- See **[docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md)** for guidelines when using AI-assisted development.

---

For more details, refer to the documentation files in the [docs/](docs/) directory.
