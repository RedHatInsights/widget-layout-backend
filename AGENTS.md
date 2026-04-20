# Widget Layout Backend - Agent Guide

## Project Overview

Go backend service for the Hybrid Cloud Console (HCC) widget dashboard system. Manages user-customizable dashboard layouts with responsive widget grids. Users create, fork, copy, reset, and export dashboard templates. Widget metadata and base templates are loaded from Kubernetes ConfigMaps at startup.

**API prefix**: `/api/widget-layout/v1`

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go 1.24+ (toolchain 1.25.7) |
| Router | chi v5 |
| ORM | GORM (PostgreSQL in prod, SQLite in tests) |
| Code generation | oapi-codegen v2 (OpenAPI 3.0 spec-first) |
| Auth | x-rh-identity header (platform-go-middlewares) |
| Metrics | Prometheus client_golang |
| Logging | logrus |
| Config | Clowder (app-common-go) / .env for local |
| Deployment | ClowdApp on OpenShift |

## Directory Structure

```
spec/openapi.yaml          # Source of truth for the API
server.cfg.yaml            # oapi-codegen config
api/generated.go           # Generated (gitignored) - run `make generate`
api/common.go              # Custom UnmarshalJSON, validation, WidgetMapping types
api/BaseWidgetDashboardTemplate.go  # Base template types
pkg/server/                # HTTP handlers (implement generated ServerInterface)
pkg/service/               # Business logic + in-memory registries
pkg/models/                # GORM models (type alias to api types)
pkg/database/              # GORM DB initialization
pkg/config/                # Env var / Clowder config loading
pkg/middlewares/            # Identity extraction middleware
pkg/test_util/             # Unique ID generators, mock factories, identity helpers
cmd/database/migrate.go    # DB migration entry point
cmd/dev/user-identity.go   # Dev tool: generate x-rh-identity header
deploy/clowdapp.yaml       # Kubernetes deployment manifest
local/                     # Docker Compose for local PostgreSQL
```

## Cross-Cutting Conventions

### Spec-First API Development

1. Edit `spec/openapi.yaml` first
2. Run `make generate` to regenerate `api/generated.go`
3. Implement or update handler methods in `pkg/server/`
4. `api/generated.go` is gitignored - never edit or commit it

### Error Response Pattern

All handlers use a consistent error response format:

```go
w.WriteHeader(status)
_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
    {Code: status, Message: err.Error()},
}})
```

### Identity Handling

- All endpoints except `/widget-mapping` require `x-rh-identity` header
- `middlewares.InjectUserIdentity` decodes the header into context
- `middlewares.GetUserIdentity(ctx)` retrieves identity (panics if missing)
- Templates are scoped per user - always filter by user ID from identity

### Coordinate System: cx/cy vs x/y

This is a critical repo-specific detail:

- **Configuration files** (ConfigMaps, JSON env vars): use `cx`/`cy`
- **REST API** (requests/responses, runtime objects): use `x`/`y`
- Reason: YAML parser treats bare `y` as boolean `true`
- `api/common.go` `UnmarshalJSON` handles the automatic conversion
- Never mix formats - config always `cx`/`cy`, API always `x`/`y`

### GORM Zero-Value Pitfall

GORM's `.Updates()` with a struct silently skips zero-value fields (`false`, `0`, `""`). When setting a field to its zero value, use a map:

```go
// Wrong - won't unset the boolean
tx.Model(&template).Updates(DashboardTemplate{Default: false})

// Correct - explicitly sets the boolean to false
tx.Model(&template).Updates(map[string]interface{}{"default": false})
```

### In-Memory Registries

Base templates and widget mappings are loaded from env vars at startup via `init()` functions in `pkg/service/`. They are stored in thread-safe registries (`BaseTemplateRegistry`, `WidgetMappingRegistry`) and never persisted to the database. Invalid config causes fatal shutdown.

### Testing with Unique IDs

All test database records must use the unique ID generator system from `pkg/test_util/`:

- `test_util.GetUniqueID()` for record IDs
- `test_util.GetUniqueUserID()` for user IDs
- `test_util.NonExistentID` / `test_util.NoDBTestID` for special cases
- Reset generators in `TestMain` before test execution

### Handler Pattern

Handlers in `pkg/server/` follow a consistent pattern:

1. Set `Content-Type: application/json`
2. Extract user identity from context
3. Call service layer function
4. On error: log with logrus, write error response
5. On success: write status + encode JSON response

## Common Pitfalls

- **Forgetting `make generate`** after editing `spec/openapi.yaml` - build will fail
- **Hardcoded test IDs** cause UNIQUE constraint failures - always use `test_util.GetUniqueID()`
- **Using `x`/`y` in config JSON** - will fail YAML parsing in production
- **GORM zero-value updates** - boolean/int fields need map-based updates
- **Missing identity header** - `GetUserIdentity` panics if identity not in context
- **Test database cleanup** - each test run creates a timestamped `.db` file; `TestMain` must clean up

## Documentation Index

| Document | Description |
|----------|-------------|
| [docs/API.md](docs/API.md) | Complete REST API documentation with endpoints, examples, schemas |
| [docs/TESTING.md](docs/TESTING.md) | Comprehensive testing guide with patterns and unique ID system |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | Widget mapping, base template config, cx/cy coordinate system |
| [docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md) | Identity header generation for local development |
| [docs/WIDGET_MIGRATION.md](docs/WIDGET_MIGRATION.md) | Widget migration procedures |
| [docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md) | Legacy AI agent guidelines |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | System design, data flow, deployment architecture |
| [docs/testing-guidelines.md](docs/testing-guidelines.md) | Concise testing rules for agents |
| [docs/api-development-guidelines.md](docs/api-development-guidelines.md) | API development conventions |
| [docs/database-guidelines.md](docs/database-guidelines.md) | GORM patterns and database conventions |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Contribution workflow, commit conventions, PR process |
