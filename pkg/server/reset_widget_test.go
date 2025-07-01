package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
)

func TestResetWidgetLayoutById(t *testing.T) {
	t.Run("should successfully reset dashboard template to base template", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template for testing
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseWidget := api.WidgetItem{
			Width:      2,
			Height:     3,
			MaxHeight:  8,
			MinHeight:  1,
			X:          intPtr(0),
			Y:          intPtr(0),
			WidgetType: "base-widget",
		}
		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "test-base-template",
			DisplayName: "Test Base Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{baseWidget}),
				Md: datatypes.NewJSONType([]api.WidgetItem{baseWidget}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{baseWidget}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{baseWidget}),
			},
		}
		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Create a dashboard template in the database that has been modified
		testUserID := test_util.GetUniqueUserID()
		templateID := test_util.GetUniqueID()

		modifiedWidget := api.WidgetItem{
			Width:      4,
			Height:     2,
			MaxHeight:  6,
			MinHeight:  1,
			X:          intPtr(1),
			Y:          intPtr(1),
			WidgetType: "modified-widget",
		}
		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: testUserID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test-base-template", // Must match base template name
				DisplayName: "Test Dashboard Template",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{modifiedWidget}),
				Md: datatypes.NewJSONType([]api.WidgetItem{modifiedWidget}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{modifiedWidget}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{modifiedWidget}),
			},
		}

		result := database.DB.Create(&testTemplate)
		require.NoError(t, result.Error, "Should be able to create test template in DB")

		// Reset the template
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful reset")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var resetTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&resetTemplate)
		require.NoError(t, err, "Should be able to decode reset template response")

		// Verify the template was reset to base template config
		assert.Equal(t, templateID, resetTemplate.ID, "Template ID should remain the same")
		assert.Equal(t, testUserID, resetTemplate.UserId, "User ID should remain the same")
		assert.Equal(t, "test-base-template", resetTemplate.TemplateBase.Name, "Template base name should remain the same")

		// Verify the config was reset to base template
		resetWidgets := resetTemplate.TemplateConfig.Lg.Data()
		require.Len(t, resetWidgets, 1, "Should have one widget from base template")
		assert.Equal(t, "base-widget", resetWidgets[0].WidgetType, "Widget should be reset to base template widget")
		assert.Equal(t, 2, resetWidgets[0].Width, "Widget width should match base template")
		assert.Equal(t, 3, resetWidgets[0].Height, "Widget height should match base template")
	})

	t.Run("should return 404 for non-existent template ID", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", test_util.NonExistentID), nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(test_util.NonExistentID))

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for non-existent template")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "record not found", "Error message should mention record not found")
	})

	t.Run("should return 403 for unauthorized access", func(t *testing.T) {
		server := setupRouter()

		// Create a template belonging to a different user
		templateID := test_util.GetUniqueID()
		templateOwnerID := test_util.GetUniqueUserID()
		requestingUserID := test_util.GetUniqueUserID()

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: templateOwnerID, // Different from requesting user
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test-template",
				DisplayName: "Test Template",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		result := database.DB.Create(&testTemplate)
		require.NoError(t, result.Error, "Should be able to create other user's template in DB")

		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status code 403 for unauthorized access")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusForbidden, errorResponse.Errors[0].Code, "Error code should be 403")
		assert.Contains(t, errorResponse.Errors[0].Message, "unauthorized", "Error message should mention unauthorized")
	})

	t.Run("should return 404 when base template does not exist", func(t *testing.T) {
		server := setupRouter()

		// Reset registry to ensure no base templates exist
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		// Create a dashboard template that references a non-existent base template
		testUserID := test_util.GetUniqueUserID()
		templateID := test_util.GetUniqueID()

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: testUserID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "non-existent-base-template", // This base template doesn't exist
				DisplayName: "Test Dashboard Template",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		result := database.DB.Create(&testTemplate)
		require.NoError(t, result.Error, "Should be able to create test template in DB")

		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 when base template doesn't exist")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "base template", "Error message should mention base template")
		assert.Contains(t, errorResponse.Errors[0].Message, "not found", "Error message should mention not found")
	})

	t.Run("should reset template with complex widget configuration", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a complex base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseWidgets := []api.WidgetItem{
			{
				Width:      1,
				Height:     2,
				MaxHeight:  5,
				MinHeight:  1,
				X:          intPtr(0),
				Y:          intPtr(0),
				WidgetType: "widget-1",
			},
			{
				Width:      2,
				Height:     3,
				MaxHeight:  6,
				MinHeight:  2,
				X:          intPtr(1),
				Y:          intPtr(0),
				WidgetType: "widget-2",
			},
		}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "complex-base-template",
			DisplayName: "Complex Base Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType(baseWidgets),
				Md: datatypes.NewJSONType(baseWidgets),
				Lg: datatypes.NewJSONType(baseWidgets),
				Xl: datatypes.NewJSONType(baseWidgets),
			},
		}
		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Create a dashboard template with different configuration
		testUserID := test_util.GetUniqueUserID()
		templateID := test_util.GetUniqueID()

		modifiedWidgets := []api.WidgetItem{
			{
				Width:      4,
				Height:     1,
				MaxHeight:  3,
				MinHeight:  1,
				X:          intPtr(2),
				Y:          intPtr(1),
				WidgetType: "different-widget",
			},
		}

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: testUserID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "complex-base-template",
				DisplayName: "Complex Dashboard Template",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType(modifiedWidgets),
				Md: datatypes.NewJSONType(modifiedWidgets),
				Lg: datatypes.NewJSONType(modifiedWidgets),
				Xl: datatypes.NewJSONType(modifiedWidgets),
			},
		}

		result := database.DB.Create(&testTemplate)
		require.NoError(t, result.Error, "Should be able to create test template in DB")

		// Reset the template
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful reset")

		var resetTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&resetTemplate)
		require.NoError(t, err, "Should be able to decode reset template response")

		// Verify the template was reset to base template config
		resetWidgets := resetTemplate.TemplateConfig.Lg.Data()
		require.Len(t, resetWidgets, 2, "Should have two widgets from base template")

		// Verify first widget
		assert.Equal(t, "widget-1", resetWidgets[0].WidgetType, "First widget should match base template")
		assert.Equal(t, 1, resetWidgets[0].Width, "First widget width should match base template")
		assert.Equal(t, 2, resetWidgets[0].Height, "First widget height should match base template")

		// Verify second widget
		assert.Equal(t, "widget-2", resetWidgets[1].WidgetType, "Second widget should match base template")
		assert.Equal(t, 2, resetWidgets[1].Width, "Second widget width should match base template")
		assert.Equal(t, 3, resetWidgets[1].Height, "Second widget height should match base template")
	})

	t.Run("should preserve template metadata when resetting", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "metadata-test-base",
			DisplayName: "Metadata Test Base",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}
		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Create a dashboard template with specific metadata
		testUserID := test_util.GetUniqueUserID()
		templateID := test_util.GetUniqueID()

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: testUserID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "metadata-test-base",
				DisplayName: "Custom Display Name", // Different from base template
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
			Default: true, // Custom default setting
		}

		result := database.DB.Create(&testTemplate)
		require.NoError(t, result.Error, "Should be able to create test template in DB")

		// Reset the template
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/reset", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ResetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful reset")

		var resetTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&resetTemplate)
		require.NoError(t, err, "Should be able to decode reset template response")

		// Verify metadata is preserved
		assert.Equal(t, templateID, resetTemplate.ID, "Template ID should be preserved")
		assert.Equal(t, testUserID, resetTemplate.UserId, "User ID should be preserved")
		assert.Equal(t, "metadata-test-base", resetTemplate.TemplateBase.Name, "Template base name should be preserved")
		assert.Equal(t, "Custom Display Name", resetTemplate.TemplateBase.DisplayName, "Custom display name should be preserved")
		assert.Equal(t, true, resetTemplate.Default, "Default setting should be preserved")
		assert.NotEmpty(t, resetTemplate.CreatedAt, "CreatedAt should be preserved")
		assert.NotEmpty(t, resetTemplate.UpdatedAt, "UpdatedAt should be updated")
	})
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
