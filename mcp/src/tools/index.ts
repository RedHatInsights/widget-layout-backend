import { ToolDefinition, ToolsCallResult } from '../types/mcp';
import { Identity } from '../types/identity';
import { validateIdentity } from '../utils/identity';
import { logger } from '../utils/logger';
import { mcpToolCallTotal, mcpToolCallDuration } from '../utils/metrics';

export interface ToolExecutor {
  name: string;
  description: string;
  inputSchema: ToolDefinition['inputSchema'];
  requiresAuth: boolean;
  execute: (args: Record<string, unknown>, identity: Identity | null) => Promise<unknown>;
}

class ToolRegistry {
  private tools: Map<string, ToolExecutor> = new Map();

  register(tool: ToolExecutor): void {
    this.tools.set(tool.name, tool);
    logger.info({ tool: tool.name }, 'mcp: Registered tool');
  }

  getAll(): ToolDefinition[] {
    return Array.from(this.tools.values()).map((tool) => ({
      name: tool.name,
      description: tool.description,
      inputSchema: tool.inputSchema,
    }));
  }

  async execute(
    name: string,
    args: Record<string, unknown> = {},
    identity: Identity | null
  ): Promise<ToolsCallResult> {
    const tool = this.tools.get(name);

    if (!tool) {
      const error = `Tool '${name}' not found`;
      logger.warn({ tool: name }, 'mcp: Tool not found');
      mcpToolCallTotal.inc({ tool: name, status: 'not_found' });
      return {
        content: [{ type: 'text', text: error }],
        isError: true,
      };
    }

    const startTime = Date.now();

    try {
      // Validate authentication
      validateIdentity(identity, tool.requiresAuth);

      logger.info(
        {
          tool: name,
          args,
          org_id: identity?.org_id,
        },
        'mcp: Executing tool'
      );

      const result = await tool.execute(args, identity);
      const duration = (Date.now() - startTime) / 1000;

      mcpToolCallDuration.observe({ tool: name }, duration);
      mcpToolCallTotal.inc({ tool: name, status: 'success' });

      logger.info(
        {
          tool: name,
          duration,
          org_id: identity?.org_id,
        },
        'mcp: Tool call completed'
      );

      return {
        content: [
          {
            type: 'text',
            text: JSON.stringify(result, null, 2),
          },
        ],
      };
    } catch (error) {
      const duration = (Date.now() - startTime) / 1000;
      mcpToolCallDuration.observe({ tool: name }, duration);
      mcpToolCallTotal.inc({ tool: name, status: 'error' });

      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      logger.error(
        {
          tool: name,
          error: errorMessage,
          duration,
          org_id: identity?.org_id,
        },
        'mcp: Tool call failed'
      );

      return {
        content: [
          {
            type: 'text',
            text: `Error: ${errorMessage}`,
          },
        ],
        isError: true,
      };
    }
  }
}

export const toolRegistry = new ToolRegistry();
