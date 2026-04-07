import nock from 'nock';
import { toolRegistry } from '../../src/tools/index';
import '../../src/tools/get-widget-mapping';

describe('get_widget_mapping tool', () => {
  afterEach(() => {
    nock.cleanAll();
  });

  it('should fetch widget mapping without authentication', async () => {
    const mockResponse = {
      data: {
        'widget-1': {
          scope: 'scope1',
          module: 'module1',
          config: {
            title: 'Widget 1',
          },
          defaults: {
            w: 2,
            h: 2,
          },
        },
      },
    };

    const scope = nock('http://localhost:8000')
      .get('/api/widget-layout/v1/widget-mapping')
      .reply(200, mockResponse);

    const result = await toolRegistry.execute('get_widget_mapping', {}, null);

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);

    const data = JSON.parse(result.content[0].text);
    expect(data.data).toHaveProperty('widget-1');
    expect(data.data['widget-1'].scope).toBe('scope1');
    expect(scope.isDone()).toBe(true);
  });

  it('should handle API errors', async () => {
    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/widget-mapping')
      .reply(500, { errors: [{ code: 500, message: 'Internal Server Error' }] });

    const result = await toolRegistry.execute('get_widget_mapping', {}, null);

    expect(result.isError).toBe(true);
  });

  it('should return proper tool definition', () => {
    const tools = toolRegistry.getAll();
    const tool = tools.find((t) => t.name === 'get_widget_mapping');

    expect(tool).toBeDefined();
    expect(tool?.description).toContain('widget registry');
  });
});
