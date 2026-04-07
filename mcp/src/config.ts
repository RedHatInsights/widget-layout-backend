import { z } from 'zod';

const configSchema = z.object({
  port: z.coerce.number().default(8001),
  widgetLayoutApiUrl: z.string().default('http://localhost:8000'),
  logLevel: z.enum(['trace', 'debug', 'info', 'warn', 'error', 'fatal']).default('info'),
  nodeEnv: z.enum(['development', 'production', 'test']).default('development'),
});

export type Config = z.infer<typeof configSchema>;

export function loadConfig(): Config {
  return configSchema.parse({
    port: process.env.PORT,
    widgetLayoutApiUrl: process.env.WIDGET_LAYOUT_API_URL,
    logLevel: process.env.LOG_LEVEL,
    nodeEnv: process.env.NODE_ENV,
  });
}
