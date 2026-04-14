import axios, { AxiosInstance, AxiosError, AxiosRequestConfig } from 'axios';
import { config } from '../config';
import { logger } from './logger';
import { mcpApiCallDuration } from './metrics';

// Normalize URL path to prevent metric cardinality explosion
// Replaces numeric/UUID segments with placeholders
function normalizeEndpoint(url: string): string {
  const path = url.split('?')[0];
  return path.replace(/\/\d+/g, '/:id').replace(/\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi, '/:uuid');
}

export class ApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string = config.widgetLayoutApiUrl) {
    this.client = axios.create({
      baseURL,
      timeout: 10000,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor for logging
    this.client.interceptors.request.use(
      (config) => {
        logger.debug(
          {
            method: config.method,
            url: config.url,
            params: config.params,
          },
          'mcp: API request'
        );
        return config;
      },
      (error) => {
        logger.error({ error }, 'mcp: API request error');
        return Promise.reject(error);
      }
    );

    // Response interceptor for logging and metrics
    this.client.interceptors.response.use(
      (response) => {
        logger.debug(
          {
            status: response.status,
            url: response.config.url,
          },
          'mcp: API response'
        );
        return response;
      },
      (error: AxiosError) => {
        logger.error(
          {
            status: error.response?.status,
            url: error.config?.url,
            message: error.message,
          },
          'mcp: API response error'
        );
        return Promise.reject(error);
      }
    );
  }

  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const startTime = Date.now();
    const endpoint = normalizeEndpoint(url);

    try {
      const response = await this.client.get<T>(url, config);
      const duration = (Date.now() - startTime) / 1000;
      mcpApiCallDuration.observe(
        { endpoint, method: 'GET', status: response.status.toString() },
        duration
      );
      return response.data;
    } catch (error) {
      const duration = (Date.now() - startTime) / 1000;
      const status = axios.isAxiosError(error) ? error.response?.status?.toString() || 'error' : 'error';
      mcpApiCallDuration.observe(
        { endpoint, method: 'GET', status },
        duration
      );
      throw error;
    }
  }
}

export const apiClient = new ApiClient();
