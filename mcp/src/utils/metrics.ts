import { register, Counter, Histogram } from 'prom-client';

// Counter for total tool calls
export const mcpToolCallTotal = new Counter({
  name: 'mcp_tool_call_total',
  help: 'Total number of MCP tool calls',
  labelNames: ['tool', 'status'] as const,
});

// Histogram for tool call duration
export const mcpToolCallDuration = new Histogram({
  name: 'mcp_tool_call_duration_seconds',
  help: 'Duration of MCP tool calls in seconds',
  labelNames: ['tool'] as const,
  buckets: [0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10],
});

// Histogram for API call duration
export const mcpApiCallDuration = new Histogram({
  name: 'mcp_api_call_duration_seconds',
  help: 'Duration of API calls to main widget-layout service in seconds',
  labelNames: ['endpoint', 'method', 'status'] as const,
  buckets: [0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10],
});

// Counter for authentication failures
export const mcpAuthFailureTotal = new Counter({
  name: 'mcp_auth_failure_total',
  help: 'Total number of authentication failures',
  labelNames: ['reason'] as const,
});

export function getMetrics(): Promise<string> {
  return register.metrics();
}

export function clearMetrics(): void {
  register.clear();
}
