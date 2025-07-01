# AI Agent Context for Widget Layout Backend

This document provides essential context and guidelines for LLM agents (like Claude, GPT, etc.) working on the Widget Layout Backend codebase. It ensures consistent, high-quality AI-assisted development practices.

## Core Principles

### 1. Testing First Philosophy
- **Always prioritize testing** - Every code change should be accompanied by appropriate tests
- **Use focused test files** - Prefer multiple smaller, focused test files over large monolithic ones
- **Leverage unique ID generators** - Use `test_util.GetUniqueID()` and `test_util.GetUniqueUserID()` to prevent test conflicts
- **Verify test isolation** - Ensure tests don't interfere with each other and can run in any order

### 2. Documentation Accuracy
- **Code-Documentation Sync** - Documentation MUST accurately reflect the actual code implementation
- **Verify before documenting** - Always check the current code state before updating documentation
- **Update all relevant docs** - When changing functionality, update README.md, docs/TESTING.md, and any other relevant documentation
- **Include examples** - Provide working code examples in documentation that match the actual implementation

### 3. Cleanup and Maintenance
- **Clean up temporary files** - Remove any temporary scripts, test files, or helper files created during development
- **No leftover artifacts** - Don't leave debug code, console.log statements, or temporary variables in production code
- **Proper git hygiene** - Ensure commits are clean and focused, with appropriate commit messages
- **Remove unused imports** - Clean up any unused imports or dependencies introduced during development

## Testing Guidelines

### Required Testing Practices

1. **Use Unique ID Generators**
   ```go
   // ✅ Correct - prevents conflicts
   templateID := test_util.GetUniqueID()
   userID := test_util.GetUniqueUserID()
   
   // ❌ Incorrect - causes conflicts
   templateID := uint(123)
   userID := "hardcoded-user"
   ```

2. **Follow Test File Organization**
   - One test file per major functionality (e.g., `get_widgets_test.go`, `update_widget_test.go`)
   - Keep common setup functions in `server_test.go`
   - Use descriptive test names that explain the scenario being tested

3. **Verify Test Isolation**
   ```go
   func TestMain(m *testing.M) {
       // Always reset generators before tests
       test_util.ResetIDGenerator()
       test_util.ResetUserIDGenerator()
       
       // Reserve special IDs to prevent conflicts
       test_util.ReserveID(test_util.NoDBTestID)
       test_util.ReserveID(test_util.NonExistentID)
       
       // ... rest of setup
   }
   ```

4. **Test All Scenarios**
   - Success cases with valid data
   - Error cases (not found, unauthorized, validation failures)
   - Edge cases (empty data, boundary conditions)
   - Authorization scenarios (different users, permissions)

### Test File Structure Requirements

Each test file should:
- Import only necessary packages
- Use the established helper functions (`setupRouter`, `withIdentityContext`, etc.)
- Follow the naming convention: `{functionality}_test.go`
- Include comprehensive test cases for the specific functionality
- Use table-driven tests where appropriate for multiple similar scenarios

## Documentation Standards

### Code-Documentation Verification Process

Before updating any documentation:

1. **Read the actual code** - Use code reading tools to understand current implementation
2. **Verify examples work** - Test any code examples provided in documentation
3. **Check for recent changes** - Look for recent commits or PRs that might affect documentation
4. **Update comprehensively** - Don't just update one section, check all related documentation

### Documentation Files to Maintain

All project documentation is organized in the `docs/` folder for better organization and clarity:

- **README.md** - High-level project overview, quick start, testing philosophy (root level)
- **docs/TESTING.md** - Comprehensive testing guide with working examples
- **docs/CONFIGURATION.md** - Widget mapping and base template configuration guide
- **docs/DEVELOPMENT_IDENTITY_HEADER.md** - Identity header generation and usage
- **docs/AI_AGENT_CONTEXT.md** - This file, for AI agents working on the codebase

### Documentation Organization

- Keep only **README.md** at the root level for immediate project overview
- All other documentation lives in **docs/** for organized access
- Update links when referencing documentation (use `docs/` prefix)
- Maintain the documentation index in README.md

### Documentation Quality Checklist

- [ ] All code examples are tested and working
- [ ] Function signatures match actual implementation
- [ ] File paths and structure are accurate
- [ ] Best practices reflect current codebase patterns
- [ ] Examples use the unique ID generator system
- [ ] Cross-references between documents are valid

## Code Quality Standards

### Go-Specific Requirements

1. **Follow Go Conventions**
   - Use `gofmt` for formatting
   - Follow Go naming conventions
   - Use proper error handling patterns
   - Include appropriate comments for exported functions

2. **Import Management**
   ```go
   // ✅ Correct - organized imports
   import (
       "context"
       "fmt"
       "net/http"
       
       "github.com/RedHatInsights/widget-layout-backend/pkg/config"
       "github.com/go-chi/chi/v5"
   )
   
   // ❌ Incorrect - unused imports
   import (
       "context"
       "fmt"
       "net/http"
       "unused/package"  // Remove this
   )
   ```

3. **Error Handling**
   - Always handle errors appropriately
   - Use meaningful error messages
   - Don't ignore errors with `_` unless explicitly justified

### Database and API Patterns

1. **Use Established Patterns**
   - Follow existing API response patterns
   - Use the established database connection patterns
   - Maintain consistency with existing error handling

2. **Security Considerations**
   - Always validate user authorization
   - Use the identity middleware properly
   - Sanitize inputs and validate data

## Cleanup Procedures

### Before Completing Any Task

1. **Remove Temporary Files**
   ```bash
   # Check for temporary files that might have been created
   find . -name "*.tmp" -o -name "temp_*" -o -name "debug_*"
   
   # Remove any test database files that weren't cleaned up
   find . -name "*-dashboard-template.db"
   ```

2. **Clean Up Code**
   - Remove debug print statements
   - Remove commented-out code
   - Remove unused variables and functions
   - Ensure proper error handling

3. **Verify Tests Pass**
   ```bash
   # Run all tests to ensure nothing is broken
   make test
   
   # Run tests multiple times to check for flaky behavior
   go test ./pkg/server -count=3
   ```

4. **Check Documentation Accuracy**
   - Verify all examples in documentation work
   - Ensure function signatures match implementation
   - Check that file paths and structure are correct

### Git Hygiene

1. **Commit Messages**
   - Use clear, descriptive commit messages
   - Follow conventional commit format when possible
   - Reference issues or PRs when relevant

2. **Commit Content**
   - Each commit should be focused on a single change
   - Don't include unrelated changes in the same commit
   - Ensure commits don't break tests

## Common Pitfalls to Avoid

### Testing Pitfalls

1. **Hardcoded IDs** - Always use the unique ID generator system
2. **Test Dependencies** - Tests should not depend on execution order
3. **Shared State** - Tests should not share mutable state
4. **Incomplete Cleanup** - Always clean up test data and temporary files

### Documentation Pitfalls

1. **Outdated Examples** - Code examples that don't match current implementation
2. **Wrong File Paths** - References to files that don't exist or have moved
3. **Incomplete Updates** - Updating one part of documentation but missing related sections
4. **Untested Examples** - Providing code examples that haven't been verified to work

### Code Quality Pitfalls

1. **Unused Imports** - Leaving imports that are no longer needed
2. **Debug Code** - Leaving debug prints or temporary code in production
3. **Inconsistent Patterns** - Not following established codebase patterns
4. **Poor Error Handling** - Ignoring errors or using generic error messages

## Verification Checklist

Before considering any task complete:

- [ ] All tests pass (`make test`)
- [ ] Tests are properly isolated and use unique ID generators
- [ ] Documentation accurately reflects code implementation
- [ ] All code examples in documentation have been tested
- [ ] No temporary files or debug code remains
- [ ] Imports are clean and necessary
- [ ] Error handling follows established patterns
- [ ] Git commits are clean and focused
- [ ] Related documentation files are updated consistently

## Tools and Commands

### Essential Commands
```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/server -v

# Check for race conditions
go test ./pkg/server -race

# Format code
gofmt -w .

# Build project
make build

# Run in development mode
make dev
```

### Debugging and Verification
```bash
# Check for unused imports
go mod tidy

# Verify test coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Check for potential issues
go vet ./...
```

## Summary

This document ensures that AI agents working on the Widget Layout Backend maintain high standards for:

- **Testing**: Comprehensive, isolated, and reliable tests using the unique ID generator system
- **Documentation**: Accurate, up-to-date documentation that reflects actual code implementation
- **Code Quality**: Clean, well-organized code following Go best practices
- **Maintenance**: Proper cleanup of temporary files and maintaining git hygiene

By following these guidelines, AI agents can contribute effectively to the codebase while maintaining its quality and reliability. 