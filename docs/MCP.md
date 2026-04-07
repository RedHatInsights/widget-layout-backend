# Model Context Protocol (MCP) Integration

This document describes the MCP integration for the widget-layout-backend service, which enables AI agents to interact with widget dashboard data via a standardized protocol.

## What is MCP?

The Model Context Protocol (MCP) is an open protocol that standardizes how AI applications provide context to Large Language Models (LLMs). It enables AI agents to:

- Discover available tools and capabilities
- Execute operations with proper authentication
- Receive structured responses in a standard format

## Architecture

The widget-layout-backend uses a **sidecar container pattern** for MCP integration:

```
┌─────────────────────────────────────────────────┐
│  Pod: widget-layout-backend                     │
│                                                  │
│  ┌──────────────────┐    ┌──────────────────┐  │
│  │  Main Container  │    │  MCP Sidecar     │  │
│  │  (Go/chi)        │◄───┤  (TypeScript)    │  │
│  │  Port: 8000      │    │  Port: 8001      │  │
│  └──────────────────┘    └──────────────────┘  │
└─────────────────────────────────────────────────┘
```

**Why a sidecar?**

- The main application is written in Go, while MCP SDK is available in TypeScript/Python
- Separates concerns: business logic in Go, MCP protocol handling in TypeScript
- Allows independent updates and separation of concerns (note: sidecar and main container share pod lifecycle)
- TypeScript provides excellent type safety and smaller container images than Python

## Available Tools

The MCP sidecar exposes 6 read-only tools for AI agents:

| Tool Name | Auth Required | Description |
|-----------|--------------|-------------|
| `hello` | No | Health check and smoke test |
| `get_widget_layouts` | Yes | List all dashboard templates for org |
| `get_widget_layout_by_id` | Yes | Get specific dashboard template by ID |
| `get_base_templates` | No | List available base templates |
| `get_widget_mapping` | No | Get widget registry/catalog |
| `export_widget_layout` | Yes | Export dashboard as shareable config |

All tools are **read-only** by design to minimize risk.

## Using the MCP Endpoint

### Endpoint

**POST** `/_private/mcp/`

### Protocol

MCP uses JSON-RPC 2.0 for all requests and responses.

### Authentication

Tools that require authentication expect an `x-rh-identity` header:

```bash
X-RH-Identity: <base64-encoded-json>
```

Example decoded identity:
```json
{
  "identity": {
    "org_id": "12345",
    "type": "User",
    "user": {
      "username": "testuser",
      "email": "test@example.com"
    }
  }
}
```

### Request Flow

1. **Initialize** - Establish protocol version and capabilities
2. **List Tools** - Discover available tools
3. **Call Tool** - Execute a specific tool
4. **Repeat** - Continue calling tools as needed

## Example Usage

### 1. Initialize Session

```bash
curl -X POST http://localhost:8001/_private/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {
        "name": "my-ai-agent",
        "version": "1.0.0"
      }
    }
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "tools": {}
    },
    "serverInfo": {
      "name": "widget-layout-mcp-sidecar",
      "version": "1.0.0"
    }
  }
}
```

### 2. List Available Tools

```bash
curl -X POST http://localhost:8001/_private/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "hello",
        "description": "Health check and smoke test...",
        "inputSchema": {
          "type": "object",
          "properties": {}
        }
      },
      {
        "name": "get_widget_layouts",
        "description": "List all widget dashboard templates...",
        "inputSchema": {
          "type": "object",
          "properties": {
            "dashboardType": {
              "type": "string",
              "description": "Optional filter by dashboard type"
            }
          }
        }
      }
    ]
  }
}
```

### 3. Call a Tool (No Auth)

```bash
curl -X POST http://localhost:8001/_private/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "hello",
      "arguments": {}
    }
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"message\":\"Hello from Widget Layout MCP Sidecar!\",\"status\":\"healthy\",\"timestamp\":\"2024-01-01T00:00:00.000Z\",\"version\":\"1.0.0\"}"
      }
    ]
  }
}
```

### 4. Call a Tool (With Auth)

First, create a base64-encoded identity:

```bash
echo -n '{"identity":{"org_id":"12345","type":"User","user":{"username":"testuser"}}}' | base64
```

Then call the tool:

```bash
curl -X POST http://localhost:8001/_private/mcp \
  -H "Content-Type: application/json" \
  -H "X-RH-Identity: eyJpZGVudGl0eSI6eyJvcmdfaWQiOiIxMjM0NSIsInR5cGUiOiJVc2VyIiwidXNlciI6eyJ1c2VybmFtZSI6InRlc3R1c2VyIn19fQ==" \
  -d '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "get_widget_layouts", "arguments": {}}}'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"data\":[{\"id\":1,\"userId\":\"user1\",\"dashboardName\":\"My Dashboard\",\"createdAt\":\"2024-01-01T00:00:00Z\",\"updatedAt\":\"2024-01-01T00:00:00Z\",\"templateConfig\":{\"sm\":[],\"md\":[],\"lg\":[],\"xl\":[]},\"templateBase\":{\"name\":\"dashboard-1\",\"displayName\":\"Dashboard 1\"},\"default\":true}],\"meta\":{\"count\":1}}"
      }
    ]
  }
}
```

## Tool Details

### hello

**Description:** Health check and smoke test

**Authentication:** Not required

**Parameters:** None

**Use Case:** Verify MCP endpoint is working

**Example:**
```json
{
  "name": "hello",
  "arguments": {}
}
```

---

### get_widget_layouts

**Description:** List all dashboard templates for authenticated organization

**Authentication:** Required

**Parameters:**
- `dashboardType` (optional, string): Filter by dashboard type

**Use Case:** "Show me all my dashboards"

**Example:**
```json
{
  "name": "get_widget_layouts",
  "arguments": {
    "dashboardType": "analytics"
  }
}
```

---

### get_widget_layout_by_id

**Description:** Get specific dashboard template details

**Authentication:** Required

**Parameters:**
- `dashboard_template_id` (required, number): The unique identifier

**Use Case:** "What's in my analytics dashboard?"

**Example:**
```json
{
  "name": "get_widget_layout_by_id",
  "arguments": {
    "dashboard_template_id": 123
  }
}
```

---

### get_base_templates

**Description:** List available base template options

**Authentication:** Not required

**Parameters:** None

**Use Case:** "What dashboard templates are available?"

**Example:**
```json
{
  "name": "get_base_templates",
  "arguments": {}
}
```

---

### get_widget_mapping

**Description:** Get widget registry/catalog

**Authentication:** Not required

**Parameters:** None

**Use Case:** "What widgets are available?"

**Example:**
```json
{
  "name": "get_widget_mapping",
  "arguments": {}
}
```

---

### export_widget_layout

**Description:** Export dashboard as shareable configuration

**Authentication:** Required

**Parameters:**
- `dashboard_template_id` (required, number): The unique identifier

**Use Case:** "Give me a shareable link for this dashboard"

**Example:**
```json
{
  "name": "export_widget_layout",
  "arguments": {
    "dashboard_template_id": 123
  }
}
```

## Error Handling

### Authentication Errors

If a tool requires authentication but no valid identity is provided:

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error: Authentication required for this tool"
      }
    ],
    "isError": true
  }
}
```

### Tool Not Found

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Tool 'invalid_tool' not found"
      }
    ],
    "isError": true
  }
}
```

### API Errors

If the main application returns an error:

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Error: Request failed with status code 404"
      }
    ],
    "isError": true
  }
}
```

## Monitoring

### Metrics

The MCP sidecar exposes Prometheus metrics at `/metrics`:

- `mcp_tool_call_total{tool, status}` - Total tool calls
- `mcp_tool_call_duration_seconds{tool}` - Tool call duration
- `mcp_api_call_duration_seconds{endpoint, method, status}` - API call duration
- `mcp_auth_failure_total{reason}` - Authentication failures

### Logging

All operations are logged with structured JSON:

```json
{
  "level": "info",
  "time": 1234567890,
  "service": "mcp-sidecar",
  "msg": "mcp: tool call completed",
  "tool": "get_widget_layouts",
  "duration": 0.234,
  "org_id": "12345",
  "req_id": "req-abc123"
}
```

## Development

See [mcp/README.md](../mcp/README.md) for development setup and instructions.

## Security Considerations

1. **Read-Only Operations**: All MCP tools are read-only to prevent accidental or malicious modifications
2. **Authentication**: Tools that access user data require valid `x-rh-identity` header
3. **Authorization**: The main application enforces authorization rules (org ownership, permissions)
4. **Input Validation**: All tool parameters are validated using Zod schemas
5. **Rate Limiting**: Consider adding rate limiting for production deployments
6. **Logging**: All tool calls are logged with org_id for audit trails

## Future Enhancements

Potential Phase 2 features:

- **Mutation Tools**: `update_widget_layout`, `copy_widget_layout`, `set_default`
- **Batch Operations**: Process multiple dashboards in one call
- **Webhooks**: Subscribe to dashboard change events
- **Caching**: Cache frequently accessed data to reduce API calls
- **GraphQL Support**: Alternative to REST API calls

## References

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-03-26/)
- [insights-rbac MCP Implementation](https://github.com/RedHatInsights/insights-rbac/pull/2537)
- [Widget Layout OpenAPI Spec](../spec/openapi.yaml)
