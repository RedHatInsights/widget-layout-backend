package api

import "fmt"

type WidgetMappingRegistry struct {
	WidgetMappings map[string]WidgetModuleFederationMetadata `json:"widgetMappings" yaml:"widgetMappings"`
}

type WidgetMappingResponse struct {
	Data map[string]WidgetModuleFederationMetadata `json:"data"`
}

func (wc *WidgetModuleFederationMetadata) GetWidgetKey() string {
	// make key in format "scope-module<-importName>"
	// This will be always unique
	// a module cannot be exposed with the same combination of the attributes values
	key := fmt.Sprintf("%s-%s", wc.Scope, wc.Module)
	if wc.ImportName != nil && *wc.ImportName != "" {
		key = fmt.Sprintf("%s-%s", key, *wc.ImportName)
	}
	return key
}

func (wmr *WidgetMappingRegistry) AddWidgetMapping(wc WidgetModuleFederationMetadata) {
	if wmr.WidgetMappings == nil {
		wmr.WidgetMappings = make(map[string]WidgetModuleFederationMetadata)
	}
	wmr.WidgetMappings[wc.GetWidgetKey()] = wc
}

func (wmr *WidgetMappingRegistry) GetWidgetMapping(name string) (WidgetModuleFederationMetadata, bool) {
	wm, exists := wmr.WidgetMappings[name]
	return wm, exists
}

func (wmr *WidgetMappingRegistry) GetAllWidgetMappings() map[string]WidgetModuleFederationMetadata {
	if wmr.WidgetMappings == nil {
		return make(map[string]WidgetModuleFederationMetadata)
	}
	return wmr.WidgetMappings
}
