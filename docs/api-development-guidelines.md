# API Development Guidelines

## Spec-First Workflow

1. Define or modify endpoints in `spec/openapi.yaml`
2. Run `make generate` to regenerate `api/generated.go`
3. Implement the `ServerInterface` method in `pkg/server/server.go`
4. Add business logic in `pkg/service/`
5. Write tests in `pkg/server/*_test.go`

Never edit `api/generated.go` directly - it is gitignored and regenerated from the spec.

## Handler Structure

All handlers in `pkg/server/` follow this pattern:

```go
func (Server) OperationName(w http.ResponseWriter, r *http.Request, params ...) {
    w.Header().Set("Content-Type", "application/json")
    id := middlewares.GetUserIdentity(r.Context())

    resp, status, err := service.DoSomething(id, params)
    if err != nil {
        logrus.Errorf("Failed to do something: %v", err)
        w.WriteHeader(status)
        _ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
            {Code: status, Message: err.Error()},
        }})
        return
    }

    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(resp)
}
```

Rules:
- Set `Content-Type` header before writing status
- Extract identity at the start (except public endpoints like `/widget-mapping`)
- Service functions return `(response, statusCode, error)` - use the status they provide
- Log errors with `logrus.Errorf` before writing response
- Use `_` for encoder errors on error paths (already writing error response)

## Error Response Format

All errors use `api.ErrorResponse` with `api.ErrorPayload`:

```json
{"errors": [{"code": 404, "message": "Dashboard template not found"}]}
```

Use HTTP status codes from the service layer - don't override them in handlers.

## Service Layer Conventions

- Service functions live in `pkg/service/`
- Return signature: `(responseType, int, error)` where int is HTTP status code
- Always validate user ownership before modifying templates
- Use GORM for all database operations (never raw SQL)
- Access base templates via `BaseTemplateRegistry.GetBase(name)` / `GetAllBases()`
- Access widget mappings via `WidgetMappingRegistry.GetWidgetMapping(key)` / `GetAllWidgetMappings()`

## List Response Format

List endpoints use a wrapper with metadata:

```go
listResponse := api.DashboardTemplateListResponse{
    Data: items,
    Meta: api.ListResponseMeta{Count: len(items)},
}
```

Exception: `GET /widget-mapping` returns `api.WidgetMappingResponse{Data: mappings}` with a key-value object instead of an array.

## Adding a New Endpoint

1. Add the path and operation to `spec/openapi.yaml`
2. Define request/response schemas in `components/schemas`
3. Run `make generate`
4. Add handler method to `Server` struct in `pkg/server/server.go`
5. Add service function in `pkg/service/`
6. Add test file `pkg/server/{operation}_test.go`
7. Update `docs/API.md` with the new endpoint documentation

## OpenAPI Spec Conventions

- Use `operationId` in camelCase (maps to Go method name)
- Reference shared schemas via `$ref`
- Include error responses (400, 403, 404, 500) for each endpoint
- Path parameters use `int64` format for template IDs
- Query parameters are optional unless explicitly required

## Request Validation

The `oapi-codegen/nethttp-middleware` validates requests against the OpenAPI spec automatically. Invalid requests get rejected before reaching handlers. Custom validation (e.g., empty strings, business rules) happens in handlers or service functions.

## Authentication

- `middlewares.InjectUserIdentity` is applied via `api.ChiServerOptions.Middlewares`
- It decodes the base64 `x-rh-identity` header using `platform-go-middlewares/v2/identity`
- Identity is stored in context under `config.IdentityContextKey`
- `GetUserIdentity(ctx)` panics if identity is missing - only call it in authenticated endpoints
