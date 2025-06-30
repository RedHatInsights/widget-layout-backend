package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
)

func TestUpdateWidgetLayoutById(t *testing.T) {
	t.Run("should return 400 for invalid JSON in request body", func(t *testing.T) {
		server := setupRouter()

		// Invalid JSON
		invalidJSON := `{"invalid": json}`

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/%d", test_util.NoDBTestID), strings.NewReader(invalidJSON))
		req.Header.Set("Content-Type", "application/json")
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.UpdateWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid JSON")
		assert.Contains(t, w.Body.String(), "Invalid request body", "Expected error message for invalid JSON")
	})

	t.Run("should return 400 for empty request body", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/%d", test_util.NoDBTestID), strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.UpdateWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for empty body")
	})

	t.Run("should handle malformed JSON gracefully", func(t *testing.T) {
		server := setupRouter()

		// Truly malformed JSON that will fail to parse
		malformedJSON := `{"templateConfig": {"lg": [{"height":}]}}`

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/%d", test_util.NoDBTestID), strings.NewReader(malformedJSON))
		req.Header.Set("Content-Type", "application/json")
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.UpdateWidgetLayoutById(w, req, int64(test_util.NoDBTestID))

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for malformed JSON")
	})

	t.Run("should parse valid JSON request body successfully", func(t *testing.T) {
		testUserID := test_util.GetUniqueUserID()
		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = testUserID
		server := setupRouter()
		result := database.DB.Create(&mockDashboard)
		assert.NoError(t, result.Error, "Should be able to create mock dashboard in DB")

		templateID := int64(mockDashboard.ID)

		// Valid structure but with complete widget data
		testWidget := api.WidgetItem{
			Height:     3,
			Width:      2,
			X:          0,
			WidgetType: "test-widget",
			Y:          0,
			Static:     false,
			Title:      "Test Widget",
			MaxHeight:  4,
			MinHeight:  1,
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})
		validTemplate := api.DashboardTemplate{
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: tm,
				Md: tm,
				Sm: tm,
				Xl: tm,
			},
		}

		requestBody, err := json.Marshal(validTemplate)
		assert.NoError(t, err, "Should be able to marshal valid template")

		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/%d", templateID), strings.NewReader(string(requestBody)))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.UpdateWidgetLayoutById(w, req, templateID)

		assert.NotEqual(t, http.StatusBadRequest, w.Code, "Should not return 400 for valid JSON structure")
		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for valid JSON structure")
		var responseTemplate api.DashboardTemplate
		err = json.NewDecoder(w.Body).Decode(&responseTemplate)
		assert.NoError(t, err, "Response should be valid JSON")
		assert.Equal(t, validTemplate.TemplateConfig, responseTemplate.TemplateConfig, "Expected template config to match")

		// Check that the widget details match
		responseWidgets := responseTemplate.TemplateConfig.Lg.Data()
		assert.Greater(t, len(responseWidgets), 0, "Should have at least one widget")
		if len(responseWidgets) > 0 {
			assert.Equal(t, validTemplate.TemplateConfig.Lg.Data()[0].Height, responseWidgets[0].Height)
		}
	})
}
