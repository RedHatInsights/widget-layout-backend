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
	"github.com/stretchr/testify/require"
	"github.com/subpop/xrhidgen"
)

func TestSetWidgetLayoutDefaultById(t *testing.T) {
	t.Run("should set template as default successfully", func(t *testing.T) {
		server := setupRouter()

		testUserID := test_util.GetUniqueUserID()

		mockDashboard := test_util.MockDashboardTemplateWithSpecificUser(testUserID)
		mockDashboard.Default = false
		require.NoError(t, database.DB.Create(&mockDashboard).Error)

		templateID := int64(mockDashboard.ID)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/default", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.SetWidgetLayoutDefaultById(w, req, templateID)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp api.DashboardTemplate
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.True(t, resp.Default, "Returned template should be marked as default")

		// Verify in DB
		var dbTemplate api.DashboardTemplate
		require.NoError(t, database.DB.First(&dbTemplate, templateID).Error)
		assert.True(t, dbTemplate.Default)
	})

	t.Run("should unset previous default when setting new one", func(t *testing.T) {
		server := setupRouter()

		testUserID := test_util.GetUniqueUserID()

		// Create first template as default
		template1 := test_util.MockDashboardTemplateWithSpecificUser(testUserID)
		template1.TemplateBase.Name = "set-default-handler-test"
		template1.Default = true
		require.NoError(t, database.DB.Create(&template1).Error)

		// Create second template as non-default
		template2 := test_util.MockDashboardTemplateWithSpecificUser(testUserID)
		template2.TemplateBase.Name = "set-default-handler-test"
		template2.Default = false
		require.NoError(t, database.DB.Create(&template2).Error)

		templateID := int64(template2.ID)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/default", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.SetWidgetLayoutDefaultById(w, req, templateID)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify template1 is no longer default
		var dbTemplate1 api.DashboardTemplate
		require.NoError(t, database.DB.First(&dbTemplate1, template1.ID).Error)
		assert.False(t, dbTemplate1.Default, "Previous default should be unset")

		// Verify template2 is now default
		var dbTemplate2 api.DashboardTemplate
		require.NoError(t, database.DB.First(&dbTemplate2, template2.ID).Error)
		assert.True(t, dbTemplate2.Default, "New default should be set")
	})

	t.Run("should return 404 for non-existent template", func(t *testing.T) {
		server := setupRouter()

		nonExistentID := int64(test_util.NonExistentID)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/default", nonExistentID), nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.SetWidgetLayoutDefaultById(w, req, nonExistentID)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.NotEmpty(t, errorResponse.Errors)
	})

	t.Run("should return 403 for unauthorized user", func(t *testing.T) {
		server := setupRouter()

		ownerID := test_util.GetUniqueUserID()
		requestingUserID := test_util.GetUniqueUserID()

		mockDashboard := test_util.MockDashboardTemplateWithSpecificUser(ownerID)
		require.NoError(t, database.DB.Create(&mockDashboard).Error)

		templateID := int64(mockDashboard.ID)

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/default", templateID), nil)
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.SetWidgetLayoutDefaultById(w, req, templateID)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.Contains(t, errorResponse.Errors[0].Message, "unauthorized")
	})

	t.Run("should panic when identity is missing from context", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("PUT", "/123/default", nil)
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			server.SetWidgetLayoutDefaultById(w, req, int64(test_util.NoDBTestID))
		}, "Should panic when identity is missing from context")
	})
}
