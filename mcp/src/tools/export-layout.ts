import { z } from 'zod';
import { toolRegistry } from './index';
import { apiClient } from '../utils/api-client';
import { Identity } from '../types/identity';
import { ExportWidgetDashboardTemplateResponse } from '../types/widget-api';

const ExportLayoutArgsSchema = z.object({
  dashboard_template_id: z.number().int().positive(),
});

toolRegistry.register({
  name: 'export_widget_layout',
  description: 'Export a dashboard template as a shareable configuration. Returns the template configuration and base information that can be used to recreate or share the dashboard.',
  inputSchema: {
    type: 'object',
    properties: {
      dashboard_template_id: {
        type: 'number',
        description: 'The unique identifier of the dashboard template to export (positive integer)',
      },
    },
    required: ['dashboard_template_id'],
  },
  requiresAuth: true,
  execute: async (args, identity) => {
    const params = ExportLayoutArgsSchema.parse(args);

    const response = await apiClient.get<ExportWidgetDashboardTemplateResponse>(
      `/api/widget-layout/v1/${params.dashboard_template_id}/export`,
      {
        headers: {
          'x-rh-identity': (identity as Identity).rawHeader,
        },
      }
    );

    return response;
  },
});
