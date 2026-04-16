# Database Guidelines

## ORM: GORM

All database access goes through GORM. Never use raw SQL unless GORM cannot express the query.

## Model Definition

Models are type aliases to API types defined in `api/`:

```go
// pkg/models/DashboardTemplate.go
type DashboardTemplate = api.DashboardTemplate
```

This means the API schema (generated from OpenAPI spec) directly defines the database schema. Field changes in the OpenAPI spec affect both the API and database.

## Database Initialization

`pkg/database/database.go` initializes GORM with PostgreSQL (production) or SQLite (tests/local). The `database.DB` global variable is the shared GORM handle.

### Production (Clowder)

PostgreSQL connection string built from Clowder config with SSL support, connection pooling defaults:
- Max idle: 10, Max open: 150, Max lifetime: 5 min
- Configurable via `DB_MAX_IDLE_CONNS`, `DB_MAX_OPEN_CONNS`, `DB_CONN_MAX_LIFETIME_MINUTES`

### Local Development

PostgreSQL via Docker Compose (`make infra`), configured through `.env` file with `PGSQL_*` vars.

### Tests

SQLite with timestamped filenames (`{timestamp}-dashboard-template.db`) for isolation. Cleaned up in `TestMain` teardown.

## GORM Patterns

### Creating Records

```go
result := database.DB.Create(&template)
if result.Error != nil {
    return nil, http.StatusInternalServerError, result.Error
}
```

### Querying with User Scope

Always filter by user ID for user-owned resources:

```go
var templates []models.DashboardTemplate
result := database.DB.Where("user_id = ?", userID).Find(&templates)
```

### Updating Records

For struct-based updates (non-zero values only):

```go
database.DB.Model(&template).Updates(models.DashboardTemplate{
    TemplateConfig: newConfig,
})
```

For zero-value updates (booleans, integers, empty strings), use a map:

```go
database.DB.Model(&template).Where("user_id = ?", userID).
    Updates(map[string]interface{}{"default": false})
```

This is critical - GORM's struct-based `Updates()` silently skips zero-value fields.

### Deleting Records

```go
result := database.DB.Delete(&models.DashboardTemplate{}, templateID)
```

## Ownership Validation

Before any mutation, verify the template belongs to the requesting user:

```go
var template models.DashboardTemplate
result := database.DB.First(&template, templateID)
if result.Error != nil {
    return nil, http.StatusNotFound, errors.New("template not found")
}
if template.UserId != userIdentity.Identity.User.UserID {
    return nil, http.StatusForbidden, errors.New("unauthorized")
}
```

## Migrations

Run via `make migrate-db` which executes `cmd/database/migrate.go`. GORM's `AutoMigrate` handles schema updates.

## Auto-Creation Pattern

When querying templates by `dashboardType` and none exist for the user, the service automatically forks the matching base template. This creates a new DB record and returns it with a 404 status to signal to the frontend that the template was just created.

## Testing Database Operations

- Use `test_util.GetUniqueID()` for record IDs to prevent UNIQUE constraint violations
- Each test run uses an isolated SQLite file
- Create test data directly via `database.DB.Create()` in test setup
- Clean up database files in `TestMain` defer
