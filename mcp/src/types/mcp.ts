// MCP Protocol Types (JSON-RPC 2.0)

export interface JsonRpcRequest {
  jsonrpc: '2.0';
  id?: string | number;
  method: string;
  params?: Record<string, unknown>;
}

export interface JsonRpcResponse {
  jsonrpc: '2.0';
  id: string | number | null;
  result?: unknown;
  error?: JsonRpcError;
}

export interface JsonRpcError {
  code: number;
  message: string;
  data?: unknown;
}

// MCP Protocol Methods
export type McpMethod = 'initialize' | 'tools/list' | 'tools/call' | 'ping';

// Initialize Request/Response
export interface InitializeParams {
  protocolVersion: string;
  capabilities: {
    tools?: Record<string, unknown>;
  };
  clientInfo: {
    name: string;
    version: string;
  };
}

export interface InitializeResult {
  protocolVersion: string;
  capabilities: {
    tools?: Record<string, unknown>;
  };
  serverInfo: {
    name: string;
    version: string;
  };
}

// Tools List Request/Response
export interface ToolsListResult {
  tools: ToolDefinition[];
}

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: {
    type: 'object';
    properties: Record<string, PropertySchema>;
    required?: string[];
  };
}

export interface PropertySchema {
  type: string;
  description?: string;
  enum?: string[];
  items?: PropertySchema;
}

// Tools Call Request/Response
export interface ToolsCallParams {
  name: string;
  arguments?: Record<string, unknown>;
}

export interface ToolsCallResult {
  content: ToolContent[];
  isError?: boolean;
}

export interface ToolContent {
  type: 'text';
  text: string;
}

// Standard JSON-RPC Error Codes
export enum JsonRpcErrorCode {
  ParseError = -32700,
  InvalidRequest = -32600,
  MethodNotFound = -32601,
  InvalidParams = -32602,
  InternalError = -32603,
}
