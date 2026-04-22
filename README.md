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

| Document | Description |
|----------|-------------|
| [AGENTS.md](AGENTS.md) | AI agent onboarding: conventions, pitfalls, docs index |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution workflow, commit conventions, PR process |
| [docs/API.md](docs/API.md) | Complete API documentation with endpoints, examples, schemas |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design, data flow, deployment architecture |
| [docs/TESTING.md](docs/TESTING.md) | Comprehensive testing guide with patterns and best practices |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | Widget mapping, base template config, cx/cy coordinate system |
| [docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md) | Identity header generation for local development |
| [docs/WIDGET_MIGRATION.md](docs/WIDGET_MIGRATION.md) | Widget migration procedures |
| [docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md) | Legacy AI agent guidelines |
| [docs/testing-guidelines.md](docs/testing-guidelines.md) | Concise testing rules for agents |
| [docs/api-development-guidelines.md](docs/api-development-guidelines.md) | API development conventions |
| [docs/database-guidelines.md](docs/database-guidelines.md) | GORM patterns and database conventions |

## Grafana Dashboard

A basic Grafana dashboard is defined in `dashboards/grafana-dashboard-insights-widget-layout-backend-general.configmap.yml`. It includes latency and 2xx response rate panels.

To preview locally:

```bash
# Start Grafana
docker run -d -p 3000:3000 --name grafana-local grafana/grafana

# Extract and import the dashboard
python3 -c "
import yaml, json
with open('dashboards/grafana-dashboard-insights-widget-layout-backend-general.configmap.yml') as f:
    cm = yaml.safe_load(f)
dashboard_json = json.loads(cm['data']['general.json'])
print(json.dumps({'dashboard': dashboard_json, 'overwrite': True}))
" > /tmp/grafana-dashboard-import.json

curl -s -X POST http://admin:admin@localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/tmp/grafana-dashboard-import.json

# Open http://localhost:3000/d/widget-layout-backend-general/widget-layout-backend
# Login: admin / admin

# Cleanup
docker stop grafana-local && docker rm grafana-local
```

> **Note:** The local Grafana instance has no Prometheus datasource, so panels will show "No data". This is expected — the purpose is to verify dashboard structure and layout, not live metrics.

## More Information

- See **[docs/API.md](docs/API.md)** for complete API documentation with endpoints, examples, and schemas.
- See the Makefile for available commands.
- See the `cmd/dev/user-identity.go` script for identity header generation logic.
- See **[docs/TESTING.md](docs/TESTING.md)** for comprehensive testing documentation.
- See **[docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md)** for guidelines when using AI-assisted development.
- See **[docs/MCP.md](docs/MCP.md)** - Model Context Protocol integration for AI agents
- See **[mcp/README.md](mcp/README.md)** - MCP sidecar development guide

---

For more details, refer to the documentation files in the [docs/](docs/) directory.
