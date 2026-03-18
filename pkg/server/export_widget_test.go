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

func TestExportWidgetLayoutById(t *testing.T) {
	t.Run("should return JSON of specific widget by ID for authorized user", func(t *testing.T) {
		server := setupRouter()

		// Create a dashboard template in the database with matching user ID
		templateID := test_util.GetUniqueID()
		testUserID := test_util.GetUniqueUserID()

		testWidget := api.WidgetItem{
			Height:     2,
			Width:      2,
			X:          test_util.IntPTR(0),
			WidgetType: "widget1",
			Y:          test_util.IntPTR(0),
			Static:     false,
			MaxHeight:  test_util.IntPTR(4),
			MinHeight:  test_util.IntPTR(1),
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})
		testTemplateConfig := api.DashboardTemplateConfig{
			Lg: tm,
			Md: tm,
			Sm: tm,
			Xl: tm,
		}
		testTemplateBase := api.DashboardTemplateBase{
			Name:        "test-dashboard",
			DisplayName: "Test Dashboard",
		}
		testTemplate := api.DashboardTemplate{
			ID:             templateID,
			UserId:         testUserID,
			DashboardName:  "Test",
			TemplateConfig: testTemplateConfig,
			TemplateBase:   testTemplateBase,
		}

		// Save to database
		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create test template in DB")

		// Test with the correct ID and authorized user
		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d/export", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ExportWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var exportResp api.ExportWidgetDashboardTemplateResponse
		err := json.Unmarshal(w.Body.Bytes(), &exportResp)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, testTemplateBase, exportResp.TemplateBase, "Expected template base to match")
		assert.Equal(t, testTemplateConfig, exportResp.TemplateConfig, "Expected template config to match")
	})

	t.Run("should not have metadata fields (userId, ID, createdAt, updatedAt, deletedAt)", func(t *testing.T) {
		server := setupRouter()

		// Create a dashboard template in the database with matching user ID
		templateID := test_util.GetUniqueID()
		testUserID := test_util.GetUniqueUserID()

		testWidget := api.WidgetItem{
			Height:     2,
			Width:      2,
			X:          test_util.IntPTR(0),
			WidgetType: "widget1",
			Y:          test_util.IntPTR(0),
			Static:     false,
			MaxHeight:  test_util.IntPTR(4),
			MinHeight:  test_util.IntPTR(1),
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})
		testTemplateConfig := api.DashboardTemplateConfig{
			Lg: tm,
			Md: tm,
			Sm: tm,
			Xl: tm,
		}
		testTemplateBase := api.DashboardTemplateBase{
			Name:        "test-dashboard",
			DisplayName: "Test Dashboard",
		}
		testTemplate := api.DashboardTemplate{
			ID:             templateID,
			UserId:         testUserID,
			TemplateConfig: testTemplateConfig,
			TemplateBase:   testTemplateBase,
		}

		// Save to database
		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create test template in DB")

		// Test with the correct ID and authorized user
		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d/export", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ExportWidgetLayoutById(w, req, int64(templateID))

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

		// Parse as raw JSON
		var rawResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &rawResp)
		assert.NoError(t, err, "Should be able to parse as raw JSON")

		assert.NotContains(t, rawResp, "userId", "Export should not contain userId")
		assert.NotContains(t, rawResp, "id", "Export should not contain id")
		assert.NotContains(t, rawResp, "ID", "Export should not contain ID")
		assert.NotContains(t, rawResp, "createdAt", "Export should not contain createdAt")
		assert.NotContains(t, rawResp, "updatedAt", "Export should not contain updatedAt")
		assert.NotContains(t, rawResp, "deletedAt", "Export should not contain deletedAt")
	})

	t.Run("should return 404 for non-existent widget ID", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d/export", test_util.NonExistentID), nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.ExportWidgetLayoutById(w, req, int64(test_util.NonExistentID))

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
		assert.Contains(t, errorResponse.Errors[0].Message, "record not found", "Error message should mention record not found")
	})

	t.Run("should return 404 for exporting template belonging to different user", func(t *testing.T) {
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
					X:          test_util.IntPTR(0),
					WidgetType: "widget1",
					Y:          test_util.IntPTR(0),
					Static:     false,
					MaxHeight:  test_util.IntPTR(4),
					MinHeight:  test_util.IntPTR(1),
				}}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create other user's template in DB")

		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d/export", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ExportWidgetLayoutById(w, req, int64(templateID))

		// Should return 404 because the template doesn't belong to the requesting user
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for unauthorized access")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Equal(t, http.StatusNotFound, errorResponse.Errors[0].Code, "Error code should be 404")
	})
}
