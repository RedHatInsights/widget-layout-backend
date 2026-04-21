import {
  JsonRpcRequest,
  JsonRpcResponse,
  JsonRpcErrorCode,
  InitializeParams,
  InitializeResult,
  ToolsListResult,
  ToolsCallParams,
  ToolsCallResult,
} from './types/mcp';
import { Identity } from './types/identity';
import { parseIdentity } from './utils/identity';
import { toolRegistry } from './tools/index';
import { logger } from './utils/logger';

// Type guards for request params validation
function isInitializeParams(params: unknown): params is InitializeParams {
  if (!params || typeof params !== 'object') return false;
  const p = params as Record<string, unknown>;
  return typeof p.protocolVersion === 'string';
}

function isToolsCallParams(params: unknown): params is ToolsCallParams {
  if (!params || typeof params !== 'object') return false;
  const p = params as Record<string, unknown>;
  return typeof p.name === 'string';
}

// Import all tools to register them
import './tools/hello';
import './tools/get-layouts';
import './tools/get-layout-by-id';
import './tools/get-base-templates';
import './tools/get-widget-mapping';
import './tools/export-layout';

export class McpServer {
  private initialized = false;
  private readonly protocolVersion = '2024-11-05';
  private readonly serverInfo = {
    name: 'widget-layout-mcp-sidecar',
    version: '1.0.0',
  };

  async handleRequest(
    request: JsonRpcRequest,
    identityHeader?: string
  ): Promise<JsonRpcResponse> {
    logger.debug(
      {
        method: request.method,
        id: request.id,
      },
      'mcp: Handling request'
    );

    try {
      let result: unknown;

      switch (request.method) {
        case 'initialize':
          if (!isInitializeParams(request.params)) {
            return this.createErrorResponse(
              request.id ?? null,
              JsonRpcErrorCode.InvalidParams,
              'Invalid initialize params: protocolVersion is required'
            );
          }
          result = await this.handleInitialize(request.params);
          break;

        case 'tools/list':
          result = this.handleToolsList();
          break;

        case 'tools/call':
          if (!isToolsCallParams(request.params)) {
            return this.createErrorResponse(
              request.id ?? null,
              JsonRpcErrorCode.InvalidParams,
              'Invalid tools/call params: name is required'
            );
          }
          result = await this.handleToolsCall(
            request.params,
            identityHeader
          );
          break;

        case 'ping':
          result = { status: 'ok' };
          break;

        default:
          return this.createErrorResponse(
            request.id ?? null,
            JsonRpcErrorCode.MethodNotFound,
            `Method '${request.method}' not found`
          );
      }

      return {
        jsonrpc: '2.0',
        id: request.id ?? null,
        result,
      };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error(
        {
          method: request.method,
          error: errorMessage,
        },
        'mcp: Request handling failed'
      );

      return this.createErrorResponse(
        request.id ?? null,
        JsonRpcErrorCode.InternalError,
        errorMessage
      );
    }
  }

  private async handleInitialize(params: InitializeParams): Promise<InitializeResult> {
    logger.info(
      {
        clientInfo: params.clientInfo,
        protocolVersion: params.protocolVersion,
      },
      'mcp: Initialize request'
    );

    this.initialized = true;

    return {
      protocolVersion: this.protocolVersion,
      capabilities: {
        tools: {},
      },
      serverInfo: this.serverInfo,
    };
  }

  private handleToolsList(): ToolsListResult {
    if (!this.initialized) {
      logger.warn('mcp: tools/list called before initialization');
    }

    const tools = toolRegistry.getAll();

    logger.info(
      {
        toolCount: tools.length,
        tools: tools.map((t) => t.name),
      },
      'mcp: Listing tools'
    );

    return { tools };
  }

  private async handleToolsCall(
    params: ToolsCallParams,
    identityHeader?: string
  ): Promise<ToolsCallResult> {
    if (!this.initialized) {
      logger.warn('mcp: tools/call called before initialization');
    }

    // Parse identity if provided
    let identity: Identity | null = null;
    if (identityHeader) {
      try {
        identity = parseIdentity(identityHeader);
      } catch (error) {
        // If identity parsing fails, we'll pass null and let the tool handle it
        logger.warn(
          { error: error instanceof Error ? error.message : 'Unknown error' },
          'mcp: Failed to parse identity, proceeding with null identity'
        );
      }
    }

    const args = params.arguments && typeof params.arguments === 'object' ? params.arguments : {};
    return await toolRegistry.execute(params.name, args, identity);
  }

  private createErrorResponse(
    id: string | number | null,
    code: JsonRpcErrorCode,
    message: string
  ): JsonRpcResponse {
    return {
      jsonrpc: '2.0',
      id,
      error: {
        code,
        message,
      },
    };
  }

  isInitialized(): boolean {
    return this.initialized;
  }
}
