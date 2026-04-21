import nock from 'nock';
import { toolRegistry } from '../../src/tools/index';
import { Identity } from '../../src/types/identity';
import '../../src/tools/get-layout-by-id';

const mockIdentity: Identity = {
  org_id: '12345',
  rawHeader: 'base64encodedheader',
};

describe('get_widget_layout_by_id tool', () => {
  afterEach(() => {
    nock.cleanAll();
  });

  it('should fetch layout by ID with valid identity', async () => {
    const mockResponse = {
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
    };

    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/1')
      .reply(200, mockResponse);

    const result = await toolRegistry.execute(
      'get_widget_layout_by_id',
      { dashboard_template_id: 1 },
      mockIdentity
    );

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);

    const data = JSON.parse(result.content[0].text);
    expect(data.id).toBe(1);
    expect(data.dashboardName).toBe('Dashboard 1');
  });

  it('should reject without authentication', async () => {
    const result = await toolRegistry.execute(
      'get_widget_layout_by_id',
      { dashboard_template_id: 1 },
      null
    );

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Authentication required');
  });

  it('should require dashboard_template_id parameter', async () => {
    const result = await toolRegistry.execute('get_widget_layout_by_id', {}, mockIdentity);

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Error');
  });

  it('should handle 404 not found', async () => {
    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/999')
      .reply(404, { errors: [{ code: 404, message: 'Not found' }] });

    const result = await toolRegistry.execute(
      'get_widget_layout_by_id',
      { dashboard_template_id: 999 },
      mockIdentity
    );

    expect(result.isError).toBe(true);
  });
});
