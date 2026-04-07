import nock from 'nock';
import { toolRegistry } from '../../src/tools/index';
import '../../src/tools/get-base-templates';

describe('get_base_templates tool', () => {
  afterEach(() => {
    nock.cleanAll();
  });

  it('should fetch base templates without authentication', async () => {
    const mockResponse = {
      data: [
        {
          name: 'default-dashboard',
          displayName: 'Default Dashboard',
          templateConfig: {
            sm: [],
            md: [],
            lg: [],
            xl: [],
          },
        },
      ],
      meta: { count: 1 },
    };

    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/base-templates')
      .reply(200, mockResponse);

    const result = await toolRegistry.execute('get_base_templates', {}, null);

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);

    const data = JSON.parse(result.content[0].text);
    expect(data.data).toHaveLength(1);
    expect(data.data[0].name).toBe('default-dashboard');
  });

  it('should handle API errors', async () => {
    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/base-templates')
      .reply(500, { errors: [{ code: 500, message: 'Internal Server Error' }] });

    const result = await toolRegistry.execute('get_base_templates', {}, null);

    expect(result.isError).toBe(true);
  });

  it('should return proper tool definition', () => {
    const tools = toolRegistry.getAll();
    const tool = tools.find((t) => t.name === 'get_base_templates');

    expect(tool).toBeDefined();
    expect(tool?.description).toContain('base dashboard templates');
  });
});
