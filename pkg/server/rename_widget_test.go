package server_test

import (
	"bytes"
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

func TestRenameWidgetLayoutById(t *testing.T) {
	t.Run("should rename template successfully", func(t *testing.T) {
		server := setupRouter()

		testUserID := test_util.GetUniqueUserID()
		mockDashboard := test_util.MockDashboardTemplateWithSpecificUser(testUserID)
		mockDashboard.DashboardName = "Old Name"
		require.NoError(t, database.DB.Create(&mockDashboard).Error)

		templateID := int64(mockDashboard.ID)
		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "New Name",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/rename", templateID), bytes.NewReader(body))
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, templateID)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp api.DashboardTemplate
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "New Name", resp.DashboardName)

		// Verify persisted in DB
		var dbTemplate api.DashboardTemplate
		require.NoError(t, database.DB.First(&dbTemplate, templateID).Error)
		assert.Equal(t, "New Name", dbTemplate.DashboardName)
	})

	t.Run("should return 400 for invalid request body", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("PUT", "/123/rename", bytes.NewReader([]byte("invalid json")))
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.NotEmpty(t, errorResponse.Errors)
		assert.Contains(t, errorResponse.Errors[0].Message, "Invalid request body")
	})

	t.Run("should return 400 for empty dashboard name", func(t *testing.T) {
		server := setupRouter()

		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "",
		})

		req, _ := http.NewRequest("PUT", "/123/rename", bytes.NewReader(body))
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.Contains(t, errorResponse.Errors[0].Message, "dashboardName is required")
	})

	t.Run("should return 400 for whitespace-only dashboard name", func(t *testing.T) {
		server := setupRouter()

		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "   ",
		})

		req, _ := http.NewRequest("PUT", "/123/rename", bytes.NewReader(body))
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.Contains(t, errorResponse.Errors[0].Message, "dashboardName is required")
	})

	t.Run("should return 404 for non-existent template", func(t *testing.T) {
		server := setupRouter()

		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "New Name",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/rename", test_util.NonExistentID), bytes.NewReader(body))
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, int64(test_util.NonExistentID))

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
		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "Unauthorized Rename",
		})

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/%d/rename", templateID), bytes.NewReader(body))
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(requestingUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.RenameWidgetLayoutById(w, req, templateID)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var errorResponse api.ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&errorResponse))
		assert.Contains(t, errorResponse.Errors[0].Message, "unauthorized")
	})

	t.Run("should panic when identity is missing from context", func(t *testing.T) {
		server := setupRouter()

		body, _ := json.Marshal(api.RenameWidgetDashboardTemplateRequest{
			DashboardName: "Valid Name",
		})

		req, _ := http.NewRequest("PUT", "/123/rename", bytes.NewReader(body))
		// No identity context set
		w := httptest.NewRecorder()

		assert.Panics(t, func() {
			server.RenameWidgetLayoutById(w, req, int64(test_util.NoDBTestID))
		}, "Should panic when identity is missing from context")
	})
}
