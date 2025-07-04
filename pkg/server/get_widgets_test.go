package server_test

import (
	"encoding/json"
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

func TestGetWidgets(t *testing.T) {
	t.Run("should return list of user's dashboard templates", func(t *testing.T) {
		server := setupRouter()

		// Create test dashboard templates in the database for the test user
		testWidget1 := api.WidgetItem{
			Height:     2,
			Width:      2,
			X:          test_util.IntPTR(0),
			WidgetType: "widget1",
			Y:          test_util.IntPTR(0),
			Static:     false,
			MaxHeight:  4,
			MinHeight:  1,
		}
		testWidget2 := api.WidgetItem{
			Height:     3,
			Width:      3,
			X:          test_util.IntPTR(2),
			WidgetType: "widget2",
			Y:          test_util.IntPTR(0),
			Static:     false,
			MaxHeight:  6,
			MinHeight:  2,
		}

		tm1 := datatypes.NewJSONType([]api.WidgetItem{testWidget1})
		tm2 := datatypes.NewJSONType([]api.WidgetItem{testWidget2})

		testTemplateConfig1 := api.DashboardTemplateConfig{
			Lg: tm1,
			Md: tm1,
			Sm: tm1,
			Xl: tm1,
		}
		testTemplateConfig2 := api.DashboardTemplateConfig{
			Lg: tm2,
			Md: tm2,
			Sm: tm2,
			Xl: tm2,
		}

		// Use unique user ID for this test to avoid conflicts with other tests
		testUserID := test_util.GetUniqueUserID()

		template1 := api.DashboardTemplate{
			ID:             uint(test_util.GetUniqueID()),
			UserId:         testUserID,
			TemplateConfig: testTemplateConfig1,
		}
		template2 := api.DashboardTemplate{
			ID:             uint(test_util.GetUniqueID()),
			UserId:         testUserID,
			TemplateConfig: testTemplateConfig2,
		}

		// Save templates to database
		result1 := database.DB.Create(&template1)
		assert.NoError(t, result1.Error, "Should be able to create test template 1 in DB")
		result2 := database.DB.Create(&template2)
		assert.NoError(t, result2.Error, "Should be able to create test template 2 in DB")

		// Simulate a request to the / endpoint with matching user identity
		req, _ := http.NewRequest("GET", "/", nil)
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)
		req = withCustomIdentityContext(req, testIdentity)
		w := httptest.NewRecorder()

		server.GetWidgetLayout(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

		resp := w.Body.Bytes()
		var parsedResp []api.DashboardTemplate
		err := json.Unmarshal(resp, &parsedResp)
		assert.NoError(t, err, "Response should be valid JSON")

		assert.Equal(t, 2, len(parsedResp), "Expected two templates in response")

		// Verify that both templates are returned (order may vary)
		foundTemplate1 := false
		foundTemplate2 := false
		for _, template := range parsedResp {
			if template.ID == template1.ID {
				foundTemplate1 = true
				assert.Equal(t, template1.UserId, template.UserId, "User ID should match for template 1")
				assert.Equal(t, template1.TemplateConfig, template.TemplateConfig, "Template config should match for template 1")
			}
			if template.ID == template2.ID {
				foundTemplate2 = true
				assert.Equal(t, template2.UserId, template.UserId, "User ID should match for template 2")
				assert.Equal(t, template2.TemplateConfig, template.TemplateConfig, "Template config should match for template 2")
			}
		}
		assert.True(t, foundTemplate1, "Template 1 should be found in response")
		assert.True(t, foundTemplate2, "Template 2 should be found in response")
	})

	t.Run("should return empty list when user has no templates", func(t *testing.T) {
		server := setupRouter()

		// Create a template for a different user to ensure we only get current user's templates
		otherUserTemplate := api.DashboardTemplate{
			ID:     uint(test_util.GetUniqueID()),
			UserId: "other-user-456",
			TemplateConfig: api.DashboardTemplateConfig{
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}
		result := database.DB.Create(&otherUserTemplate)
		assert.NoError(t, result.Error, "Should be able to create other user's template in DB")

		// Request with different user identity (no templates for this user)
		req, _ := http.NewRequest("GET", "/", nil)
		differentUserIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr("different-user-789")},
			xrhidgen.Entitlements{},
		)
		req = withCustomIdentityContext(req, differentUserIdentity)
		w := httptest.NewRecorder()

		server.GetWidgetLayout(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

		resp := w.Body.Bytes()
		var parsedResp []api.DashboardTemplate
		err := json.Unmarshal(resp, &parsedResp)
		assert.NoError(t, err, "Response should be valid JSON")

		assert.Equal(t, 0, len(parsedResp), "Expected empty list when user has no templates")
	})

	t.Run("should set Content-Type to application/json", func(t *testing.T) {
		server := setupRouter()
		req, _ := http.NewRequest("GET", "/", nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()
		server.GetWidgetLayout(w, req)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")
	})

	t.Run("should return valid JSON", func(t *testing.T) {
		server := setupRouter()
		req, _ := http.NewRequest("GET", "/", nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()
		server.GetWidgetLayout(w, req)
		var js []api.DashboardTemplate
		err := json.Unmarshal(w.Body.Bytes(), &js)
		assert.NoError(t, err, "Response should be valid JSON")
	})
}
