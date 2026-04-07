import { z } from 'zod';
import { toolRegistry } from './index';
import { apiClient } from '../utils/api-client';
import { Identity } from '../types/identity';
import { DashboardTemplate } from '../types/widget-api';

const GetLayoutByIdArgsSchema = z.object({
  dashboard_template_id: z.number().int().positive(),
});

toolRegistry.register({
  name: 'get_widget_layout_by_id',
  description: 'Get a specific dashboard template by ID. Returns detailed configuration including widget positions and sizes for all breakpoints (sm, md, lg, xl).',
  inputSchema: {
    type: 'object',
    properties: {
      dashboard_template_id: {
        type: 'integer',
        description: 'The unique identifier of the dashboard template',
        minimum: 1,
      },
    },
    required: ['dashboard_template_id'],
  },
  requiresAuth: true,
  execute: async (args, identity) => {
    const params = GetLayoutByIdArgsSchema.parse(args);

    const response = await apiClient.get<DashboardTemplate>(
      `/api/widget-layout/v1/${params.dashboard_template_id}`,
      {
        headers: {
          'x-rh-identity': (identity as Identity).rawHeader,
        },
      }
    );

    return response;
  },
});
