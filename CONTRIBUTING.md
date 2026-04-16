# Contributing to Widget Layout Backend

## Prerequisites

- Go 1.24+ (toolchain 1.25.7)
- Docker and Docker Compose (for local PostgreSQL)
- Access to the [OpenAPI spec](spec/openapi.yaml)

## Getting Started

1. Clone the repository
2. Copy `.env.example` to `.env` and configure database credentials
3. Start local infrastructure: `make infra`
4. Run database migrations: `make migrate-db`
5. Start the server: `make dev`

## Development Workflow

### Making API Changes

This project follows **spec-first development**:

1. Edit `spec/openapi.yaml` with your API changes
2. Run `make generate` to regenerate `api/generated.go`
3. Implement or update handler methods in `pkg/server/`
4. Add or update service logic in `pkg/service/`
5. Write tests following the patterns in `pkg/server/*_test.go`
6. Run `make test` to verify all tests pass

### Making Non-API Changes

1. Implement your changes in the appropriate package
2. Write or update tests
3. Run `make test`
4. Run `go vet ./...` for static analysis

## Testing

Run all tests:

```bash
make test
```

Key testing rules:

- Use `test_util.GetUniqueID()` and `test_util.GetUniqueUserID()` for all test data - never hardcode IDs
- Follow one test file per feature (e.g., `get_widgets_test.go`, `update_widget_test.go`)
- Use table-driven tests for multiple similar scenarios
- See [docs/TESTING.md](docs/TESTING.md) for comprehensive patterns

## Commit Conventions

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): short description

TICKET-KEY
Optional body explaining what and why.
```

### Types

- `feat` - new feature or endpoint
- `fix` - bug fix
- `refactor` - code restructuring without behavior change
- `test` - adding or updating tests
- `docs` - documentation changes
- `chore` - maintenance tasks

### Scopes

Use the package or domain: `api`, `server`, `service`, `models`, `config`, `middleware`, `test`.

### Examples

```
feat(api): add dashboard template rename endpoint

RHCLOUD-12345
Add PATCH /{id}/rename endpoint to allow users to rename their dashboard templates.
```

```
fix(service): use map update for boolean default field

RHCLOUD-46426
GORM's struct-based Updates() skips zero-value booleans. Switch to map-based
update to properly unset the default flag on previous templates.
```

## Pull Request Guidelines

- Keep PRs focused on a single change
- Include tests for new functionality
- Run `make test` and `go vet ./...` before opening
- Reference the Jira ticket in the PR description
- If you modified `spec/openapi.yaml`, confirm `make generate` runs cleanly

## Project Structure

See [AGENTS.md](AGENTS.md) for directory structure and architecture overview.
See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for system design details.
