# Testing Guide

This document provides comprehensive information about testing in the Widget Layout Backend project.

## Quick Start

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/server -v

# Run tests multiple times to check for flaky behavior
go test ./pkg/server -count=3

# Run tests with race detection
go test ./pkg/server -race
```

## Test Structure

The project uses Go's built-in testing framework with the following structure:

- `pkg/server/server_test.go` - Main API endpoint tests
- `pkg/models/*_test.go` - Database model tests
- `pkg/middlewares/identity_test.go` - Middleware tests
- `pkg/test_util/` - Test utilities and helpers

## Unique ID Generator

To prevent test conflicts and ensure reliable test execution, we use a unique ID generator system.

### Overview

The unique ID generator (`pkg/test_util/unique_id_generator.go`) provides:

- **Collision-free IDs**: Ensures no two tests use the same ID
- **Thread safety**: Safe for concurrent test execution
- **Reserved IDs**: Special constants for specific test scenarios
- **Reset functionality**: Clean state between test runs

## User ID Generator

For preventing user-specific conflicts in tests, we also provide a unique user ID generator.

### Overview

The unique user ID generator (also in `pkg/test_util/unique_id_generator.go`) provides:

- **Collision-free User IDs**: Ensures no two tests use the same user ID
- **Thread safety**: Safe for concurrent test execution
- **Consistent format**: Generates user IDs in the format "test-user-{counter}"
- **Reset functionality**: Clean state between test runs

### Reserved ID Constants

```go
const (
    // NonExistentID is used for testing scenarios with IDs that don't exist in the database
    NonExistentID uint = 999999
    
    // NoDBTestID is used for testing scenarios where database operations are mocked or skipped
    NoDBTestID uint = 123456
)
```

### Usage Examples

#### Getting Unique IDs for DB Records

```go
func TestCreateTemplate(t *testing.T) {
    // Get a unique ID for this test
    templateID := test_util.GetUniqueID()
    // Get a unique user ID for this test
    userID := test_util.GetUniqueUserID()
    
    template := api.DashboardTemplate{
        ID:     templateID,
        UserId: userID,
        // ... other fields
    }
    
    result := database.DB.Create(&template)
    assert.NoError(t, result.Error)
}
```

#### Using Reserved IDs for Special Cases

```go
func TestNonExistentTemplate(t *testing.T) {
    server := setupRouter()
    
    // Use the reserved constant for non-existent ID tests
    req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", test_util.NonExistentID), nil)
    req = withIdentityContext(req)
    w := httptest.NewRecorder()
    
    server.GetWidgetLayoutById(w, req, int64(test_util.NonExistentID))
    
    assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateWithoutDB(t *testing.T) {
    server := setupRouter()
    
    // Use the reserved constant for tests that don't create DB records
    templateUpdate := `{"templateConfig": {"lg": [...]}}`
    
    req, _ := http.NewRequest("PATCH", fmt.Sprintf("/%d", test_util.NoDBTestID), 
                             strings.NewReader(templateUpdate))
    req = withIdentityContext(req)
    w := httptest.NewRecorder()
    
    server.UpdateWidgetLayoutById(w, req, int64(test_util.NoDBTestID))
    
    // Test validation logic without DB dependencies
}
```

#### Using Unique User IDs

```go
func TestUserAuthorization(t *testing.T) {
    server := setupRouter()
    
    // Create templates for different users
    userID1 := test_util.GetUniqueUserID()
    userID2 := test_util.GetUniqueUserID()
    
    template1 := api.DashboardTemplate{
        ID:     uint(test_util.GetUniqueID()),
        UserId: userID1,
        // ... other fields
    }
    template2 := api.DashboardTemplate{
        ID:     uint(test_util.GetUniqueID()),
        UserId: userID2,
        // ... other fields
    }
    
    // Create templates in database
    database.DB.Create(&template1)
    database.DB.Create(&template2)
    
    // Test that user1 can only access their own template
    req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", template1.ID), nil)
    req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
        xrhidgen.Identity{},
        xrhidgen.User{UserID: stringPtr(userID1)},
        xrhidgen.Entitlements{},
    ))
    w := httptest.NewRecorder()
    
    server.GetWidgetLayoutById(w, req, int64(template1.ID))
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Test Setup and Cleanup

#### TestMain Pattern

```go
func TestMain(m *testing.M) {
    // ... database setup ...
    
    // Reset the unique ID generators for clean tests
    test_util.ResetIDGenerator()
    test_util.ResetUserIDGenerator()
    
    // Reserve hardcoded IDs used in special test scenarios
    test_util.ReserveID(test_util.NoDBTestID)
    test_util.ReserveID(test_util.NonExistentID)
    
    // Reserve commonly used user IDs to prevent conflicts
    test_util.ReserveUserID("user-123")       // Used in some existing tests
    test_util.ReserveUserID("different-user") // Used in authorization tests
    
    exitCode := m.Run()
    
    // ... cleanup ...
    os.Exit(exitCode)
}
```

### Best Practices

#### ✅ Do

- Use `test_util.GetUniqueID()` for all database record IDs
- Use `test_util.GetUniqueUserID()` for all user IDs in tests
- Use reserved constants (`NonExistentID`, `NoDBTestID`) for special test cases
- Reset both ID generators in `TestMain` before running tests
- Reserve special IDs and user IDs in `TestMain` to prevent conflicts

#### ❌ Don't

- Use hardcoded IDs or user IDs in tests (except reserved constants)
- Reuse IDs or user IDs across different test cases
- Forget to reset the generators between test runs
- Use magic numbers or hardcoded strings for test identifiers

### Helper Functions

The test utilities provide several helper functions:

```go
// Generate unique IDs
templateID := test_util.GetUniqueID()
userID := test_util.GetUniqueUserID()

// Generate a unique dashboard template with random ID
template := test_util.MockDashboardTemplate()

// Generate a template with a specific user ID
template := test_util.MockDashboardTemplateWithSpecificUser("specific-user-123")

// Generate a template with guaranteed unique ID (though MockDashboardTemplate already uses unique IDs)
template := test_util.MockDashboardTemplateWithUniqueID()

// Generate test identity for authentication
identity := test_util.GenerateIdentityStruct()

// Generate identity with specific user ID
identity := test_util.GenerateIdentityStructFromTemplate(
    xrhidgen.Identity{},
    xrhidgen.User{UserID: stringPtr(userID)},
    xrhidgen.Entitlements{},
)

// Generate identity header string for HTTP requests
headerValue := test_util.GenerateIdentityHeader()
```

## Test Categories

### Unit Tests

Test individual functions and methods in isolation:

```go
func TestValidateTemplateConfig(t *testing.T) {
    config := api.DashboardTemplateConfig{...}
    err := validateTemplateConfig(config)
    assert.NoError(t, err)
}
```

### Integration Tests

Test API endpoints with database interactions:

```go
func TestCreateTemplate(t *testing.T) {
    server := setupRouter()
    templateID := test_util.GetUniqueID()
    
    // Create template via API
    // Verify in database
    // Test response
}
```

### Authentication Tests

Test identity middleware and authorization:

```go
func TestUnauthorizedAccess(t *testing.T) {
    // Test without identity header
    // Test with invalid identity
    // Test with different user's resources
}
```

## Database Testing

### Test Database Isolation

Each test run uses a unique database file to prevent conflicts:

```go
func TestMain(m *testing.M) {
    now := time.Now().UnixNano()
    dbName := fmt.Sprintf("%d-dashboard-template.db", now)
    cfg.DatabaseConfig.DBName = dbName
    
    // ... test execution ...
    
    // Cleanup
    os.Remove(dbName)
}
```

### Model Testing

Test database models and relationships:

```go
func TestDashboardTemplateModel(t *testing.T) {
    template := &models.DashboardTemplate{
        ID:     test_util.GetUniqueID(),
        UserId: "test-user",
        // ... other fields
    }
    
    result := database.DB.Create(template)
    assert.NoError(t, result.Error)
    
    // Test retrieval, updates, deletion
}
```

## Mock Data and Fixtures

### Using Test Utilities

```go
// Generate mock dashboard template
template := test_util.MockDashboardTemplate()

// Generate mock identity
identity := test_util.GenerateIdentityStruct()

// Create template with specific user
template := test_util.MockDashboardTemplateWithSpecificUser("user-123")
```

### Custom Test Data

```go
func createTestWidget() api.WidgetItem {
    return api.WidgetItem{
        Height:     2,
        Width:      4,
        X:          0,
        Y:          0,
        WidgetType: "test-widget",
        Static:     false,
        Title:      "Test Widget",
        MaxHeight:  6,
        MinHeight:  1,
    }
}

func createTestTemplateWithUser(userID string) api.DashboardTemplate {
    template := test_util.MockDashboardTemplate()
    template.UserId = userID
    return template
}
```

## Debugging Tests

### Verbose Output

```bash
# Run tests with verbose output
go test ./pkg/server -v

# Run specific test
go test ./pkg/server -v -run TestGetWidgetLayoutById
```

### Test Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Race Detection

```bash
# Check for race conditions
go test ./... -race
```

## Common Patterns

### API Test Setup

```go
func TestAPIEndpoint(t *testing.T) {
    t.Run("success case", func(t *testing.T) {
        server := setupRouter()
        
        // Create test data with unique ID
        templateID := test_util.GetUniqueID()
        testTemplate := api.DashboardTemplate{
            ID:     templateID,
            UserId: "user-123",
            // ... other fields
        }
        database.DB.Create(&testTemplate)
        
        // Make request
        req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", templateID), nil)
        req = withIdentityContext(req)
        w := httptest.NewRecorder()
        
        server.GetWidgetLayoutById(w, req, int64(templateID))
        
        // Verify response
        assert.Equal(t, http.StatusOK, w.Code)
        
        var response api.DashboardTemplate
        json.NewDecoder(w.Body).Decode(&response)
        assert.Equal(t, testTemplate.ID, response.ID)
    })
}
```

### Error Case Testing

```go
func TestErrorCases(t *testing.T) {
    t.Run("not found", func(t *testing.T) {
        server := setupRouter()
        
        req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", test_util.NonExistentID), nil)
        req = withIdentityContext(req)
        w := httptest.NewRecorder()
        
        server.GetWidgetLayoutById(w, req, int64(test_util.NonExistentID))
        
        assert.Equal(t, http.StatusNotFound, w.Code)
        
        var errorResponse api.ErrorResponse
        json.NewDecoder(w.Body).Decode(&errorResponse)
        assert.NotEmpty(t, errorResponse.Errors)
    })
}
```

## Troubleshooting

### Common Issues

1. **UNIQUE constraint failed**: Use `test_util.GetUniqueID()` instead of hardcoded IDs
2. **Flaky tests**: Ensure proper test isolation and unique test data
3. **Database conflicts**: Check that TestMain properly cleans up database files
4. **Authentication failures**: Verify identity context is properly set

### Test Debugging

```go
// Add debug output in tests
t.Logf("Generated template ID: %d", templateID)
t.Logf("Response body: %s", w.Body.String())

// Use testify assertions for better error messages
assert.Equal(t, expected, actual, "Descriptive error message")
```

## Contributing

When adding new tests:

1. Use the unique ID generator for all database records
2. Follow existing patterns for test structure
3. Add appropriate error case testing
4. Ensure tests are isolated and don't depend on each other
5. Update this documentation if adding new testing utilities

For more information about the project structure, see the main [README.md](README.md).
