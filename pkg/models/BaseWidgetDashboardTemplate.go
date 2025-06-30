package models

import "github.com/RedHatInsights/widget-layout-backend/api"

type BaseWidgetDashboardTemplate struct {
	Name           string                      `json:"name" yaml:"name"`                                   // The name of the dashboard template
	DisplayName    string                      `json:"displayName" yaml:"displayName"`                     // The display name of the dashboard template
	TemplateConfig api.DashboardTemplateConfig `json:"templateConfig" yaml:"templateConfig"`               // The configuration of the dashboard template
	FrontendRef    string                      `json:"frontendRef,omitempty" yaml:"frontendRef,omitempty"` // The frontend reference for the dashboard template
}

type BaseWidgetDashboardTemplateRegistry struct {
	BaseWidgetDashboardTemplates map[string]BaseWidgetDashboardTemplate `json:"baseWidgetDashboardTemplates" yaml:"baseWidgetDashboardTemplates"` // List of base widget dashboard templates
}

func (b *BaseWidgetDashboardTemplate) ToDashboardTemplate() api.DashboardTemplate {
	return api.DashboardTemplate{
		TemplateBase: api.DashboardTemplateBase{
			Name:        b.Name,
			DisplayName: b.DisplayName,
		},
		TemplateConfig: b.TemplateConfig,
	}
}

func (br *BaseWidgetDashboardTemplateRegistry) AddBase(bt BaseWidgetDashboardTemplate) {
	if br.BaseWidgetDashboardTemplates == nil {
		br.BaseWidgetDashboardTemplates = make(map[string]BaseWidgetDashboardTemplate)
	}
	br.BaseWidgetDashboardTemplates[bt.Name] = bt
}

func (br *BaseWidgetDashboardTemplateRegistry) GetBase(name string) (BaseWidgetDashboardTemplate, bool) {
	bt, exists := br.BaseWidgetDashboardTemplates[name]
	return bt, exists
}
