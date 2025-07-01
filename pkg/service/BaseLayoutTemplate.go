package service

import (
	"encoding/json"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/sirupsen/logrus"
)

var BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

// LoadBaseTemplatesFromConfig loads base widget dashboard templates from config string.
func LoadBaseTemplatesFromConfig(configString string) error {
	if configString == "" {
		return nil
	}
	var baseTemplates []api.BaseWidgetDashboardTemplate
	err := json.Unmarshal([]byte(configString), &baseTemplates)
	if err != nil {
		return err
	}
	for _, bt := range baseTemplates {
		BaseTemplateRegistry.AddBase(bt)
	}
	logrus.Infof("Loaded %d base widget dashboard templates", len(BaseTemplateRegistry.BaseWidgetDashboardTemplates))
	return nil
}

func init() {
	cfg := config.GetConfig()
	if err := LoadBaseTemplatesFromConfig(cfg.BaseWidgetDashboardTemplates); err != nil {
		logrus.Fatalln("Failed to parse base widget dashboard templates, shutting down the service", err)
	}
}
