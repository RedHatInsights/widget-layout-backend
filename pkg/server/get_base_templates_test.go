package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestGetBaseWidgetDashboardTemplates(t *testing.T) {
	t.Run("should return empty array when no base templates exist", func(t *testing.T) {
		server := setupRouter()

		// Reset registry to ensure no templates exist
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		req, _ := http.NewRequest("GET", "/base-templates", nil)
		w := httptest.NewRecorder()

		server.GetBaseWidgetDashboardTemplates(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var templates []api.BaseWidgetDashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&templates)
		require.NoError(t, err, "Should be able to decode response as array")
		assert.Empty(t, templates, "Should return empty array when no templates exist")
	})

	t.Run("should return array of base templates", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add test templates
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate1 := api.BaseWidgetDashboardTemplate{
			Name:        "template-1",
			DisplayName: "Template 1",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		baseTemplate2 := api.BaseWidgetDashboardTemplate{
			Name:        "template-2",
			DisplayName: "Template 2",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate1)
		service.BaseTemplateRegistry.AddBase(baseTemplate2)

		req, _ := http.NewRequest("GET", "/base-templates", nil)
		w := httptest.NewRecorder()

		server.GetBaseWidgetDashboardTemplates(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var templates []api.BaseWidgetDashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&templates)
		require.NoError(t, err, "Should be able to decode response as array")
		assert.Len(t, templates, 2, "Should return array with 2 templates")

		// Verify both templates are present (order doesn't matter in map iteration)
		templateNames := make(map[string]string)
		for _, template := range templates {
			templateNames[template.Name] = template.DisplayName
		}

		assert.Equal(t, "Template 1", templateNames["template-1"], "Template 1 should be present")
		assert.Equal(t, "Template 2", templateNames["template-2"], "Template 2 should be present")
	})
}

func TestGetBaseWidgetDashboardTemplateByName(t *testing.T) {
	t.Run("should return specific base template by name", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a test template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "test-template",
			DisplayName: "Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  5,
						MinHeight:  1,
						X:          intPtr(0),
						Y:          intPtr(0),
						WidgetType: "test-widget",
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		req, _ := http.NewRequest("GET", "/base-templates/test-template", nil)
		w := httptest.NewRecorder()

		server.GetBaseWidgetDashboardTemplateByName(w, req, "test-template")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var template api.BaseWidgetDashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&template)
		require.NoError(t, err, "Should be able to decode template response")

		assert.Equal(t, "test-template", template.Name, "Template name should match")
		assert.Equal(t, "Test Template", template.DisplayName, "Template display name should match")

		// Verify template config
		widgets := template.TemplateConfig.Sm.Data()
		require.Len(t, widgets, 1, "Should have one widget in sm config")
		assert.Equal(t, "test-widget", widgets[0].WidgetType, "Widget type should match")
	})

	t.Run("should return 404 for non-existent base template", func(t *testing.T) {
		server := setupRouter()

		// Reset registry to ensure no templates exist
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		req, _ := http.NewRequest("GET", "/base-templates/non-existent", nil)
		w := httptest.NewRecorder()

		server.GetBaseWidgetDashboardTemplateByName(w, req, "non-existent")

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for non-existent template")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "Base template not found", "Error message should mention base template not found")
	})

	t.Run("should return correct template among multiple templates", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add multiple templates
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		template1 := api.BaseWidgetDashboardTemplate{
			Name:        "template-1",
			DisplayName: "Template 1",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		template2 := api.BaseWidgetDashboardTemplate{
			Name:        "template-2",
			DisplayName: "Template 2",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(template1)
		service.BaseTemplateRegistry.AddBase(template2)

		req, _ := http.NewRequest("GET", "/base-templates/template-2", nil)
		w := httptest.NewRecorder()

		server.GetBaseWidgetDashboardTemplateByName(w, req, "template-2")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

		var template api.BaseWidgetDashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&template)
		require.NoError(t, err, "Should be able to decode template response")

		assert.Equal(t, "template-2", template.Name, "Should return the correct template")
		assert.Equal(t, "Template 2", template.DisplayName, "Should return the correct template display name")
	})
}
