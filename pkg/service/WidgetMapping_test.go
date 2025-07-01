package service_test

import (
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWidgetMappings(t *testing.T) {
	t.Run("should return empty map when no widget mappings exist", func(t *testing.T) {
		// Reset registry to ensure clean state
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		mappings := service.GetWidgetMappings()

		assert.NotNil(t, mappings, "Should return a non-nil map")
		assert.Empty(t, mappings, "Should return empty map when no mappings exist")
	})

	t.Run("should return all widget mappings from registry", func(t *testing.T) {
		// Reset registry and add test widget mappings
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create test widget mappings
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:  "test-scope-1",
			Module: "test-module-1",
			Config: api.WidgetConfiguration{
				Title: "Test Widget 1",
				Icon:  "test-icon-1",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(6),
				MinHeight: test_util.IntPTR(1),
			},
		}

		widget2 := api.WidgetModuleFederationMetadata{
			Scope:  "test-scope-2",
			Module: "test-module-2",
			Config: api.WidgetConfiguration{
				Title: "Test Widget 2",
				Icon:  "test-icon-2",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(1),
				Height:    test_util.IntPTR(2),
				MaxHeight: test_util.IntPTR(4),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Add widgets to registry
		service.WidgetMappingRegistry.AddWidgetMapping(widget1)
		service.WidgetMappingRegistry.AddWidgetMapping(widget2)

		mappings := service.GetWidgetMappings()

		assert.NotNil(t, mappings, "Should return a non-nil map")
		assert.Len(t, mappings, 2, "Should return both widget mappings")

		// Verify first widget
		widget1Key := widget1.GetWidgetKey()
		assert.Contains(t, mappings, widget1Key, "Should contain first widget mapping")
		retrievedWidget1 := mappings[widget1Key]
		assert.Equal(t, "test-scope-1", retrievedWidget1.Scope, "First widget scope should match")
		assert.Equal(t, "test-module-1", retrievedWidget1.Module, "First widget module should match")
		assert.Equal(t, "Test Widget 1", retrievedWidget1.Config.Title, "First widget title should match")

		// Verify second widget
		widget2Key := widget2.GetWidgetKey()
		assert.Contains(t, mappings, widget2Key, "Should contain second widget mapping")
		retrievedWidget2 := mappings[widget2Key]
		assert.Equal(t, "test-scope-2", retrievedWidget2.Scope, "Second widget scope should match")
		assert.Equal(t, "test-module-2", retrievedWidget2.Module, "Second widget module should match")
		assert.Equal(t, "Test Widget 2", retrievedWidget2.Config.Title, "Second widget title should match")
	})

	t.Run("should handle widget mapping with import name", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		importName := "CustomImport"
		widget := api.WidgetModuleFederationMetadata{
			Scope:      "test-scope",
			Module:     "test-module",
			ImportName: &importName,
			Config: api.WidgetConfiguration{
				Title: "Test Widget with Import",
				Icon:  "test-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(3),
				Height:    test_util.IntPTR(4),
				MaxHeight: test_util.IntPTR(8),
				MinHeight: test_util.IntPTR(2),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		mappings := service.GetWidgetMappings()

		assert.Len(t, mappings, 1, "Should contain one widget mapping")

		// Verify the widget key includes import name
		expectedKey := "test-scope-test-module-CustomImport"
		assert.Contains(t, mappings, expectedKey, "Should contain widget with import name in key")
		retrievedWidget := mappings[expectedKey]
		assert.Equal(t, "test-scope", retrievedWidget.Scope, "Widget scope should match")
		assert.Equal(t, "test-module", retrievedWidget.Module, "Widget module should match")
		assert.Equal(t, "CustomImport", *retrievedWidget.ImportName, "Widget import name should match")
	})

	t.Run("should handle widget mapping with feature flag", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		featureFlag := "enable-test-feature"
		widget := api.WidgetModuleFederationMetadata{
			Scope:       "test-scope",
			Module:      "test-module",
			FeatureFlag: &featureFlag,
			Config: api.WidgetConfiguration{
				Title: "Test Widget with Feature Flag",
				Icon:  "test-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(2),
				MaxHeight: test_util.IntPTR(5),
				MinHeight: test_util.IntPTR(1),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		mappings := service.GetWidgetMappings()

		assert.Len(t, mappings, 1, "Should contain one widget mapping")

		widgetKey := widget.GetWidgetKey()
		retrievedWidget := mappings[widgetKey]
		assert.Equal(t, "enable-test-feature", *retrievedWidget.FeatureFlag, "Widget feature flag should match")
	})

	t.Run("should handle widget mapping with header link", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		widget := api.WidgetModuleFederationMetadata{
			Scope:  "test-scope",
			Module: "test-module",
			Config: api.WidgetConfiguration{
				Title: "Test Widget with Header Link",
				Icon:  "test-icon",
				HeaderLink: &api.WidgetHeaderLink{
					Name: "View Details",
					Href: "https://example.com/details",
				},
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(4),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(7),
				MinHeight: test_util.IntPTR(2),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		mappings := service.GetWidgetMappings()

		assert.Len(t, mappings, 1, "Should contain one widget mapping")

		widgetKey := widget.GetWidgetKey()
		retrievedWidget := mappings[widgetKey]
		require.NotNil(t, retrievedWidget.Config.HeaderLink, "Widget should have header link")
		assert.Equal(t, "View Details", retrievedWidget.Config.HeaderLink.Name, "Header link name should match")
		assert.Equal(t, "https://example.com/details", retrievedWidget.Config.HeaderLink.Href, "Header link href should match")
	})

	t.Run("should handle widget mapping with permissions", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		permissions := []string{"read:widgets", "write:widgets", "admin:widgets"}
		widget := api.WidgetModuleFederationMetadata{
			Scope:  "test-scope",
			Module: "test-module",
			Config: api.WidgetConfiguration{
				Title:       "Test Widget with Permissions",
				Icon:        "test-icon",
				Permissions: &permissions,
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(3),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(6),
				MinHeight: test_util.IntPTR(1),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		mappings := service.GetWidgetMappings()

		assert.Len(t, mappings, 1, "Should contain one widget mapping")

		widgetKey := widget.GetWidgetKey()
		retrievedWidget := mappings[widgetKey]
		assert.Equal(t, &permissions, retrievedWidget.Config.Permissions, "Widget permissions should match")
		assert.Len(t, *retrievedWidget.Config.Permissions, 3, "Should have all permissions")
	})

	t.Run("should handle duplicate widget keys by overwriting previous widget", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create first widget
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:  "duplicate-scope",
			Module: "duplicate-module",
			Config: api.WidgetConfiguration{
				Title: "First Widget",
				Icon:  "first-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(1),
				Height:    test_util.IntPTR(1),
				MaxHeight: test_util.IntPTR(3),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Create second widget with same scope/module (will have same key)
		widget2 := api.WidgetModuleFederationMetadata{
			Scope:  "duplicate-scope",
			Module: "duplicate-module",
			Config: api.WidgetConfiguration{
				Title: "Second Widget",
				Icon:  "second-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(2),
				MaxHeight: test_util.IntPTR(4),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Verify they have the same key
		assert.Equal(t, widget1.GetWidgetKey(), widget2.GetWidgetKey(), "Widgets should have the same key")

		// Add first widget
		service.WidgetMappingRegistry.AddWidgetMapping(widget1)
		mappings := service.GetWidgetMappings()
		assert.Len(t, mappings, 1, "Should have one widget after adding first widget")

		// Verify first widget is in registry
		widgetKey := widget1.GetWidgetKey()
		retrievedWidget := mappings[widgetKey]
		assert.Equal(t, "First Widget", retrievedWidget.Config.Title, "Should have first widget title")
		assert.Equal(t, "first-icon", retrievedWidget.Config.Icon, "Should have first widget icon")

		// Add second widget with same key (should overwrite first)
		service.WidgetMappingRegistry.AddWidgetMapping(widget2)
		mappings = service.GetWidgetMappings()
		assert.Len(t, mappings, 1, "Should still have one widget after adding duplicate")

		// Verify second widget has overwritten first
		retrievedWidget = mappings[widgetKey]
		assert.Equal(t, "Second Widget", retrievedWidget.Config.Title, "Should have second widget title (overwritten)")
		assert.Equal(t, "second-icon", retrievedWidget.Config.Icon, "Should have second widget icon (overwritten)")
		assert.Equal(t, 2, *retrievedWidget.Defaults.Width, "Should have second widget width (overwritten)")
		assert.Equal(t, 2, *retrievedWidget.Defaults.Height, "Should have second widget height (overwritten)")
	})

	t.Run("should handle duplicate keys with different import names", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create first widget without import name
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:  "same-scope",
			Module: "same-module",
			Config: api.WidgetConfiguration{
				Title: "Widget Without Import",
				Icon:  "no-import-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(1),
				Height:    test_util.IntPTR(1),
				MaxHeight: test_util.IntPTR(3),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Create second widget with import name (different key)
		importName := "CustomImport"
		widget2 := api.WidgetModuleFederationMetadata{
			Scope:      "same-scope",
			Module:     "same-module",
			ImportName: &importName,
			Config: api.WidgetConfiguration{
				Title: "Widget With Import",
				Icon:  "import-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(2),
				MaxHeight: test_util.IntPTR(4),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Verify they have different keys
		assert.NotEqual(t, widget1.GetWidgetKey(), widget2.GetWidgetKey(), "Widgets should have different keys due to import name")

		// Add both widgets
		service.WidgetMappingRegistry.AddWidgetMapping(widget1)
		service.WidgetMappingRegistry.AddWidgetMapping(widget2)

		mappings := service.GetWidgetMappings()
		assert.Len(t, mappings, 2, "Should have two widgets with different keys")

		// Verify both widgets exist
		widget1Key := widget1.GetWidgetKey()
		widget2Key := widget2.GetWidgetKey()

		retrievedWidget1 := mappings[widget1Key]
		assert.Equal(t, "Widget Without Import", retrievedWidget1.Config.Title, "First widget should exist")
		assert.Nil(t, retrievedWidget1.ImportName, "First widget should not have import name")

		retrievedWidget2 := mappings[widget2Key]
		assert.Equal(t, "Widget With Import", retrievedWidget2.Config.Title, "Second widget should exist")
		assert.Equal(t, "CustomImport", *retrievedWidget2.ImportName, "Second widget should have import name")
	})
}
