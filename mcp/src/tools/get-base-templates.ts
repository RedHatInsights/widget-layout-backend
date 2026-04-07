import { toolRegistry } from './index';
import { apiClient } from '../utils/api-client';
import { BaseWidgetDashboardTemplateListResponse } from '../types/widget-api';

toolRegistry.register({
  name: 'get_base_templates',
  description: 'List all available base dashboard templates. These are predefined templates that users can fork to create their own custom dashboards. No authentication required.',
  inputSchema: {
    type: 'object',
    properties: {},
  },
  requiresAuth: false,
  execute: async () => {
    const response = await apiClient.get<BaseWidgetDashboardTemplateListResponse>(
      '/api/widget-layout/v1/base-templates'
    );

    return response;
  },
});
