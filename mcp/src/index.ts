import express, { Request, Response, NextFunction } from 'express';
import { loadConfig } from './config';
import { logger } from './utils/logger';
import { getMetrics } from './utils/metrics';
import { McpServer } from './server';
import { JsonRpcRequest } from './types/mcp';

const config = loadConfig();
const app = express();
const mcpServer = new McpServer();

// Boot readiness flag - set to true after server starts listening
let serverBootReady = false;

// Middleware
app.use(express.json());

// JSON parse error handler - must come right after express.json()
app.use((err: Error, _req: Request, res: Response, next: NextFunction): void => {
  if (err instanceof SyntaxError && 'body' in err) {
    res.status(400).json({
      jsonrpc: '2.0',
      id: null,
      error: {
        code: -32700,
        message: `Parse error: ${err.message}`,
      },
    });
    return;
  }
  next(err);
});

// Request ID middleware
app.use((req: Request, res: Response, next: NextFunction) => {
  const reqId = req.headers['x-request-id'] || `req-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
  res.locals.reqId = reqId;
  next();
});

// Request logging middleware
app.use((req: Request, res: Response, next: NextFunction) => {
  const start = Date.now();

  res.on('finish', () => {
    const duration = Date.now() - start;
    logger.info(
      {
        method: req.method,
        path: req.path,
        status: res.statusCode,
        duration,
        req_id: res.locals.reqId,
      },
      'HTTP request'
    );
  });

  next();
});

// Health check endpoint
app.get('/healthz', (_req: Request, res: Response) => {
  res.status(200).json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  });
});

// Readiness check endpoint
app.get('/ready', (_req: Request, res: Response) => {
  res.status(serverBootReady ? 200 : 503).json({
    status: serverBootReady ? 'ready' : 'not ready',
    timestamp: new Date().toISOString(),
  });
});

// Prometheus metrics endpoint
app.get('/metrics', async (_req: Request, res: Response) => {
  try {
    const metrics = await getMetrics();
    res.set('Content-Type', 'text/plain');
    res.send(metrics);
  } catch (error) {
    logger.error({ error }, 'Failed to get metrics');
    res.status(500).json({ error: 'Failed to get metrics' });
  }
});

// MCP endpoint
app.post('/_private/mcp', async (req: Request, res: Response): Promise<void> => {
  try {
    // Validate that body is an object
    if (!req.body || typeof req.body !== 'object' || Array.isArray(req.body)) {
      res.status(400).json({
        jsonrpc: '2.0',
        id: null,
        error: {
          code: -32600,
          message: 'Invalid Request: body must be a JSON-RPC object',
        },
      });
      return;
    }

    const request = req.body as JsonRpcRequest;
    const identityHeader = req.headers['x-rh-identity'] as string | undefined;

    logger.debug(
      {
        method: request.method,
        id: request.id,
        hasIdentity: !!identityHeader,
        req_id: res.locals.reqId,
      },
      'mcp: Received MCP request'
    );

    // Validate JSON-RPC request
    if (!request.jsonrpc || request.jsonrpc !== '2.0') {
      res.status(400).json({
        jsonrpc: '2.0',
        id: request.id ?? null,
        error: {
          code: -32600,
          message: 'Invalid Request: jsonrpc must be "2.0"',
        },
      });
      return;
    }

    // Validate method is a non-empty string
    if (!request.method || typeof request.method !== 'string' || request.method.trim() === '') {
      res.status(400).json({
        jsonrpc: '2.0',
        id: request.id ?? null,
        error: {
          code: -32600,
          message: 'Invalid Request: method must be a non-empty string',
        },
      });
      return;
    }

    // Validate id is number, string, or null (not objects/arrays)
    if (request.id !== undefined && request.id !== null &&
        typeof request.id !== 'string' && typeof request.id !== 'number') {
      res.status(400).json({
        jsonrpc: '2.0',
        id: null,
        error: {
          code: -32600,
          message: 'Invalid Request: id must be a string, number, or null',
        },
      });
      return;
    }

    // Validate params is object or array (not primitives or null)
    if (request.params !== undefined &&
        (request.params === null || typeof request.params !== 'object')) {
      res.status(400).json({
        jsonrpc: '2.0',
        id: request.id ?? null,
        error: {
          code: -32600,
          message: 'Invalid Request: params must be an object or array',
        },
      });
      return;
    }

    const response = await mcpServer.handleRequest(request, identityHeader);
    res.json(response);
  } catch (error) {
    logger.error(
      {
        error,
        req_id: res.locals.reqId,
      },
      'mcp: Unhandled error in MCP endpoint'
    );

    res.status(500).json({
      jsonrpc: '2.0',
      id: null,
      error: {
        code: -32603,
        message: 'Internal server error',
      },
    });
  }
});

// 404 handler
app.use((_req: Request, res: Response) => {
  res.status(404).json({
    error: 'Not Found',
    message: 'The requested endpoint does not exist',
  });
});

// Error handler
app.use((err: Error, _req: Request, res: Response, _next: NextFunction): void => {
  logger.error({ error: err }, 'Unhandled error');
  res.status(500).json({
    error: 'Internal Server Error',
    message: config.nodeEnv === 'development' ? err.message : 'An unexpected error occurred',
  });
});

// Start server
const server = app.listen(config.port, () => {
  serverBootReady = true;
  logger.info(
    {
      port: config.port,
      nodeEnv: config.nodeEnv,
      widgetLayoutApiUrl: config.widgetLayoutApiUrl,
    },
    'MCP Sidecar server started'
  );
});

// Handle server listen errors
server.on('error', (err: Error) => {
  logger.error(
    {
      error: err,
      port: config.port,
      nodeEnv: config.nodeEnv,
    },
    'Failed to start MCP Sidecar server'
  );
  serverBootReady = false;
  process.exit(1);
});

// Graceful shutdown
const gracefulShutdown = (signal: string) => {
  logger.info({ signal }, 'Received shutdown signal');

  // Force shutdown after 10 seconds
  const forceTimeout = setTimeout(() => {
    logger.error('Forced shutdown after timeout');
    process.exit(1);
  }, 10000);

  server.close(() => {
    clearTimeout(forceTimeout);
    logger.info('Server closed');
    process.exit(0);
  });
};

process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
process.on('SIGINT', () => gracefulShutdown('SIGINT'));

export default app;
