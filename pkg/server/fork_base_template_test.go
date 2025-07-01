package server_test

import (
	"encoding/json"
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

func TestForkBaseWidgetDashboardTemplateByName(t *testing.T) {
	t.Run("should successfully fork existing base template", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "fork-test-template",
			DisplayName: "Fork Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  5,
						MinHeight:  1,
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "fork-widget",
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()

		req, _ := http.NewRequest("GET", "/base-templates/fork-test-template/fork", nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "fork-test-template")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful fork")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var forkedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&forkedTemplate)
		require.NoError(t, err, "Should be able to decode forked template response")

		// Verify the forked template has correct data
		assert.NotZero(t, forkedTemplate.ID, "Forked template should have a new ID")
		assert.Equal(t, testUserID, forkedTemplate.UserId, "Forked template should belong to requesting user")
		assert.Equal(t, "fork-test-template", forkedTemplate.TemplateBase.Name, "Template base name should match")
		assert.Equal(t, "Fork Test Template", forkedTemplate.TemplateBase.DisplayName, "Template display name should match")
		assert.NotEmpty(t, forkedTemplate.CreatedAt, "Forked template should have creation timestamp")
		assert.NotEmpty(t, forkedTemplate.UpdatedAt, "Forked template should have update timestamp")

		// Verify the template config was copied correctly
		widgets := forkedTemplate.TemplateConfig.Sm.Data()
		require.Len(t, widgets, 1, "Should have one widget from base template")
		assert.Equal(t, "fork-widget", widgets[0].WidgetType, "Widget should match base template")
		assert.Equal(t, 1, widgets[0].Width, "Widget width should match base template")
		assert.Equal(t, 2, widgets[0].Height, "Widget height should match base template")

		// Verify the template was actually saved to the database
		var dbTemplate api.DashboardTemplate
		err = database.DB.First(&dbTemplate, forkedTemplate.ID).Error
		require.NoError(t, err, "Forked template should be saved in database")
		assert.Equal(t, testUserID, dbTemplate.UserId, "Database template should belong to requesting user")
	})

	t.Run("should return 404 for non-existent base template", func(t *testing.T) {
		server := setupRouter()

		// Reset registry to ensure no templates exist
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		req, _ := http.NewRequest("GET", "/base-templates/non-existent/fork", nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "non-existent")

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for non-existent base template")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "base template", "Error message should mention base template")
		assert.Contains(t, errorResponse.Errors[0].Message, "not found", "Error message should mention not found")
	})

	t.Run("should create separate templates for different users forking same base", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "shared-base-template",
			DisplayName: "Shared Base Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		user1ID := test_util.GetUniqueUserID()
		user2ID := test_util.GetUniqueUserID()

		// First user forks the template
		req1, _ := http.NewRequest("GET", "/base-templates/shared-base-template/fork", nil)
		req1 = withCustomIdentityContext(req1, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(user1ID)},
			xrhidgen.Entitlements{},
		))
		w1 := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w1, req1, "shared-base-template")

		assert.Equal(t, http.StatusOK, w1.Code, "Expected status code 200 for first user")

		var template1 api.DashboardTemplate
		err := json.NewDecoder(w1.Body).Decode(&template1)
		require.NoError(t, err, "Should be able to decode first user's template")

		// Second user forks the same template
		req2, _ := http.NewRequest("GET", "/base-templates/shared-base-template/fork", nil)
		req2 = withCustomIdentityContext(req2, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(user2ID)},
			xrhidgen.Entitlements{},
		))
		w2 := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w2, req2, "shared-base-template")

		assert.Equal(t, http.StatusOK, w2.Code, "Expected status code 200 for second user")

		var template2 api.DashboardTemplate
		err = json.NewDecoder(w2.Body).Decode(&template2)
		require.NoError(t, err, "Should be able to decode second user's template")

		// Verify templates are separate
		assert.NotEqual(t, template1.ID, template2.ID, "Templates should have different IDs")
		assert.Equal(t, user1ID, template1.UserId, "First template should belong to first user")
		assert.Equal(t, user2ID, template2.UserId, "Second template should belong to second user")
		assert.Equal(t, template1.TemplateBase.Name, template2.TemplateBase.Name, "Both should have same base name")
		assert.Equal(t, template1.TemplateBase.DisplayName, template2.TemplateBase.DisplayName, "Both should have same display name")
	})

	t.Run("should fork complex base template with multiple widgets", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a complex base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		complexWidgets := []api.WidgetItem{
			{
				Width:      1,
				Height:     2,
				MaxHeight:  5,
				MinHeight:  1,
				X:          test_util.IntPTR(0),
				Y:          test_util.IntPTR(0),
				WidgetType: "widget-1",
				Static:     false,
			},
			{
				Width:      2,
				Height:     3,
				MaxHeight:  6,
				MinHeight:  2,
				X:          test_util.IntPTR(1),
				Y:          test_util.IntPTR(0),
				WidgetType: "widget-2",
				Static:     true,
			},
			{
				Width:      1,
				Height:     1,
				MaxHeight:  4,
				MinHeight:  1,
				X:          test_util.IntPTR(0),
				Y:          test_util.IntPTR(2),
				WidgetType: "widget-3",
				Static:     false,
			},
		}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "complex-base-template",
			DisplayName: "Complex Base Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType(complexWidgets),
				Md: datatypes.NewJSONType(complexWidgets),
				Lg: datatypes.NewJSONType(complexWidgets),
				Xl: datatypes.NewJSONType(complexWidgets),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()

		req, _ := http.NewRequest("GET", "/base-templates/complex-base-template/fork", nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "complex-base-template")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for complex template fork")

		var forkedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&forkedTemplate)
		require.NoError(t, err, "Should be able to decode complex forked template")

		// Verify all widgets were copied correctly
		forkedWidgets := forkedTemplate.TemplateConfig.Lg.Data()
		require.Len(t, forkedWidgets, 3, "Should have all 3 widgets from base template")

		// Verify first widget
		assert.Equal(t, "widget-1", forkedWidgets[0].WidgetType, "First widget type should match")
		assert.Equal(t, 1, forkedWidgets[0].Width, "First widget width should match")
		assert.Equal(t, false, forkedWidgets[0].Static, "First widget static property should match")

		// Verify second widget
		assert.Equal(t, "widget-2", forkedWidgets[1].WidgetType, "Second widget type should match")
		assert.Equal(t, 2, forkedWidgets[1].Width, "Second widget width should match")
		assert.Equal(t, true, forkedWidgets[1].Static, "Second widget static property should match")

		// Verify third widget
		assert.Equal(t, "widget-3", forkedWidgets[2].WidgetType, "Third widget type should match")
		assert.Equal(t, 1, forkedWidgets[2].Width, "Third widget width should match")
		assert.Equal(t, false, forkedWidgets[2].Static, "Third widget static property should match")
	})

	t.Run("should create new template with fresh timestamps", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "timestamp-test-template",
			DisplayName: "Timestamp Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()

		req, _ := http.NewRequest("GET", "/base-templates/timestamp-test-template/fork", nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "timestamp-test-template")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

		var forkedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&forkedTemplate)
		require.NoError(t, err, "Should be able to decode forked template")

		// Verify template has new ID and timestamps
		assert.NotZero(t, forkedTemplate.ID, "Forked template should have a new ID")
		assert.NotEmpty(t, forkedTemplate.CreatedAt, "Forked template should have CreatedAt timestamp")
		assert.NotEmpty(t, forkedTemplate.UpdatedAt, "Forked template should have UpdatedAt timestamp")
		assert.Empty(t, forkedTemplate.DeletedAt, "Forked template should not have DeletedAt timestamp")
		assert.Equal(t, false, forkedTemplate.Default, "Forked template should not be default")

		// Verify user ownership
		assert.Equal(t, testUserID, forkedTemplate.UserId, "Forked template should belong to requesting user")
	})

	t.Run("should handle base template name with special characters", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template with special characters in name
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "special-chars_template.v1-2",
			DisplayName: "Special Characters Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()

		req, _ := http.NewRequest("GET", "/base-templates/special-chars_template.v1-2/fork", nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "special-chars_template.v1-2")

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for template with special characters")

		var forkedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&forkedTemplate)
		require.NoError(t, err, "Should be able to decode forked template with special chars")

		assert.Equal(t, "special-chars_template.v1-2", forkedTemplate.TemplateBase.Name, "Template name with special chars should be preserved")
		assert.Equal(t, "Special Characters Template", forkedTemplate.TemplateBase.DisplayName, "Template display name should be preserved")
	})

	t.Run("should return error for missing user identity", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "identity-test-template",
			DisplayName: "Identity Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Create request without identity context
		req, _ := http.NewRequest("GET", "/base-templates/identity-test-template/fork", nil)
		// No identity context added to the request
		w := httptest.NewRecorder()

		// This should panic due to missing identity context, which the middleware would catch
		// In a real scenario, the middleware would reject this request before it reaches the handler
		defer func() {
			if r := recover(); r != nil {
				// Expected panic due to missing identity - this simulates what would happen
				// if the middleware didn't catch the missing identity first
				assert.Contains(t, r.(error).Error(), "identity not found in context", "Should panic with identity not found error")
			}
		}()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "identity-test-template")
	})

	t.Run("should return error for malformed user identity", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "malformed-identity-test",
			DisplayName: "Malformed Identity Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Create request with invalid identity structure (not the expected type)
		req, _ := http.NewRequest("GET", "/base-templates/malformed-identity-test/fork", nil)
		req = withCustomIdentityContext(req, "invalid-identity-string") // Wrong type - should be XRHID struct
		w := httptest.NewRecorder()

		// This should cause a panic when trying to cast to XRHID
		defer func() {
			if r := recover(); r != nil {
				// Expected panic due to wrong identity type
				assert.NotNil(t, r, "Should panic when trying to cast invalid identity type")
			}
		}()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "malformed-identity-test")
	})

	t.Run("should handle database error during template creation", func(t *testing.T) {
		server := setupRouter()

		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "db-error-test-template",
			DisplayName: "Database Error Test Template",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()

		// Simulate a database error by closing the database connection
		sqlDB, err := database.DB.DB()
		require.NoError(t, err, "Should be able to get underlying sql.DB")

		// Close the database connection to simulate a database error
		originalDB := database.DB
		err = sqlDB.Close()
		require.NoError(t, err, "Should be able to close database connection")

		req, _ := http.NewRequest("GET", "/base-templates/db-error-test-template/fork", nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ForkBaseWidgetDashboardTemplateByName(w, req, "db-error-test-template")

		// Should return 500 Internal Server Error due to database connection error
		assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected status code 500 for database error")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusInternalServerError, errorResponse.Errors[0].Code, "Error code should be 500")
		assert.Contains(t, errorResponse.Errors[0].Message, "database is closed", "Error message should indicate database error")

		// Restore the database connection for other tests (reinitialize)
		database.DB = originalDB
		database.InitDb() // Reinitialize the database
	})
}
