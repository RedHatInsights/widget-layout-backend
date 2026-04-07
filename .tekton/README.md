# Tekton Pipelines

This directory contains Konflux CI/CD pipelines for building and testing the widget-layout-backend application.

## Pipelines

### Main Application (Go)

- **widget-layout-backend-cfb93-push.yaml** - Builds and pushes the main Go API on master branch
- **widget-layout-backend-cfb93-pull-request.yaml** - PR validation for main Go API

### MCP Sidecar (TypeScript)

- **mcp-sidecar-push.yaml** - Builds and pushes the MCP sidecar on master branch
- **mcp-sidecar-pull-request.yaml** - PR validation for MCP sidecar with linting and testing

## MCP Sidecar Pipeline Details

The MCP sidecar pipelines include comprehensive testing and validation:

### Pull Request Pipeline

Triggered on: Pull requests to `master` branch

**Steps:**
1. **Clone Repository** - Clone source code
2. **Install Dependencies** - Run `npm ci` in mcp/ directory
3. **Lint Code** - Run ESLint (`npm run lint`)
4. **Type Check** - Run TypeScript compiler (`npm run typecheck`)
5. **Run Tests** - Execute Jest tests with coverage (`npm test -- --ci --coverage`)
6. **Build Image** - Build Docker image using `mcp/Dockerfile`
7. **Push Image** - Push to quay.io with `on-pr-*` tag (expires in 5 days)

**Test Output:**
- All 28 tests must pass
- Coverage report generated
- Test artifacts stored as trusted artifacts

### Push Pipeline

Triggered on: Pushes to `master` branch

**Steps:**
1. **Clone Repository** - Clone source code
2. **Install Dependencies** - Run `npm ci`
3. **Run Tests** - Execute Jest tests with coverage
4. **Build Image** - Build production Docker image
5. **Push Image** - Push to quay.io with commit SHA tag

## Image Locations

**MCP Sidecar Images:**
- Production: `quay.io/redhat-user-workloads/hcc-platex-services-tenant/mcp-sidecar:<commit-sha>`
- PR Builds: `quay.io/redhat-user-workloads/hcc-platex-services-tenant/mcp-sidecar:on-pr-<commit-sha>`

**Main Application Images:**
- Production: `quay.io/redhat-user-workloads/hcc-platex-services-tenant/widget-layout-backend-cfb93:<commit-sha>`

## Local Testing

Before pushing changes, run tests locally:

```bash
# Test main application
make test

# Test MCP sidecar
make test-mcp

# Lint MCP sidecar
make lint-mcp

# Type check MCP sidecar
cd mcp && npm run typecheck
```

## Pipeline Features

Both pipelines use Konflux's trusted artifacts pattern:
- ✅ OCI-based artifact storage
- ✅ SBOM generation
- ✅ Security scanning
- ✅ Image signing
- ✅ Hermetic builds support

## Test Requirements

**MCP Sidecar:**
- All 28 Jest tests must pass
- No ESLint errors
- No TypeScript compilation errors
- Coverage threshold: 80% (branches, functions, lines, statements)

**Coverage Breakdown:**
- 7 test suites
- 28 individual tests
- Test files:
  - `tests/tools/hello.test.ts`
  - `tests/tools/get-layouts.test.ts`
  - `tests/tools/get-layout-by-id.test.ts`
  - `tests/tools/get-base-templates.test.ts`
  - `tests/tools/get-widget-mapping.test.ts`
  - `tests/tools/export-layout.test.ts`
  - `tests/utils/identity.test.ts`

## Troubleshooting

### Pipeline Failures

**Tests Failing:**
```bash
# Run tests locally to debug
cd mcp
npm test -- --verbose
```

**Linting Errors:**
```bash
# Check and fix linting issues
cd mcp
npm run lint
npm run lint:fix
```

**Type Errors:**
```bash
# Check TypeScript types
cd mcp
npm run typecheck
```

### Image Pull Issues

If the pipeline-built image can't be pulled:
1. Check quay.io repository permissions
2. Verify image tag exists
3. Check Konflux build logs

## Adding New Tests

When adding new MCP tools:

1. Create test file in `mcp/tests/tools/<tool-name>.test.ts`
2. Follow existing test patterns:
   - Test successful execution
   - Test authentication (if required)
   - Test error handling
   - Test parameter validation
3. Run tests locally: `npm test`
4. Ensure coverage thresholds are met

## Reference

- [Konflux Documentation](https://konflux-ci.dev/)
- [Tekton Pipelines](https://tekton.dev/)
- [Trusted Artifacts](https://konflux-ci.dev/architecture/ADR/0036-trusted-artifacts.html)
