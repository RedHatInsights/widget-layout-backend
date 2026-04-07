# Widget Layout MCP Sidecar

TypeScript-based Model Context Protocol (MCP) sidecar container that enables AI agents to interact with widget dashboard data via a standardized JSON-RPC 2.0 protocol.

## Overview

This sidecar container runs alongside the main `widget-layout-backend` Go application and provides a standardized MCP interface for AI agents. It acts as a bridge, translating MCP tool calls into HTTP requests to the main application's REST API.

## Architecture

```text
┌─────────────────────────────────────────────────┐
│  Pod: widget-layout-backend                     │
│                                                  │
│  ┌──────────────────┐    ┌──────────────────┐  │
│  │  Main Container  │    │  MCP Sidecar     │  │
│  │  (Go/chi)        │◄───┤  (TypeScript)    │  │
│  │  Port: 8000      │    │  Port: 8001      │  │
│  └──────────────────┘    └──────────────────┘  │
│         │                          ▲            │
│         │ Internal                 │            │
│         │ localhost                │            │
│         │ HTTP calls               │            │
│         │                          │            │
└─────────┼──────────────────────────┼────────────┘
          │                          │
          │                          │
    Public API                 MCP Endpoint
   /api/widget-layout/v1/    /_private/mcp/
```

## Available Tools

All tools are **read-only** to minimize risk and complexity.

### 1. `hello` (No auth required)
Health check and smoke test for the MCP endpoint.

**Parameters:** None

**Example:**
```json
{
  "name": "hello",
  "arguments": {}
}
```

### 2. `get_widget_layouts` (Auth required)
List all widget dashboard templates for the authenticated organization.

**Parameters:**
- `dashboardType` (optional): Filter by dashboard type

**Example:**
```json
{
  "name": "get_widget_layouts",
  "arguments": {
    "dashboardType": "analytics"
  }
}
```

### 3. `get_widget_layout_by_id` (Auth required)
Get a specific dashboard template by ID.

**Parameters:**
- `dashboard_template_id` (required): The unique identifier of the dashboard template

**Example:**
```json
{
  "name": "get_widget_layout_by_id",
  "arguments": {
    "dashboard_template_id": 123
  }
}
```

### 4. `get_base_templates` (No auth required)
List all available base dashboard templates.

**Parameters:** None

**Example:**
```json
{
  "name": "get_base_templates",
  "arguments": {}
}
```

### 5. `get_widget_mapping` (No auth required)
Get the widget registry/catalog with module federation metadata.

**Parameters:** None

**Example:**
```json
{
  "name": "get_widget_mapping",
  "arguments": {}
}
```

### 6. `export_widget_layout` (Auth required)
Export a dashboard template as a shareable configuration.

**Parameters:**
- `dashboard_template_id` (required): The unique identifier of the dashboard template to export

**Example:**
```json
{
  "name": "export_widget_layout",
  "arguments": {
    "dashboard_template_id": 123
  }
}
```

## Development

### Prerequisites

- Node.js 20.x or higher
- npm

### Setup

```bash
cd mcp
npm install
```

### Development Mode

```bash
npm run dev
```

Or from the root directory:
```bash
make dev-mcp
```

### Environment Variables

- `PORT` - Server port (default: 8001)
- `WIDGET_LAYOUT_API_URL` - URL of the main widget-layout API (default: http://localhost:8000)
- `LOG_LEVEL` - Logging level: trace, debug, info, warn, error, fatal (default: info)
- `NODE_ENV` - Node environment: development, production, test (default: development)

### Running Tests

```bash
npm test
```

Or from the root directory:
```bash
make test-mcp
```

For coverage:
```bash
npm run test:coverage
```

### Linting

```bash
npm run lint
npm run lint:fix
```

Or from the root directory:
```bash
make lint-mcp
```

## Building

### Docker Image

```bash
docker build -t widget-layout-mcp-sidecar:latest .
```

Or from the root directory:
```bash
make build-mcp
```

### Local Build

```bash
npm run build
```

The compiled JavaScript will be in the `dist/` directory.

## Authentication

Tools that require authentication expect an `x-rh-identity` header containing a base64-encoded JSON object:

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

The sidecar validates the identity and forwards the header to the main application.

## API Endpoints

### MCP Endpoint

**POST** `/_private/mcp`

Accepts JSON-RPC 2.0 requests following the MCP specification.

**Example Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "get_widget_layouts",
    "arguments": {}
  }
}
```

**Example Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"data\": [...], \"meta\": {\"count\": 5}}"
      }
    ]
  }
}
```

### Health Check

**GET** `/healthz`

Returns server health status.

### Readiness Check

**GET** `/ready`

Returns readiness status (useful for Kubernetes readiness probes).

### Metrics

**GET** `/metrics`

Returns Prometheus metrics in text format.

## Metrics

The sidecar exposes the following Prometheus metrics:

- `mcp_tool_call_total{tool, status}` - Counter for total tool calls
- `mcp_tool_call_duration_seconds{tool}` - Histogram for tool call duration
- `mcp_api_call_duration_seconds{endpoint, method, status}` - Histogram for API call duration
- `mcp_auth_failure_total{reason}` - Counter for authentication failures

## Logging

All logs are structured JSON (in production) or pretty-printed (in development) using Pino.

Logs use the `mcp:` prefix for easy filtering:

```
{"level":"info","time":1234567890,"service":"mcp-sidecar","msg":"mcp: tool call completed","tool":"get_widget_layouts","duration":0.234,"org_id":"12345"}
```

## Project Structure

```
mcp/
├── src/
│   ├── types/           # TypeScript type definitions
│   │   ├── mcp.ts       # MCP protocol types
│   │   ├── identity.ts  # x-rh-identity types
│   │   └── widget-api.ts # Widget API response types
│   ├── tools/           # MCP tool implementations
│   │   ├── index.ts     # Tool registry
│   │   ├── hello.ts
│   │   ├── get-layouts.ts
│   │   ├── get-layout-by-id.ts
│   │   ├── get-base-templates.ts
│   │   ├── get-widget-mapping.ts
│   │   └── export-layout.ts
│   ├── utils/           # Utility modules
│   │   ├── api-client.ts # HTTP client
│   │   ├── identity.ts   # Identity parsing
│   │   ├── logger.ts     # Structured logging
│   │   └── metrics.ts    # Prometheus metrics
│   ├── config.ts        # Configuration
│   ├── server.ts        # MCP server implementation
│   └── index.ts         # Express server entry point
├── tests/               # Unit and integration tests
├── Dockerfile           # Container build
└── package.json         # Dependencies
```

## References

- [MCP Specification](https://modelcontextprotocol.io/specification/2025-03-26/)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
- [insights-rbac MCP PR #2537](https://github.com/RedHatInsights/insights-rbac/pull/2537)
