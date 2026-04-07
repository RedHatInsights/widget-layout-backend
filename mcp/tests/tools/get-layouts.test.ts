import nock from 'nock';
import { toolRegistry } from '../../src/tools/index';
import { Identity } from '../../src/types/identity';
import '../../src/tools/get-layouts';

const mockIdentity: Identity = {
  org_id: '12345',
  rawHeader: 'base64encodedheader',
};

describe('get_widget_layouts tool', () => {
  afterEach(() => {
    nock.cleanAll();
  });

  it('should fetch layouts with valid identity', async () => {
    const mockResponse = {
      data: [
        {
          id: 1,
          userId: 'user1',
          dashboardName: 'Dashboard 1',
          createdAt: '2024-01-01T00:00:00Z',
          updatedAt: '2024-01-01T00:00:00Z',
          templateConfig: {
            sm: [],
            md: [],
            lg: [],
            xl: [],
          },
          templateBase: {
            name: 'dashboard-1',
            displayName: 'Dashboard 1',
          },
        },
      ],
      meta: { count: 1 },
    };

    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/')
      .reply(200, mockResponse);

    const result = await toolRegistry.execute('get_widget_layouts', {}, mockIdentity);

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);

    const data = JSON.parse(result.content[0].text);
    expect(data.data).toHaveLength(1);
    expect(data.meta.count).toBe(1);
  });

  it('should reject without authentication', async () => {
    const result = await toolRegistry.execute('get_widget_layouts', {}, null);

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Authentication required');
  });

  it('should handle API errors', async () => {
    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/')
      .reply(500, { errors: [{ code: 500, message: 'Internal Server Error' }] });

    const result = await toolRegistry.execute('get_widget_layouts', {}, mockIdentity);

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Error');
  });

  it('should pass dashboardType filter parameter', async () => {
    const mockResponse = {
      data: [],
      meta: { count: 0 },
    };

    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/')
      .query({ dashboardType: 'analytics' })
      .reply(200, mockResponse);

    const result = await toolRegistry.execute(
      'get_widget_layouts',
      { dashboardType: 'analytics' },
      mockIdentity
    );

    expect(result.isError).toBeUndefined();
  });
});
