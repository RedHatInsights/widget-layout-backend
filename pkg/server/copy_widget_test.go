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

func TestCopyWidgetLayoutById(t *testing.T) {
	t.Run("should successfully copy existing dashboard template", func(t *testing.T) {
		server := setupRouter()

		// Create an original dashboard template in the database
		originalUserID := test_util.GetUniqueUserID()
		copyingUserID := test_util.GetUniqueUserID()

		originalDashboard := test_util.MockDashboardTemplate()
		originalDashboard.UserId = originalUserID // Different user - to verify copying works across users
		originalDashboard.TemplateBase.Name = "Original Template"
		originalDashboard.TemplateBase.DisplayName = "Original Display Name"
		result := database.DB.Create(&originalDashboard)
		assert.NoError(t, result.Error, "Should be able to create original dashboard in DB")

		templateID := int64(originalDashboard.ID)

		// Perform the COPY request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/copy", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(copyingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.CopyWidgetLayoutById(w, req, templateID)

		// Check the response
		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 201 for successful copy")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var copiedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&copiedTemplate)
		assert.NoError(t, err, "Should be able to decode copied template response")

		// Verify the copied template has different ID but same base content, and correct user ID
		assert.NotEqual(t, originalDashboard.ID, copiedTemplate.ID, "Copied template should have different ID")
		assert.Equal(t, copyingUserID, copiedTemplate.UserId, "Copied template should belong to copying user")
		assert.Equal(t, originalDashboard.TemplateBase.Name, copiedTemplate.TemplateBase.Name, "Copied template should have same name")
		assert.Equal(t, originalDashboard.TemplateBase.DisplayName, copiedTemplate.TemplateBase.DisplayName, "Copied template should have same display name")
		assert.Equal(t, originalDashboard.TemplateConfig, copiedTemplate.TemplateConfig, "Copied template should have same template config")

		// Verify both templates exist in the database
		var originalFromDB api.DashboardTemplate
		err = database.DB.First(&originalFromDB, originalDashboard.ID).Error
		assert.NoError(t, err, "Original template should still exist in database")

		var copiedFromDB api.DashboardTemplate
		err = database.DB.First(&copiedFromDB, copiedTemplate.ID).Error
		assert.NoError(t, err, "Copied template should exist in database")

		// Verify they are separate entities with correct ownership
		assert.NotEqual(t, originalFromDB.ID, copiedFromDB.ID, "Templates should have different IDs in database")
		assert.Equal(t, originalUserID, originalFromDB.UserId, "Original template should belong to original user")
		assert.Equal(t, copyingUserID, copiedFromDB.UserId, "Copied template should belong to copying user")
	})

	t.Run("should return 404 for non-existent template", func(t *testing.T) {
		server := setupRouter()

		nonExistentID := int64(test_util.NonExistentID)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/copy", nonExistentID), nil)
		req, _ = withUniqueUserIdentityContext(req)
		w := httptest.NewRecorder()

		server.CopyWidgetLayoutById(w, req, nonExistentID)

		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for non-existent template")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse, "Error response should contain error messages")
		assert.Contains(t, errorResponse.Errors[0].Message, "record not found", "Error should mention record not found")
	})

	t.Run("should copy template with correct widget configuration", func(t *testing.T) {
		server := setupRouter()

		// Create a template with specific widget configuration
		testWidget := api.WidgetItem{
			Height:     3,
			Width:      4,
			X:          test_util.IntPTR(1),
			WidgetType: "special-widget",
			Y:          test_util.IntPTR(2),
			Static:     true,
			MaxHeight:  6,
			MinHeight:  2,
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})

		templateID := test_util.GetUniqueID()
		originalUserID := test_util.GetUniqueUserID()

		originalDashboard := api.DashboardTemplate{
			ID: templateID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "Widget Config Test",
				DisplayName: "Widget Configuration Test Template",
			},
			UserId: originalUserID,
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: tm,
				Md: tm,
				Sm: tm,
				Xl: tm,
			},
		}

		result := database.DB.Create(&originalDashboard)
		assert.NoError(t, result.Error, "Should be able to create original dashboard with widget config")

		templateIDInt64 := int64(templateID)

		// Copy the template
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/copy", templateIDInt64), nil)
		req, _ = withUniqueUserIdentityContext(req)
		w := httptest.NewRecorder()

		server.CopyWidgetLayoutById(w, req, templateIDInt64)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 201 for successful copy")

		var copiedTemplate api.DashboardTemplate
		err := json.NewDecoder(w.Body).Decode(&copiedTemplate)
		assert.NoError(t, err, "Should be able to decode copied template response")

		// Verify widget configuration is copied correctly
		originalWidgets := originalDashboard.TemplateConfig.Lg.Data()
		copiedWidgets := copiedTemplate.TemplateConfig.Lg.Data()

		assert.Equal(t, len(originalWidgets), len(copiedWidgets), "Copied template should have same number of widgets")
		if len(copiedWidgets) > 0 && len(originalWidgets) > 0 {
			assert.Equal(t, originalWidgets[0].Height, copiedWidgets[0].Height, "Widget height should match")
			assert.Equal(t, originalWidgets[0].Width, copiedWidgets[0].Width, "Widget width should match")
			assert.Equal(t, originalWidgets[0].X, copiedWidgets[0].X, "Widget X position should match")
			assert.Equal(t, originalWidgets[0].Y, copiedWidgets[0].Y, "Widget Y position should match")
			assert.Equal(t, originalWidgets[0].WidgetType, copiedWidgets[0].WidgetType, "Widget type should match")
			assert.Equal(t, originalWidgets[0].Static, copiedWidgets[0].Static, "Widget static property should match")
			assert.Equal(t, originalWidgets[0].MaxHeight, copiedWidgets[0].MaxHeight, "Widget max height should match")
			assert.Equal(t, originalWidgets[0].MinHeight, copiedWidgets[0].MinHeight, "Widget min height should match")
		}
	})
}
