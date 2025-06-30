package service_test

import (
	"os"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

// Test data constants
const validBaseTemplateJSON = `[
	{
		"name": "test-template-1",
		"displayName": "Test Template 1",
		"templateConfig": {
			"sm": [
				{
					"w": 1,
					"h": 4,
					"maxH": 10,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-1"
				}
			],
			"md": [
				{
					"w": 2,
					"h": 4,
					"maxH": 10,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-1"
				}
			],
			"lg": [
				{
					"w": 3,
					"h": 4,
					"maxH": 10,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-1"
				}
			],
			"xl": [
				{
					"w": 4,
					"h": 4,
					"maxH": 10,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-1"
				}
			]
		},
		"frontendRef": "test-frontend-1"
	},
	{
		"name": "test-template-2",
		"displayName": "Test Template 2",
		"templateConfig": {
			"sm": [
				{
					"w": 1,
					"h": 3,
					"maxH": 8,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-2"
				}
			],
			"md": [
				{
					"w": 2,
					"h": 3,
					"maxH": 8,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-2"
				}
			],
			"lg": [
				{
					"w": 3,
					"h": 3,
					"maxH": 8,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-2"
				}
			],
			"xl": [
				{
					"w": 4,
					"h": 3,
					"maxH": 8,
					"minH": 1,
					"x": 0,
					"y": 0,
					"i": "widget-2"
				}
			]
		},
		"frontendRef": "test-frontend-2"
	}
]`

const invalidJSON = `[
	{
		"name": "test-template",
		"displayName": "Test Template",
		"templateConfig": {
			"sm": [
				{
					"w": 1,
					"h": 4
					// Missing comma - invalid JSON
				}
			]
		}
	}
]`

func TestLoadBaseTemplatesFromConfig(t *testing.T) {
	t.Run("should load valid base templates from JSON config", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		err := service.LoadBaseTemplatesFromConfig(validBaseTemplateJSON)
		require.NoError(t, err)

		// Verify templates were loaded
		assert.Len(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates, 2)

		// Verify first template
		template1, exists := service.BaseTemplateRegistry.GetBase("test-template-1")
		assert.True(t, exists)
		assert.Equal(t, "test-template-1", template1.Name)
		assert.Equal(t, "Test Template 1", template1.DisplayName)
		assert.Equal(t, "test-frontend-1", template1.FrontendRef)

		// Verify template config structure
		assert.NotNil(t, template1.TemplateConfig.Sm)
		assert.NotNil(t, template1.TemplateConfig.Md)
		assert.NotNil(t, template1.TemplateConfig.Lg)
		assert.NotNil(t, template1.TemplateConfig.Xl)

		// Verify second template
		template2, exists := service.BaseTemplateRegistry.GetBase("test-template-2")
		assert.True(t, exists)
		assert.Equal(t, "test-template-2", template2.Name)
		assert.Equal(t, "Test Template 2", template2.DisplayName)
		assert.Equal(t, "test-frontend-2", template2.FrontendRef)
	})

	t.Run("should handle empty config string", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		err := service.LoadBaseTemplatesFromConfig("")
		require.NoError(t, err)

		// Registry should remain empty
		assert.Empty(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates)
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		err := service.LoadBaseTemplatesFromConfig(invalidJSON)
		assert.Error(t, err)

		// Registry should remain empty on error
		assert.Empty(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates)
	})

	t.Run("should handle malformed JSON structure", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		malformedJSON := `{"not": "an array"}`
		err := service.LoadBaseTemplatesFromConfig(malformedJSON)
		assert.Error(t, err)

		// Registry should remain empty on error
		assert.Empty(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates)
	})

	t.Run("should handle completely invalid JSON", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		invalidJSONString := `not json at all`
		err := service.LoadBaseTemplatesFromConfig(invalidJSONString)
		assert.Error(t, err)

		// Registry should remain empty on error
		assert.Empty(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates)
	})

	t.Run("should handle partial template data", func(t *testing.T) {
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		partialTemplateJSON := `[
			{
				"name": "partial-template",
				"displayName": "Partial Template",
				"templateConfig": {
					"sm": [],
					"md": [],
					"lg": [],
					"xl": []
				}
			}
		]`

		err := service.LoadBaseTemplatesFromConfig(partialTemplateJSON)
		require.NoError(t, err)

		// Verify template was loaded
		assert.Len(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates, 1)

		template, exists := service.BaseTemplateRegistry.GetBase("partial-template")
		assert.True(t, exists)
		assert.Equal(t, "partial-template", template.Name)
		assert.Equal(t, "Partial Template", template.DisplayName)
		assert.Empty(t, template.FrontendRef) // Should be empty when not provided
	})
}

func TestBaseWidgetDashboardTemplateRegistry(t *testing.T) {
	t.Run("should add and retrieve base templates", func(t *testing.T) {
		registry := models.BaseWidgetDashboardTemplateRegistry{}

		template := models.BaseWidgetDashboardTemplate{
			Name:        "test-template",
			DisplayName: "Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
			FrontendRef: "test-frontend",
		}

		registry.AddBase(template)

		retrievedTemplate, exists := registry.GetBase("test-template")
		assert.True(t, exists)
		assert.Equal(t, "test-template", retrievedTemplate.Name)
		assert.Equal(t, "Test Template", retrievedTemplate.DisplayName)
		assert.Equal(t, "test-frontend", retrievedTemplate.FrontendRef)
	})

	t.Run("should handle non-existent template", func(t *testing.T) {
		registry := models.BaseWidgetDashboardTemplateRegistry{}

		template, exists := registry.GetBase("non-existent-template")
		assert.False(t, exists)
		assert.Empty(t, template.Name)
	})

	t.Run("should overwrite existing template with same name", func(t *testing.T) {
		registry := models.BaseWidgetDashboardTemplateRegistry{}

		// Add first template
		template1 := models.BaseWidgetDashboardTemplate{
			Name:        "same-name",
			DisplayName: "First Template",
			FrontendRef: "first-frontend",
		}
		registry.AddBase(template1)

		// Add second template with same name
		template2 := models.BaseWidgetDashboardTemplate{
			Name:        "same-name",
			DisplayName: "Second Template",
			FrontendRef: "second-frontend",
		}
		registry.AddBase(template2)

		// Should have only one template (the second one)
		assert.Len(t, registry.BaseWidgetDashboardTemplates, 1)

		retrievedTemplate, exists := registry.GetBase("same-name")
		assert.True(t, exists)
		assert.Equal(t, "Second Template", retrievedTemplate.DisplayName)
		assert.Equal(t, "second-frontend", retrievedTemplate.FrontendRef)
	})
}

func TestBaseWidgetDashboardTemplateToDashboardTemplate(t *testing.T) {
	t.Run("should convert BaseWidgetDashboardTemplate to DashboardTemplate", func(t *testing.T) {
		baseTemplate := models.BaseWidgetDashboardTemplate{
			Name:        "test-template",
			DisplayName: "Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     4,
						MaxHeight:  10,
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
			FrontendRef: "test-frontend",
		}

		dashboardTemplate := baseTemplate.ToDashboardTemplate()

		assert.Equal(t, "test-template", dashboardTemplate.TemplateBase.Name)
		assert.Equal(t, "Test Template", dashboardTemplate.TemplateBase.DisplayName)
		assert.Equal(t, baseTemplate.TemplateConfig, dashboardTemplate.TemplateConfig)
	})
}

func TestConfigurationIntegration(t *testing.T) {
	t.Run("should work with environment variable format", func(t *testing.T) {
		// This test simulates how the configuration would be used in practice
		// Reset registry for clean test
		service.BaseTemplateRegistry = models.BaseWidgetDashboardTemplateRegistry{}

		// Set environment variable temporarily
		originalEnv := os.Getenv("BASE_LAYOUTS")
		defer func() {
			if originalEnv != "" {
				os.Setenv("BASE_LAYOUTS", originalEnv)
			} else {
				os.Unsetenv("BASE_LAYOUTS")
			}
		}()

		os.Setenv("BASE_LAYOUTS", validBaseTemplateJSON)

		// Load from config (simulating what happens in init)
		err := service.LoadBaseTemplatesFromConfig(os.Getenv("BASE_LAYOUTS"))
		require.NoError(t, err)

		// Verify templates are loaded
		assert.Len(t, service.BaseTemplateRegistry.BaseWidgetDashboardTemplates, 2)

		// Verify we can retrieve templates
		template1, exists := service.BaseTemplateRegistry.GetBase("test-template-1")
		assert.True(t, exists)
		assert.Equal(t, "Test Template 1", template1.DisplayName)

		template2, exists := service.BaseTemplateRegistry.GetBase("test-template-2")
		assert.True(t, exists)
		assert.Equal(t, "Test Template 2", template2.DisplayName)
	})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
