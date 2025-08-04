package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetWidgetMapping(t *testing.T) {
	t.Run("should return empty map when no widget mappings exist", func(t *testing.T) {
		// Reset registry to ensure clean state
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		server := setupRouter()
		req, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w := httptest.NewRecorder()

		server.GetWidgetMapping(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

		var response api.WidgetMappingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Should be able to unmarshal response")

		assert.NotNil(t, response, "Response should not be nil")
		assert.Empty(t, response.Data, "Response should be empty map when no mappings exist")
	})

	t.Run("should return all widget mappings when they exist", func(t *testing.T) {
		// Reset registry and add test widget mappings
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create test widget mappings
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:  "insights",
			Module: "dashboard-widget",
			Config: api.WidgetConfiguration{
				Title: "Dashboard Overview",
				Icon:  "dashboard-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(6),
				MinHeight: test_util.IntPTR(1),
			},
		}

		widget2 := api.WidgetModuleFederationMetadata{
			Scope:  "monitoring",
			Module: "alerts-widget",
			Config: api.WidgetConfiguration{
				Title: "Alert Status",
				Icon:  "alert-icon",
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

		server := setupRouter()
		req, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w := httptest.NewRecorder()

		server.GetWidgetMapping(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

		var response api.WidgetMappingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Should be able to unmarshal response")

		assert.Len(t, response.Data, 2, "Should return both widget mappings")

		// Verify first widget
		widget1Key := widget1.GetWidgetKey()
		assert.Contains(t, response.Data, widget1Key, "Should contain first widget mapping")
		retrievedWidget1 := response.Data[widget1Key]
		assert.Equal(t, "insights", retrievedWidget1.Scope, "First widget scope should match")
		assert.Equal(t, "dashboard-widget", retrievedWidget1.Module, "First widget module should match")
		assert.Equal(t, "Dashboard Overview", retrievedWidget1.Config.Title, "First widget title should match")
		assert.Equal(t, "dashboard-icon", retrievedWidget1.Config.Icon, "First widget icon should match")

		// Verify second widget
		widget2Key := widget2.GetWidgetKey()
		assert.Contains(t, response.Data, widget2Key, "Should contain second widget mapping")
		retrievedWidget2 := response.Data[widget2Key]
		assert.Equal(t, "monitoring", retrievedWidget2.Scope, "Second widget scope should match")
		assert.Equal(t, "alerts-widget", retrievedWidget2.Module, "Second widget module should match")
		assert.Equal(t, "Alert Status", retrievedWidget2.Config.Title, "Second widget title should match")
		assert.Equal(t, "alert-icon", retrievedWidget2.Config.Icon, "Second widget icon should match")
	})

	t.Run("should return widget mapping with all optional fields", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		importName := "CustomComponent"
		featureFlag := "enable-advanced-widgets"
		permissions := []api.Permission{
			api.Permission{
				Method: "isOrgAdmin",
			},
		}

		widget := api.WidgetModuleFederationMetadata{
			Scope:       "advanced",
			Module:      "complex-widget",
			ImportName:  &importName,
			FeatureFlag: &featureFlag,
			Config: api.WidgetConfiguration{
				Title:       "Advanced Widget",
				Icon:        "advanced-icon",
				Permissions: &permissions,
				HeaderLink: &api.WidgetHeaderLink{
					Name: "View Details",
					Href: "https://console.redhat.com/insights/dashboard",
				},
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(3),
				Height:    test_util.IntPTR(4),
				MaxHeight: test_util.IntPTR(8),
				MinHeight: test_util.IntPTR(2),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		server := setupRouter()
		req, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w := httptest.NewRecorder()

		server.GetWidgetMapping(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

		var response api.WidgetMappingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Should be able to unmarshal response")

		assert.Len(t, response.Data, 1, "Should contain one widget mapping")

		widgetKey := widget.GetWidgetKey()
		assert.Contains(t, response.Data, widgetKey, "Should contain the widget mapping")
		retrievedWidget := response.Data[widgetKey]

		// Verify basic fields
		assert.Equal(t, "advanced", retrievedWidget.Scope, "Widget scope should match")
		assert.Equal(t, "complex-widget", retrievedWidget.Module, "Widget module should match")
		assert.Equal(t, "CustomComponent", *retrievedWidget.ImportName, "Import name should match")
		assert.Equal(t, "enable-advanced-widgets", *retrievedWidget.FeatureFlag, "Feature flag should match")

		// Verify config
		assert.Equal(t, "Advanced Widget", retrievedWidget.Config.Title, "Widget title should match")
		assert.Equal(t, "advanced-icon", retrievedWidget.Config.Icon, "Widget icon should match")
		require.NotNil(t, retrievedWidget.Config.Permissions, "Permissions should not be nil")
		assert.Equal(t, &permissions, retrievedWidget.Config.Permissions, "Permissions should match")
		require.NotNil(t, retrievedWidget.Config.HeaderLink, "Header link should not be nil")
		assert.Equal(t, "View Details", retrievedWidget.Config.HeaderLink.Name, "Header link name should match")
		assert.Equal(t, "https://console.redhat.com/insights/dashboard", retrievedWidget.Config.HeaderLink.Href, "Header link href should match")

		// Verify defaults
		assert.Equal(t, 3, *retrievedWidget.Defaults.Width, "Default width should match")
		assert.Equal(t, 4, *retrievedWidget.Defaults.Height, "Default height should match")
		assert.Equal(t, 8, *retrievedWidget.Defaults.MaxHeight, "Default max height should match")
		assert.Equal(t, 2, *retrievedWidget.Defaults.MinHeight, "Default min height should match")
	})

	t.Run("should handle complex widget keys correctly", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create widgets with different key combinations
		importName1 := "ImportedComponent"
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:      "scope1",
			Module:     "module1",
			ImportName: &importName1,
			Config: api.WidgetConfiguration{
				Title: "Widget with Import",
				Icon:  "icon1",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(1),
				Height:    test_util.IntPTR(1),
				MaxHeight: test_util.IntPTR(3),
				MinHeight: test_util.IntPTR(1),
			},
		}

		widget2 := api.WidgetModuleFederationMetadata{
			Scope:  "scope1",
			Module: "module1",
			Config: api.WidgetConfiguration{
				Title: "Widget without Import",
				Icon:  "icon2",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(2),
				MaxHeight: test_util.IntPTR(4),
				MinHeight: test_util.IntPTR(1),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget1)
		service.WidgetMappingRegistry.AddWidgetMapping(widget2)

		server := setupRouter()
		req, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w := httptest.NewRecorder()

		server.GetWidgetMapping(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")

		var response api.WidgetMappingResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err, "Should be able to unmarshal response")

		assert.Len(t, response.Data, 2, "Should contain both widgets despite same scope/module")

		// Verify both widgets are present with correct keys
		widget1Key := "scope1-module1-ImportedComponent"
		widget2Key := "scope1-module1"

		assert.Contains(t, response.Data, widget1Key, "Should contain widget with import name in key")
		assert.Contains(t, response.Data, widget2Key, "Should contain widget without import name in key")

		// Verify widget1 (with import name)
		retrievedWidget1 := response.Data[widget1Key]
		assert.Equal(t, "Widget with Import", retrievedWidget1.Config.Title, "First widget title should match")
		assert.Equal(t, "ImportedComponent", *retrievedWidget1.ImportName, "Import name should match")

		// Verify widget2 (without import name)
		retrievedWidget2 := response.Data[widget2Key]
		assert.Equal(t, "Widget without Import", retrievedWidget2.Config.Title, "Second widget title should match")
		assert.Nil(t, retrievedWidget2.ImportName, "Import name should be nil")
	})

	t.Run("should return proper content type header", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		server := setupRouter()
		req, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w := httptest.NewRecorder()

		server.GetWidgetMapping(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Should return application/json content type")
	})

	t.Run("should preserve widget mapping data integrity", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Add widget with specific data
		widget := api.WidgetModuleFederationMetadata{
			Scope:  "test-scope",
			Module: "test-module",
			Config: api.WidgetConfiguration{
				Title: "Data Integrity Test",
				Icon:  "test-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(2),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(6),
				MinHeight: test_util.IntPTR(1),
			},
		}

		service.WidgetMappingRegistry.AddWidgetMapping(widget)

		// Make multiple requests to ensure data consistency
		server := setupRouter()

		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/widget-mapping", nil)
			w := httptest.NewRecorder()

			server.GetWidgetMapping(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Should return 200 OK on request %d", i+1)

			var response api.WidgetMappingResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err, "Should be able to unmarshal response on request %d", i+1)

			assert.Len(t, response.Data, 1, "Should always return one widget mapping on request %d", i+1)

			widgetKey := widget.GetWidgetKey()
			retrievedWidget := response.Data[widgetKey]
			assert.Equal(t, "Data Integrity Test", retrievedWidget.Config.Title, "Widget title should be consistent on request %d", i+1)
			assert.Equal(t, 2, *retrievedWidget.Defaults.Width, "Widget width should be consistent on request %d", i+1)
		}
	})

	t.Run("should return overwritten widget when duplicate keys are added", func(t *testing.T) {
		// Reset registry
		service.WidgetMappingRegistry = api.WidgetMappingRegistry{}

		// Create first widget
		widget1 := api.WidgetModuleFederationMetadata{
			Scope:  "duplicate-test",
			Module: "duplicate-module",
			Config: api.WidgetConfiguration{
				Title: "Original Widget",
				Icon:  "original-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(1),
				Height:    test_util.IntPTR(1),
				MaxHeight: test_util.IntPTR(3),
				MinHeight: test_util.IntPTR(1),
			},
		}

		// Create second widget with same key
		widget2 := api.WidgetModuleFederationMetadata{
			Scope:  "duplicate-test",
			Module: "duplicate-module",
			Config: api.WidgetConfiguration{
				Title: "Overwriting Widget",
				Icon:  "overwriting-icon",
			},
			Defaults: api.WidgetBaseDimensions{
				Width:     test_util.IntPTR(3),
				Height:    test_util.IntPTR(3),
				MaxHeight: test_util.IntPTR(6),
				MinHeight: test_util.IntPTR(2),
			},
		}

		// Add first widget
		service.WidgetMappingRegistry.AddWidgetMapping(widget1)

		server := setupRouter()
		req1, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w1 := httptest.NewRecorder()

		server.GetWidgetMapping(w1, req1)

		assert.Equal(t, http.StatusOK, w1.Code, "Should return 200 OK")

		var response1 api.WidgetMappingResponse
		err := json.Unmarshal(w1.Body.Bytes(), &response1)
		require.NoError(t, err, "Should be able to unmarshal first response")

		assert.Len(t, response1.Data, 1, "Should have one widget")
		widgetKey := widget1.GetWidgetKey()
		retrievedWidget1 := response1.Data[widgetKey]
		assert.Equal(t, "Original Widget", retrievedWidget1.Config.Title, "Should have original widget")

		// Add second widget with same key (overwrites first)
		service.WidgetMappingRegistry.AddWidgetMapping(widget2)

		req2, _ := http.NewRequest("GET", "/widget-mapping", nil)
		w2 := httptest.NewRecorder()

		server.GetWidgetMapping(w2, req2)

		assert.Equal(t, http.StatusOK, w2.Code, "Should return 200 OK")

		var response2 api.WidgetMappingResponse
		err = json.Unmarshal(w2.Body.Bytes(), &response2)
		require.NoError(t, err, "Should be able to unmarshal second response")

		assert.Len(t, response2.Data, 1, "Should still have one widget")
		retrievedWidget2 := response2.Data[widgetKey]
		assert.Equal(t, "Overwriting Widget", retrievedWidget2.Config.Title, "Should have overwritten widget")
		assert.Equal(t, "overwriting-icon", retrievedWidget2.Config.Icon, "Should have overwritten icon")
		assert.Equal(t, 3, *retrievedWidget2.Defaults.Width, "Should have overwritten width")
		assert.Equal(t, 3, *retrievedWidget2.Defaults.Height, "Should have overwritten height")
	})
}
