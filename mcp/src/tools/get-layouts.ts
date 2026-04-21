import { z } from 'zod';
import { toolRegistry } from './index';
import { apiClient } from '../utils/api-client';
import { Identity } from '../types/identity';
import { DashboardTemplateListResponse } from '../types/widget-api';

const GetLayoutsArgsSchema = z.object({
  dashboardType: z.string().optional(),
});

toolRegistry.register({
  name: 'get_widget_layouts',
  description: 'List all widget dashboard templates for the authenticated organization. Optionally filter by dashboard type.',
  inputSchema: {
    type: 'object',
    properties: {
      dashboardType: {
        type: 'string',
        description: 'Optional filter by dashboard type',
      },
    },
  },
  requiresAuth: true,
  execute: async (args, identity) => {
    const params = GetLayoutsArgsSchema.parse(args);

    const response = await apiClient.get<DashboardTemplateListResponse>(
      '/api/widget-layout/v1/',
      {
        params: params.dashboardType ? { dashboardType: params.dashboardType } : {},
        headers: {
          'x-rh-identity': (identity as Identity).rawHeader,
        },
      }
    );

    return response;
  },
});
