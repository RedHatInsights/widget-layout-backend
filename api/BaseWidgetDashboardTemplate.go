package api

type BaseWidgetDashboardTemplateRegistry struct {
	BaseWidgetDashboardTemplates map[string]BaseWidgetDashboardTemplate `json:"baseWidgetDashboardTemplates" yaml:"baseWidgetDashboardTemplates"` // List of base widget dashboard templates
}

func (b *BaseWidgetDashboardTemplate) ToDashboardTemplate() DashboardTemplate {
	return DashboardTemplate{
		TemplateBase: DashboardTemplateBase{
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

func (br *BaseWidgetDashboardTemplateRegistry) GetAllBases() map[string]BaseWidgetDashboardTemplate {
	if br.BaseWidgetDashboardTemplates == nil {
		return make(map[string]BaseWidgetDashboardTemplate)
	}
	return br.BaseWidgetDashboardTemplates
}
