import { resolveLogLevel } from '../../src/config';

describe('resolveLogLevel', () => {
  it('defaults to info when unset or empty', () => {
    expect(resolveLogLevel(undefined)).toBe('info');
    expect(resolveLogLevel('')).toBe('info');
    expect(resolveLogLevel('   ')).toBe('info');
  });

  it('accepts allowed levels case-insensitively (logrus Debug on saas files)', () => {
    expect(resolveLogLevel('debug')).toBe('debug');
    expect(resolveLogLevel('Debug')).toBe('debug');
    expect(resolveLogLevel('DEBUG')).toBe('debug');
    expect(resolveLogLevel('WARN')).toBe('warn');
    expect(resolveLogLevel('fatal')).toBe('fatal');
  });

  it('falls back to info for unrecognised values', () => {
    expect(resolveLogLevel('verbose')).toBe('info');
    expect(resolveLogLevel('not-a-level')).toBe('info');
  });
});

describe('config', () => {
  const keys = ['PORT', 'METRICS_PORT', 'WIDGET_LAYOUT_API_URL', 'LOG_LEVEL', 'NODE_ENV'] as const;
  let saved: Record<string, string | undefined>;

  beforeEach(() => {
    saved = {};
    for (const key of keys) {
      saved[key] = process.env[key];
    }
  });

  afterEach(() => {
    for (const key of keys) {
      if (saved[key] === undefined) {
        delete process.env[key];
      } else {
        process.env[key] = saved[key];
      }
    }
    jest.resetModules();
  });

  async function loadConfig() {
    jest.resetModules();
    return import('../../src/config');
  }

  it('defaults metricsPort to 9000', async () => {
    delete process.env.METRICS_PORT;
    const { config } = await loadConfig();
    expect(config.metricsPort).toBe(9000);
  });

  it('reads METRICS_PORT from the environment', async () => {
    process.env.METRICS_PORT = '9000';
    const { config } = await loadConfig();
    expect(config.metricsPort).toBe(9000);
  });

  it('does not throw when LOG_LEVEL is Debug', async () => {
    process.env.LOG_LEVEL = 'Debug';
    const { config } = await loadConfig();
    expect(config.logLevel).toBe('debug');
  });
});
