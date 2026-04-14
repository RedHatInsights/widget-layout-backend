import nock from 'nock';
import { toolRegistry } from '../../src/tools/index';
import { Identity } from '../../src/types/identity';
import '../../src/tools/export-layout';

const mockIdentity: Identity = {
  org_id: '12345',
  rawHeader: 'base64encodedheader',
};

describe('export_widget_layout tool', () => {
  afterEach(() => {
    nock.cleanAll();
  });

  it('should export layout with valid identity', async () => {
    const mockResponse = {
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
      .get('/api/widget-layout/v1/1/export')
      .reply(200, mockResponse);

    const result = await toolRegistry.execute(
      'export_widget_layout',
      { dashboard_template_id: 1 },
      mockIdentity
    );

    expect(result.isError).toBeUndefined();
    expect(result.content).toHaveLength(1);

    const data = JSON.parse(result.content[0].text);
    expect(data).toHaveProperty('templateConfig');
    expect(data).toHaveProperty('templateBase');
  });

  it('should reject without authentication', async () => {
    const result = await toolRegistry.execute(
      'export_widget_layout',
      { dashboard_template_id: 1 },
      null
    );

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Authentication required');
  });

  it('should require dashboard_template_id parameter', async () => {
    const result = await toolRegistry.execute('export_widget_layout', {}, mockIdentity);

    expect(result.isError).toBe(true);
    expect(result.content[0].text).toContain('Error');
  });

  it('should handle 403 forbidden', async () => {
    nock('http://localhost:8000')
      .get('/api/widget-layout/v1/1/export')
      .reply(403, { errors: [{ code: 403, message: 'Forbidden' }] });

    const result = await toolRegistry.execute(
      'export_widget_layout',
      { dashboard_template_id: 1 },
      mockIdentity
    );

    expect(result.isError).toBe(true);
  });
});
