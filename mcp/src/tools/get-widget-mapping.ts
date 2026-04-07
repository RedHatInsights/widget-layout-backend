import { toolRegistry } from './index';
import { apiClient } from '../utils/api-client';
import { WidgetMappingResponse } from '../types/widget-api';

toolRegistry.register({
  name: 'get_widget_mapping',
  description: 'Get the widget registry/catalog containing all available widgets with their module federation metadata, configuration, permissions, and default dimensions. No authentication required.',
  inputSchema: {
    type: 'object',
    properties: {},
  },
  requiresAuth: false,
  execute: async () => {
    const response = await apiClient.get<WidgetMappingResponse>(
      '/api/widget-layout/v1/widget-mapping'
    );

    return response;
  },
});
