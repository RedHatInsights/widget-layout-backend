package service

import (
	"encoding/json"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/sirupsen/logrus"
)

var WidgetMappingRegistry = api.WidgetMappingRegistry{
	WidgetMappings: make(map[string]api.WidgetModuleFederationMetadata),
}

func LoadWidgetMappingsFromConfig(configString string) error {
	if configString == "" {
		return nil
	}
	var widgetMappings []api.WidgetModuleFederationMetadata
	err := json.Unmarshal([]byte(configString), &widgetMappings)
	if err != nil {
		return err
	}
	for _, wm := range widgetMappings {
		WidgetMappingRegistry.AddWidgetMapping(wm)
	}
	logrus.Infof("Loaded %d widget mappings", len(WidgetMappingRegistry.WidgetMappings))
	return nil
}

func init() {
	cfg := config.GetConfig()
	if err := LoadWidgetMappingsFromConfig(cfg.WidgetMappingConfig); err != nil {
		logrus.Fatalln("Failed to parse widget mappings, shutting down the service", err)
	}
}

// GetWidgetMappings returns all widget mappings from the registry
func GetWidgetMappings() map[string]api.WidgetModuleFederationMetadata {
	mappings := WidgetMappingRegistry.GetAllWidgetMappings()
	logrus.Debugf("Retrieved %d widget mappings", len(mappings))
	return mappings
}
