# Widget Layout Backend API Documentation

## Overview

The Widget Layout Backend provides a RESTful API for managing dashboard templates and widget configurations. This API enables users to create, read, update, and delete personalized dashboard layouts, as well as manage base templates and widget mappings.

**Base URL**: `/api/widget-layout/v1`

## Authentication

All endpoints (except `/widget-mapping`) require a valid `x-rh-identity` header containing user identity information. For local development and testing, use the identity header generator:

```bash
make generate-identity
```

For detailed instructions on generating and using identity headers, see [docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md).

## Common Response Formats

### Success Response
All successful responses return the requested data with appropriate HTTP status codes (200, 201, 204).

### Error Response
```json
{
  "errors": [
    {
      "code": 404,
      "message": "Dashboard template not found"
    }
  ]
}
```

## API Endpoints

### Dashboard Templates

Dashboard templates are user-specific widget layouts that define how widgets are arranged on a dashboard across different screen sizes (sm, md, lg, xl).

#### GET `/`
Get all dashboard templates for the authenticated user.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "userId": "user-123",
    "createdAt": "2024-01-01T12:00:00Z",
    "updatedAt": "2024-01-01T12:00:00Z",
    "templateConfig": {
      "sm": [
        {
          "w": 2,
          "h": 2,
          "x": 0,
          "y": 0,
          "i": "widget1",
          "static": false,
          "maxH": 4,
          "minH": 1
        }
      ],
      "md": [...],
      "lg": [...],
      "xl": [...]
    },
    "templateBase": {
      "name": "custom-dashboard-template",
      "displayName": "My Custom Dashboard"
    },
    "default": true
  }
]
```

**Error Responses:**
- `500` - Internal server error

#### GET `/{dashboardTemplateId}`
Get a specific dashboard template by ID.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/1' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "id": 1,
  "userId": "user-123",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T12:00:00Z",
  "templateConfig": {
    "sm": [
      {
        "w": 2,
        "h": 2,
        "x": 0,
        "y": 0,
        "i": "widget1",
        "static": false,
        "maxH": 4,
        "minH": 1
      }
    ],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "custom-dashboard-template",
    "displayName": "My Custom Dashboard"
  },
  "default": false
}
```

**Error Responses:**
- `403` - Unauthorized access (template belongs to different user)
- `404` - Dashboard template not found
- `500` - Internal server error

#### PATCH `/{dashboardTemplateId}`
Update a specific dashboard template.

**Request:**
```bash
curl -X PATCH \
  'http://localhost:8080/api/widget-layout/v1/1' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...' \
  -H 'Content-Type: application/json' \
  -d '{
    "templateConfig": {
      "sm": [
        {
          "w": 3,
          "h": 2,
          "x": 0,
          "y": 0,
          "i": "widget1",
          "static": false,
          "maxH": 4,
          "minH": 1
        }
      ],
      "md": [...],
      "lg": [...],
      "xl": [...]
    }
  }'
```

**Response (200 OK):**
```json
{
  "id": 1,
  "userId": "user-123",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T12:30:00Z",
  "templateConfig": {
    "sm": [
      {
        "w": 3,
        "h": 2,
        "x": 0,
        "y": 0,
        "i": "widget1",
        "static": false,
        "maxH": 4,
        "minH": 1
      }
    ],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "custom-dashboard-template",
    "displayName": "My Custom Dashboard"
  },
  "default": false
}
```

**Error Responses:**
- `400` - Bad request (invalid template data)
- `403` - Unauthorized access (template belongs to different user)
- `404` - Dashboard template not found
- `500` - Internal server error

#### DELETE `/{dashboardTemplateId}`
Delete a specific dashboard template.

**Request:**
```bash
curl -X DELETE \
  'http://localhost:8080/api/widget-layout/v1/1' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (204 No Content):**
```
(Empty response body)
```

**Error Responses:**
- `403` - Unauthorized access (template belongs to different user)
- `404` - Dashboard template not found
- `500` - Internal server error

#### POST `/{dashboardTemplateId}/copy`
Create a copy of a specific dashboard template.

**Request:**
```bash
curl -X POST \
  'http://localhost:8080/api/widget-layout/v1/1/copy' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "id": 2,
  "userId": "user-123",
  "createdAt": "2024-01-01T13:00:00Z",
  "updatedAt": "2024-01-01T13:00:00Z",
  "templateConfig": {
    "sm": [...],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "custom-dashboard-template-copy",
    "displayName": "Copy of My Custom Dashboard"
  },
  "default": false
}
```

**Error Responses:**
- `404` - Dashboard template not found
- `500` - Internal server error

#### POST `/{dashboardTemplateId}/default`
Set a specific dashboard template as the default.

**Request:**
```bash
curl -X POST \
  'http://localhost:8080/api/widget-layout/v1/1/default' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "id": 1,
  "userId": "user-123",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T13:30:00Z",
  "templateConfig": {
    "sm": [...],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "custom-dashboard-template",
    "displayName": "My Custom Dashboard"
  },
  "default": true
}
```

**Error Responses:**
- `403` - Unauthorized access (template belongs to different user)
- `404` - Dashboard template not found
- `500` - Internal server error

#### POST `/{dashboardTemplateId}/reset`
Reset a specific dashboard template to its default state.

**Request:**
```bash
curl -X POST \
  'http://localhost:8080/api/widget-layout/v1/1/reset' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "id": 1,
  "userId": "user-123",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T14:00:00Z",
  "templateConfig": {
    "sm": [...],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "custom-dashboard-template",
    "displayName": "My Custom Dashboard"
  },
  "default": false
}
```

**Error Responses:**
- `403` - Unauthorized access (template belongs to different user)
- `404` - Dashboard template not found
- `500` - Internal server error

### Base Templates

Base templates are predefined widget layouts that serve as starting points for creating custom dashboard templates.

#### GET `/base-templates`
Get all available base widget dashboard templates.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/base-templates' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
[
  {
    "name": "default-dashboard",
    "displayName": "Default Dashboard",
    "templateConfig": {
      "sm": [
        {
          "w": 2,
          "h": 2,
          "x": 0,
          "y": 0,
          "i": "insights-dashboard-widget",
          "static": false,
          "maxH": 4,
          "minH": 1
        }
      ],
      "md": [...],
      "lg": [...],
      "xl": [...]
    }
  }
]
```

**Error Responses:**
- `500` - Internal server error

#### GET `/base-templates/{baseTemplateName}`
Get a specific base widget dashboard template by name.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/base-templates/default-dashboard' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "name": "default-dashboard",
  "displayName": "Default Dashboard",
  "templateConfig": {
    "sm": [
      {
        "w": 2,
        "h": 2,
        "x": 0,
        "y": 0,
        "i": "insights-dashboard-widget",
        "static": false,
        "maxH": 4,
        "minH": 1
      }
    ],
    "md": [...],
    "lg": [...],
    "xl": [...]
  }
}
```

**Error Responses:**
- `404` - Base template not found
- `500` - Internal server error

#### GET `/base-templates/{baseTemplateName}/fork`
Create a user-specific copy of a base template.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/base-templates/default-dashboard/fork' \
  -H 'x-rh-identity: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...'
```

**Response (200 OK):**
```json
{
  "id": 3,
  "userId": "user-123",
  "createdAt": "2024-01-01T15:00:00Z",
  "updatedAt": "2024-01-01T15:00:00Z",
  "templateConfig": {
    "sm": [...],
    "md": [...],
    "lg": [...],
    "xl": [...]
  },
  "templateBase": {
    "name": "default-dashboard",
    "displayName": "Default Dashboard"
  },
  "default": false
}
```

**Error Responses:**
- `404` - Base template not found
- `500` - Internal server error

### Widget Mapping

Widget mapping provides metadata about available widgets including their configurations, dimensions, and module federation information.

#### GET `/widget-mapping`
Get the mapping of all available widgets.

**Request:**
```bash
curl -X GET \
  'http://localhost:8080/api/widget-layout/v1/widget-mapping'
```

**Response (200 OK):**
```json
{
  "insights@dashboard-widget": {
    "scope": "insights",
    "module": "dashboard-widget",
    "importName": "DashboardWidget",
    "featureFlag": "",
    "config": {
      "title": "Dashboard Overview",
      "icon": "dashboard-icon",
      "headerLink": {
        "name": "View Dashboard",
        "href": "/dashboard"
      },
      "permissions": ["dashboard:read"]
    },
    "defaults": {
      "w": 2,
      "h": 3,
      "maxH": 6,
      "minH": 1
    }
  },
  "monitoring@alerts-widget": {
    "scope": "monitoring",
    "module": "alerts-widget",
    "importName": "AlertsWidget",
    "featureFlag": "alerts.enabled",
    "config": {
      "title": "Alert Status",
      "icon": "alert-icon",
      "headerLink": {
        "name": "View Alerts",
        "href": "/alerts"
      },
      "permissions": ["alerts:read"]
    },
    "defaults": {
      "w": 1,
      "h": 2,
      "maxH": 4,
      "minH": 1
    }
  }
}
```

**Error Responses:**
- `500` - Internal server error

> **📋 Configuration Details**: For information about how widget mappings and base templates are configured, see **[docs/CONFIGURATION.md](docs/CONFIGURATION.md)**. This includes JSON structure examples, environment variable setup, and the important cx/cy coordinate system details.

---

## Data Schemas

### WidgetItem
Represents a single widget in a dashboard layout.

```json
{
  "w": 2,              // Width (1-4)
  "h": 2,              // Height (minimum 1)
  "x": 0,              // X position (0-3)
  "y": 0,              // Y position (minimum 0)
  "i": "widget-id",    // Widget type identifier
  "static": false,     // Whether widget is locked
  "maxH": 4,           // Maximum height
  "minH": 1            // Minimum height
}
```

### DashboardTemplateConfig
Defines widget layouts for different screen sizes.

```json
{
  "sm": [/* WidgetItem array for small screens */],
  "md": [/* WidgetItem array for medium screens */],
  "lg": [/* WidgetItem array for large screens */],
  "xl": [/* WidgetItem array for extra large screens */]
}
```

### DashboardTemplate
Complete dashboard template with metadata and configuration.

```json
{
  "id": 1,
  "userId": "user-123",
  "createdAt": "2024-01-01T12:00:00Z",
  "updatedAt": "2024-01-01T12:00:00Z",
  "deletedAt": null,
  "templateConfig": {/* DashboardTemplateConfig */},
  "templateBase": {
    "name": "dashboard-template-v1",
    "displayName": "Template Display Name"
  },
  "default": false
}
```

### BaseWidgetDashboardTemplate
Base template definition without user-specific metadata.

```json
{
  "name": "insights-dashboard-template",
  "displayName": "Template Display Name",
  "templateConfig": {/* DashboardTemplateConfig */}
}
```

### WidgetModuleFederationMetadata
Metadata for widget module federation and configuration.

```json
{
  "scope": "insights",
  "module": "dashboard-widget",
  "importName": "DashboardWidget",
  "featureFlag": "feature.enabled",
  "config": {
    "title": "Widget Title",
    "icon": "widget-icon",
    "headerLink": {
      "name": "Link Name",
      "href": "/link-url"
    },
    "permissions": ["permission:read"]
  },
  "defaults": {
    "w": 2,
    "h": 3,
    "maxH": 6,
    "minH": 1
  }
}
```

## Error Handling

All errors follow a consistent format:

```json
{
  "errors": [
    {
      "code": 400,
      "message": "Detailed error message"
    }
  ]
}
```

### Common Error Codes

- `400` - Bad Request (invalid data)
- `403` - Forbidden (unauthorized access)
- `404` - Not Found (resource doesn't exist)
- `500` - Internal Server Error

## Authorization

The API implements user-based authorization where:

1. **User Identity**: Extracted from `x-rh-identity` header
2. **Template Ownership**: Users can only access their own dashboard templates
3. **Base Templates**: Available to all authenticated users
4. **Widget Mapping**: Available without authentication

## Testing the API

For testing the API locally:

1. **Start the server**: `make dev`
2. **Generate identity header**: `make generate-identity`
3. **Use the provided curl examples** with your generated identity header

For comprehensive testing patterns and examples, see [docs/TESTING.md](docs/TESTING.md).

## OpenAPI Specification

The complete OpenAPI specification is available at:
- YAML format: `/api/widget-layout/v1/openapi.yaml`
- JSON format: `/api/widget-layout/v1/openapi.json`

## Related Documentation

- [docs/TESTING.md](docs/TESTING.md) - Testing patterns and examples
- [docs/CONFIGURATION.md](docs/CONFIGURATION.md) - Widget mapping and base template configuration
- [docs/DEVELOPMENT_IDENTITY_HEADER.md](docs/DEVELOPMENT_IDENTITY_HEADER.md) - Identity header for local development
- [docs/AI_AGENT_CONTEXT.md](docs/AI_AGENT_CONTEXT.md) - AI-assisted development guidelines
