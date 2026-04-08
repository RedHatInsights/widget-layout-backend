import { toolRegistry } from './index';
// eslint-disable-next-line @typescript-eslint/no-var-requires
const { version } = require('../../package.json');

toolRegistry.register({
  name: 'hello',
  description: 'Health check and smoke test for the MCP endpoint. Returns a welcome message and server status.',
  inputSchema: {
    type: 'object',
    properties: {},
  },
  requiresAuth: false,
  execute: async () => {
    return {
      message: 'Hello from Widget Layout MCP Sidecar!',
      status: 'healthy',
      timestamp: new Date().toISOString(),
      version: process.env.MCP_VERSION || version || 'unknown',
    };
  },
});
