package models

import "github.com/RedHatInsights/widget-layout-backend/api"

type BaseWidgetDashboardTemplate struct {
	Name           string                      `json:"name" yaml:"name"`                                   // The name of the dashboard template
	DisplayName    string                      `json:"displayName" yaml:"displayName"`                     // The display name of the dashboard template
	TemplateConfig api.DashboardTemplateConfig `json:"templateConfig" yaml:"templateConfig"`               // The configuration of the dashboard template
	FrontendRef    string                      `json:"frontendRef,omitempty" yaml:"frontendRef,omitempty"` // The frontend reference for the dashboard template
}
