import { toolRegistry } from '../../src/tools/index';
import '../../src/tools/hello';

describe('hello tool', () => {
  it('should execute successfully without authentication', async () => {
    const result = await toolRegistry.execute('hello', {}, null);

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);
    expect(result.content[0].type).toBe('text');

    const data = JSON.parse(result.content[0].text);
    expect(data).toHaveProperty('message');
    expect(data).toHaveProperty('status', 'healthy');
    expect(data).toHaveProperty('timestamp');
    expect(data).toHaveProperty('version');
  });

  it('should return proper tool definition', () => {
    const tools = toolRegistry.getAll();
    const helloTool = tools.find((t) => t.name === 'hello');

    expect(helloTool).toBeDefined();
    expect(helloTool?.description).toContain('Health check');
    expect(helloTool?.inputSchema.type).toBe('object');
  });
});
