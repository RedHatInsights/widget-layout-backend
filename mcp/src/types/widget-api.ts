// Widget Layout API Response Types (based on OpenAPI spec)

export interface WidgetItem {
  w: number;
  h: number;
  maxH?: number;
  minH?: number;
  x: number | null;
  y: number | null;
  i: string;
  static?: boolean;
}

export interface DashboardTemplateConfig {
  sm: WidgetItem[];
  md: WidgetItem[];
  lg: WidgetItem[];
  xl: WidgetItem[];
}

export interface DashboardTemplateBase {
  name: string;
  displayName: string;
}

export interface DashboardTemplate {
  id: number;
  userId: string;
  dashboardName: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
  templateConfig: DashboardTemplateConfig;
  templateBase: DashboardTemplateBase;
  default?: boolean;
}

export interface ListResponseMeta {
  count: number;
}

export interface DashboardTemplateListResponse {
  data: DashboardTemplate[];
  meta: ListResponseMeta;
}

export interface BaseWidgetDashboardTemplate {
  name: string;
  displayName: string;
  templateConfig: DashboardTemplateConfig;
}

export interface BaseWidgetDashboardTemplateListResponse {
  data: BaseWidgetDashboardTemplate[];
  meta: ListResponseMeta;
}

export interface WidgetHeaderLink {
  title: string;
  href: string;
}

export interface Permission {
  method: string;
  args?: unknown[];
}

export interface WidgetConfiguration {
  title: string;
  icon?: string;
  headerLink?: WidgetHeaderLink;
  permissions?: Permission[];
}

export interface WidgetBaseDimensions {
  w: number | null;
  h: number | null;
  maxH?: number;
  minH?: number;
}

export interface WidgetModuleFederationMetadata {
  scope: string;
  module: string;
  importName?: string;
  featureFlag?: string;
  config: WidgetConfiguration;
  defaults: WidgetBaseDimensions;
}

export interface WidgetMappingResponse {
  data: Record<string, WidgetModuleFederationMetadata>;
}

export interface ExportWidgetDashboardTemplateResponse {
  templateConfig: DashboardTemplateConfig;
  templateBase: DashboardTemplateBase;
}

export interface ErrorPayload {
  code: number;
  message: string;
}

export interface ErrorResponse {
  errors: ErrorPayload[];
}
