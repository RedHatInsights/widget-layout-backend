import { z } from 'zod';

const LOG_LEVELS = ['trace', 'debug', 'info', 'warn', 'error', 'fatal'] as const;
export type LogLevel = (typeof LOG_LEVELS)[number];

const isLogLevel = (value: string): value is LogLevel =>
  (LOG_LEVELS as readonly string[]).includes(value);

// logrus-style values like "Debug" are used on the Go saas targets.
// Accept those case-insensitively; anything else falls back to info so
// the process does not throw at import time and CrashLoop.
export function resolveLogLevel(value: unknown): LogLevel {
  if (typeof value !== 'string' || value.trim() === '') {
    return 'info';
  }
  const normalized = value.trim().toLowerCase();
  return isLogLevel(normalized) ? normalized : 'info';
}

const configSchema = z.object({
  port: z.coerce.number().default(8001),
  metricsPort: z.coerce.number().default(9000),
  widgetLayoutApiUrl: z.string().default('http://localhost:8000'),
  logLevel: z.enum(LOG_LEVELS).default('info'),
  nodeEnv: z.enum(['development', 'production', 'test']).default('development'),
});

export type Config = z.infer<typeof configSchema>;

export const config = configSchema.parse({
  port: process.env.PORT,
  metricsPort: process.env.METRICS_PORT,
  widgetLayoutApiUrl: process.env.WIDGET_LAYOUT_API_URL,
  logLevel: resolveLogLevel(process.env.LOG_LEVEL),
  nodeEnv: process.env.NODE_ENV,
});
