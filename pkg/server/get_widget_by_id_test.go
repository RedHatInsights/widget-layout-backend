package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
)

func TestGetWidgetLayoutById(t *testing.T) {
	t.Run("should return specific widget by ID for authorized user", func(t *testing.T) {
		server := setupRouter()

		// Create a dashboard template in the database with matching user ID
		templateID := test_util.GetUniqueID()
		testUserID := test_util.GetUniqueUserID()

		testWidget := api.WidgetItem{
			Height:     2,
			Width:      2,
			X:          0,
			WidgetType: "widget1",
			Y:          0,
			Static:     false,
			Title:      "Sample Widget",
			MaxHeight:  4,
			MinHeight:  1,
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})
		testTemplateConfig := api.DashboardTemplateConfig{
			Lg: tm,
			Md: tm,
			Sm: tm,
			Xl: tm,
		}
		testTemplate := api.DashboardTemplate{
			ID:             templateID,
			UserId:         testUserID,
			TemplateConfig: testTemplateConfig,
		}

		// Save to database
		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create test template in DB")

		// Test with the correct ID and authorized user
		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.GetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var parsedResp api.DashboardTemplate
		err := json.Unmarshal(w.Body.Bytes(), &parsedResp)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, testTemplate.ID, parsedResp.ID, "Expected widget ID to match")
		assert.Equal(t, testTemplate.UserId, parsedResp.UserId, "Expected user ID to match")
		assert.Equal(t, testTemplate.TemplateConfig, parsedResp.TemplateConfig, "Expected template config to match")
	})

	t.Run("should return 404 for non-existent widget ID", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", test_util.NonExistentID), nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.GetWidgetLayoutById(w, req, int64(test_util.NonExistentID))

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "record not found", "Error message should mention record not found")
	})

	t.Run("should return 404 for template belonging to different user", func(t *testing.T) {
		server := setupRouter()

		// Create a template belonging to a different user
		templateID := test_util.GetUniqueID()
		otherUserID := test_util.GetUniqueUserID()
		requestingUserID := test_util.GetUniqueUserID()

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: otherUserID, // Different from requesting user
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: datatypes.NewJSONType([]api.WidgetItem{{
					Height:     2,
					Width:      2,
					X:          0,
					WidgetType: "widget1",
					Y:          0,
					Static:     false,
					Title:      "Other User Widget",
					MaxHeight:  4,
					MinHeight:  1,
				}}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create other user's template in DB")

		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.GetWidgetLayoutById(w, req, int64(templateID))

		// Should return 404 because the template doesn't belong to the requesting user
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for unauthorized access")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
	})

	t.Run("should return valid JSON for existing authorized widget", func(t *testing.T) {
		server := setupRouter()

		// Create a template for the authorized user
		templateID := test_util.GetUniqueID()
		testUserID := test_util.GetUniqueUserID()

		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: testUserID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "Test Template",
				DisplayName: "Test Display Name",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: datatypes.NewJSONType([]api.WidgetItem{{
					Height:     3,
					Width:      4,
					X:          1,
					WidgetType: "test-widget",
					Y:          2,
					Static:     true,
					Title:      "Test Widget Title",
					MaxHeight:  6,
					MinHeight:  1,
				}}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create test template in DB")

		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.GetWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var parsedTemplate api.DashboardTemplate
		err := json.Unmarshal(w.Body.Bytes(), &parsedTemplate)
		assert.NoError(t, err, "Response should be valid JSON")

		// Verify all template properties
		assert.Equal(t, testTemplate.ID, parsedTemplate.ID, "Template ID should match")
		assert.Equal(t, testTemplate.UserId, parsedTemplate.UserId, "User ID should match")
		assert.Equal(t, testTemplate.TemplateBase.Name, parsedTemplate.TemplateBase.Name, "Template name should match")
		assert.Equal(t, testTemplate.TemplateBase.DisplayName, parsedTemplate.TemplateBase.DisplayName, "Template display name should match")

		// Verify widget configuration
		originalWidgets := testTemplate.TemplateConfig.Lg.Data()
		responseWidgets := parsedTemplate.TemplateConfig.Lg.Data()
		assert.Equal(t, len(originalWidgets), len(responseWidgets), "Widget count should match")
		if len(responseWidgets) > 0 {
			assert.Equal(t, originalWidgets[0].Height, responseWidgets[0].Height, "Widget height should match")
			assert.Equal(t, originalWidgets[0].Width, responseWidgets[0].Width, "Widget width should match")
			assert.Equal(t, originalWidgets[0].WidgetType, responseWidgets[0].WidgetType, "Widget type should match")
			assert.Equal(t, originalWidgets[0].Title, responseWidgets[0].Title, "Widget title should match")
		}
	})
}
