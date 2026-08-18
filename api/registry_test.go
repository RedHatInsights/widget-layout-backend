package api_test

import (
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestBaseWidgetDashboardTemplateRegistry(t *testing.T) {
	t.Run("AddBase should initialize map and add template", func(t *testing.T) {
		registry := api.BaseWidgetDashboardTemplateRegistry{}
		bt := api.BaseWidgetDashboardTemplate{
			Name:        "test-base",
			DisplayName: "Test Base",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		registry.AddBase(bt)

		result, exists := registry.GetBase("test-base")
		assert.True(t, exists)
		assert.Equal(t, "test-base", result.Name)
		assert.Equal(t, "Test Base", result.DisplayName)
	})

	t.Run("GetBase should return false for non-existent template", func(t *testing.T) {
		registry := api.BaseWidgetDashboardTemplateRegistry{}
		registry.AddBase(api.BaseWidgetDashboardTemplate{
			Name:        "existing",
			DisplayName: "Existing",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		})

		_, exists := registry.GetBase("non-existent")
		assert.False(t, exists)
	})

	t.Run("GetAllBases should return all templates", func(t *testing.T) {
		registry := api.BaseWidgetDashboardTemplateRegistry{}
		emptyConfig := api.DashboardTemplateConfig{
			Sm: datatypes.NewJSONType([]api.WidgetItem{}),
			Md: datatypes.NewJSONType([]api.WidgetItem{}),
			Lg: datatypes.NewJSONType([]api.WidgetItem{}),
			Xl: datatypes.NewJSONType([]api.WidgetItem{}),
		}

		registry.AddBase(api.BaseWidgetDashboardTemplate{Name: "base-1", DisplayName: "Base 1", TemplateConfig: emptyConfig})
		registry.AddBase(api.BaseWidgetDashboardTemplate{Name: "base-2", DisplayName: "Base 2", TemplateConfig: emptyConfig})

		all := registry.GetAllBases()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "base-1")
		assert.Contains(t, all, "base-2")
	})

	t.Run("GetAllBases should return empty map when nil", func(t *testing.T) {
		registry := api.BaseWidgetDashboardTemplateRegistry{}
		all := registry.GetAllBases()
		assert.NotNil(t, all)
		assert.Len(t, all, 0)
	})

	t.Run("AddBase should overwrite existing template with same name", func(t *testing.T) {
		registry := api.BaseWidgetDashboardTemplateRegistry{}
		emptyConfig := api.DashboardTemplateConfig{
			Sm: datatypes.NewJSONType([]api.WidgetItem{}),
			Md: datatypes.NewJSONType([]api.WidgetItem{}),
			Lg: datatypes.NewJSONType([]api.WidgetItem{}),
			Xl: datatypes.NewJSONType([]api.WidgetItem{}),
		}

		registry.AddBase(api.BaseWidgetDashboardTemplate{Name: "overwrite-test", DisplayName: "Original", TemplateConfig: emptyConfig})
		registry.AddBase(api.BaseWidgetDashboardTemplate{Name: "overwrite-test", DisplayName: "Updated", TemplateConfig: emptyConfig})

		result, exists := registry.GetBase("overwrite-test")
		assert.True(t, exists)
		assert.Equal(t, "Updated", result.DisplayName)

		all := registry.GetAllBases()
		assert.Len(t, all, 1)
	})
}

func TestToDashboardTemplate(t *testing.T) {
	t.Run("should convert base template to dashboard template", func(t *testing.T) {
		bt := api.BaseWidgetDashboardTemplate{
			Name:        "convert-test",
			DisplayName: "Convert Test",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{Width: 1, Height: 2, WidgetType: "widget-1", X: intPtr(0), Y: intPtr(0)},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		dt := bt.ToDashboardTemplate()

		assert.Equal(t, "convert-test", dt.TemplateBase.Name)
		assert.Equal(t, "Convert Test", dt.TemplateBase.DisplayName)

		smWidgets := dt.TemplateConfig.Sm.Data()
		require.Len(t, smWidgets, 1)
		assert.Equal(t, "widget-1", smWidgets[0].WidgetType)
		assert.Equal(t, uint(0), dt.ID, "Converted template should have zero ID")
	})
}

func TestWidgetMappingRegistry(t *testing.T) {
	t.Run("AddWidgetMapping should initialize map and add mapping", func(t *testing.T) {
		registry := api.WidgetMappingRegistry{}
		mapping := api.WidgetModuleFederationMetadata{
			Scope:  "testScope",
			Module: "./TestModule",
		}

		registry.AddWidgetMapping(mapping)

		result, exists := registry.GetWidgetMapping("testScope-./TestModule")
		assert.True(t, exists)
		assert.Equal(t, "testScope", result.Scope)
		assert.Equal(t, "./TestModule", result.Module)
	})

	t.Run("GetWidgetMapping should return false for non-existent mapping", func(t *testing.T) {
		registry := api.WidgetMappingRegistry{}
		registry.AddWidgetMapping(api.WidgetModuleFederationMetadata{
			Scope:  "scope1",
			Module: "./Module1",
		})

		_, exists := registry.GetWidgetMapping("non-existent-key")
		assert.False(t, exists)
	})

	t.Run("GetAllWidgetMappings should return all mappings", func(t *testing.T) {
		registry := api.WidgetMappingRegistry{}
		registry.AddWidgetMapping(api.WidgetModuleFederationMetadata{Scope: "scope1", Module: "./Mod1"})
		registry.AddWidgetMapping(api.WidgetModuleFederationMetadata{Scope: "scope2", Module: "./Mod2"})

		all := registry.GetAllWidgetMappings()
		assert.Len(t, all, 2)
	})

	t.Run("GetAllWidgetMappings should return empty map when nil", func(t *testing.T) {
		registry := api.WidgetMappingRegistry{}
		all := registry.GetAllWidgetMappings()
		assert.NotNil(t, all)
		assert.Len(t, all, 0)
	})
}

func TestGetWidgetKey(t *testing.T) {
	t.Run("should generate key without importName", func(t *testing.T) {
		wm := api.WidgetModuleFederationMetadata{
			Scope:  "myScope",
			Module: "./MyModule",
		}
		assert.Equal(t, "myScope-./MyModule", wm.GetWidgetKey())
	})

	t.Run("should include importName in key when set", func(t *testing.T) {
		importName := "customImport"
		wm := api.WidgetModuleFederationMetadata{
			Scope:      "myScope",
			Module:     "./MyModule",
			ImportName: &importName,
		}
		assert.Equal(t, "myScope-./MyModule-customImport", wm.GetWidgetKey())
	})

	t.Run("should not include empty importName in key", func(t *testing.T) {
		emptyImport := ""
		wm := api.WidgetModuleFederationMetadata{
			Scope:      "myScope",
			Module:     "./MyModule",
			ImportName: &emptyImport,
		}
		assert.Equal(t, "myScope-./MyModule", wm.GetWidgetKey())
	})
}
