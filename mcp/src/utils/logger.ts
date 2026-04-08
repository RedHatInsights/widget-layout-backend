import pino from 'pino';
import { loadConfig } from '../config';

const config = loadConfig();

export const logger = pino({
  level: config.logLevel,
  transport: config.nodeEnv === 'development' ? {
    target: 'pino-pretty',
    options: {
      colorize: true,
      translateTime: 'HH:MM:ss Z',
      ignore: 'pid,hostname',
    },
  } : undefined,
  base: {
    service: 'mcp-sidecar',
  },
  formatters: {
    level: (label) => {
      return { level: label };
    },
  },
});

export function createRequestLogger(reqId: string) {
  return logger.child({ req_id: reqId });
}
