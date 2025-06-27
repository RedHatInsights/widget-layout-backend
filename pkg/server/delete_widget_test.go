package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/subpop/xrhidgen"
	"gorm.io/gorm"
)

func TestDeleteWidgetLayoutById(t *testing.T) {
	t.Run("should successfully delete existing dashboard template", func(t *testing.T) {
		server := setupRouter()

		// Create a mock dashboard template in the database with matching user ID
		testUserID := test_util.GetUniqueUserID()
		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = testUserID
		result := database.DB.Create(&mockDashboard)
		assert.NoError(t, result.Error, "Should be able to create mock dashboard in DB")

		templateID := int64(mockDashboard.ID)

		// Verify the template exists before deletion
		var existingTemplate api.DashboardTemplate
		err := database.DB.First(&existingTemplate, templateID).Error
		assert.NoError(t, err, "Template should exist in database before deletion")

		// Perform the DELETE request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%d", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.DeleteWidgetLayoutById(w, req, templateID)

		// Check the response
		assert.Equal(t, http.StatusNoContent, w.Code, "Expected status code 204 for successful deletion")
		assert.Empty(t, w.Body.String(), "Response body should be empty for successful deletion")

		// Verify the template no longer exists in the database
		var deletedTemplate api.DashboardTemplate
		err = database.DB.First(&deletedTemplate, templateID).Error
		assert.Error(t, err, "Template should not exist in database after deletion")
		assert.Contains(t, err.Error(), "record not found", "Should get record not found error")
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound), "Should get record not found error for deleted template")
	})

	t.Run("should return 404 for non-existent template", func(t *testing.T) {
		server := setupRouter()

		nonExistentID := int64(test_util.NonExistentID)

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%d", nonExistentID), nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.DeleteWidgetLayoutById(w, req, nonExistentID)

		// non-existent templates result in an empty template with ID 0, which fails authorization
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404 for non-existent template")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
		assert.Contains(t, errorResponse.Errors[0].Message, "record not found", "Error should mention record not found")
	})

	t.Run("should return 403 for unauthorized deletion", func(t *testing.T) {
		server := setupRouter()

		// Create a mock dashboard template with a different user ID
		templateOwnerID := test_util.GetUniqueUserID()
		requestingUserID := test_util.GetUniqueUserID()

		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = templateOwnerID
		result := database.DB.Create(&mockDashboard)
		assert.NoError(t, result.Error, "Should be able to create mock dashboard in DB")

		templateID := int64(mockDashboard.ID)

		// Try to delete with different user identity
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%d", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.DeleteWidgetLayoutById(w, req, templateID)

		assert.Equal(t, http.StatusForbidden, w.Code, "Expected status code 403 for unauthorized deletion")

		var errorResponse api.ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.Contains(t, errorResponse.Errors[0].Message, "unauthorized", "Error should mention unauthorized access")

		// Verify the template still exists in the database (wasn't deleted)
		var stillExistingTemplate api.DashboardTemplate
		err = database.DB.First(&stillExistingTemplate, templateID).Error
		assert.NoError(t, err, "Template should still exist in database after failed deletion")
		assert.Equal(t, mockDashboard.ID, stillExistingTemplate.ID, "Template ID should match")
	})
}
