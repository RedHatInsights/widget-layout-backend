# Testing Guidelines

Concise rules for writing and maintaining tests. For comprehensive examples and patterns, see [TESTING.md](TESTING.md).

## Unique ID System

- Use `test_util.GetUniqueID()` for all database record IDs
- Use `test_util.GetUniqueUserID()` for all user IDs
- Use `test_util.NonExistentID` for 404 test scenarios
- Use `test_util.NoDBTestID` for tests that skip database operations
- Never hardcode numeric IDs or user ID strings in tests

## TestMain Setup

Every test package with database access must have a `TestMain` that:

1. Creates a timestamped SQLite database file (e.g., `{timestamp}-dashboard-template.db`)
2. Resets ID generators: `test_util.ResetIDGenerator()`, `test_util.ResetUserIDGenerator()`
3. Reserves special IDs: `test_util.ReserveID(test_util.NoDBTestID)`, `test_util.ReserveID(test_util.NonExistentID)`
4. Reserves hardcoded user IDs used in legacy tests (e.g., `"user-123"`, `"different-user"`)
5. Cleans up the database file after tests complete

## Test File Organization

- One test file per feature: `get_widgets_test.go`, `update_widget_test.go`, `copy_widget_test.go`
- Common setup functions live in `server_test.go`: `setupRouter()`, `withIdentityContext()`
- Use table-driven tests (`t.Run` subtests) for multiple scenarios of the same operation
- Test all paths: success, not-found, unauthorized, validation failure, edge cases

## Identity in Tests

- `test_util.GenerateIdentityStruct()` creates a default test identity
- `test_util.GenerateIdentityStructFromTemplate()` creates identity with specific user ID
- `withIdentityContext(req)` injects default identity into request context
- `withCustomIdentityContext(req, identity)` injects specific identity
- Always test authorization: verify users cannot access other users' templates

## Mock Data

- `test_util.MockDashboardTemplate()` creates a template with unique IDs
- `test_util.MockDashboardTemplateWithSpecificUser(userID)` creates for a specific user
- `test_util.MockDashboardTemplateWithUniqueID()` creates with guaranteed unique ID

## Test Isolation

- Tests must not depend on execution order
- Tests must not share mutable state
- Each test creates its own data with unique IDs
- Test databases are per-run (timestamped filenames), preventing cross-run interference

## Running Tests

```bash
make test                    # All tests with coverage
go test ./pkg/server -v      # Verbose server tests
go test ./pkg/server -race   # Race condition detection
go test ./pkg/server -count=3  # Reliability check
```

## Before Committing

1. Run `make test` - all tests must pass
2. Check for leftover `.db` files: `find . -name "*-dashboard-template.db"`
3. Remove any debug `t.Logf` or `fmt.Println` added during development
