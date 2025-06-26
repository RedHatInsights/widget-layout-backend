package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/RedHatInsights/widget-layout-backend/pkg/server"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Helper function to convert string to string pointer
func stringPtr(s string) *string {
	return &s
}

func TestMain(m *testing.M) {
	cfg := config.GetConfig()
	now := time.Now().UnixNano()
	dbName := fmt.Sprintf("%d-dashboard-template.db", now)
	cfg.TestMode = true
	cfg.DatabaseConfig.DBName = dbName

	database.InitDb()
	// Load the models into the tmp database
	database.DB.AutoMigrate(
		&models.DashboardTemplate{},
	)

	// Reset the unique ID generator for clean tests - this should be done before every test run
	test_util.ResetIDGenerator()

	// Reserve hardcoded IDs that are still used in some tests that don't create DB records
	test_util.ReserveID(test_util.NoDBTestID)    // Used in update tests that don't create DB records
	test_util.ReserveID(test_util.NonExistentID) // Used for non-existent ID tests

	exitCode := m.Run()

	err := os.Remove(dbName)

	if err != nil {
		fmt.Printf("Error removing test database file %s: %v\n", dbName, err)
	}

	os.Exit(exitCode)
}

func setupRouter() *server.Server {
	r := chi.NewRouter()
	server := server.NewServer(r)
	return server
}

func withIdentityContext(req *http.Request) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, config.IdentityContextKey, test_util.GenerateIdentityStruct())
	return req.WithContext(ctx)
}

func TestGetWidgets(t *testing.T) {
	t.Run("should return list of user's dashboard templates", func(t *testing.T) {
		server := setupRouter()

		// Create test dashboard templates in the database for the test user
		testWidget1 := api.WidgetItem{
			Height:     2,
			Width:      2,
			X:          0,
			WidgetType: "widget1",
			Y:          0,
			Static:     false,
			Title:      "Sample Widget 1",
			MaxHeight:  4,
			MinHeight:  1,
		}
		testWidget2 := api.WidgetItem{
			Height:     3,
			Width:      3,
			X:          2,
			WidgetType: "widget2",
			Y:          0,
			Static:     false,
			Title:      "Sample Widget 2",
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

		template1 := api.DashboardTemplate{
			ID:             uint(test_util.GetUniqueID()),
			UserId:         "user-123", // Match the identity from withIdentityContext
			TemplateConfig: testTemplateConfig1,
		}
		template2 := api.DashboardTemplate{
			ID:             uint(test_util.GetUniqueID()),
			UserId:         "user-123", // Match the identity from withIdentityContext
			TemplateConfig: testTemplateConfig2,
		}

		// Save templates to database
		result1 := database.DB.Create(&template1)
		assert.NoError(t, result1.Error, "Should be able to create test template 1 in DB")
		result2 := database.DB.Create(&template2)
		assert.NoError(t, result2.Error, "Should be able to create test template 2 in DB")

		// Simulate a request to the / endpoint
		req, _ := http.NewRequest("GET", "/", nil)
		req = withIdentityContext(req)
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
		ctx := context.Background()
		differentUserIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr("different-user-789")},
			xrhidgen.Entitlements{},
		)
		ctx = context.WithValue(ctx, config.IdentityContextKey, differentUserIdentity)
		req = req.WithContext(ctx)
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

func TestGetWidgetLayoutById(t *testing.T) {
	t.Run("should return specific widget by ID for authorized user", func(t *testing.T) {
		server := setupRouter()

		// Create a dashboard template in the database with matching user ID
		templateID := test_util.GetUniqueID()
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
			UserId:         "user-123", // Match the identity from withIdentityContext
			TemplateConfig: testTemplateConfig,
		}

		// Save to database
		result := database.DB.Create(&testTemplate)
		assert.NoError(t, result.Error, "Should be able to create test template in DB")

		// Test with the correct ID and authorized user
		req, _ := http.NewRequest("GET", fmt.Sprintf("/%d", templateID), nil)
		req = withIdentityContext(req)
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
		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: "different-user", // Different from "user-123" in withIdentityContext
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
		req = withIdentityContext(req) // This uses "user-123" identity
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
		testTemplate := api.DashboardTemplate{
			ID:     templateID,
			UserId: "user-123", // Match the identity from withIdentityContext
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
		req = withIdentityContext(req)
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
		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = "user-123" // Match the identity from withIdentityContext
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
		req = withIdentityContext(req)
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

func TestDeleteWidgetLayoutById(t *testing.T) {
	t.Run("should successfully delete existing dashboard template", func(t *testing.T) {
		server := setupRouter()

		// Create a mock dashboard template in the database with matching user ID
		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = "user-123" // Match the identity from withIdentityContext
		result := database.DB.Create(&mockDashboard)
		assert.NoError(t, result.Error, "Should be able to create mock dashboard in DB")

		templateID := int64(mockDashboard.ID)

		// Verify the template exists before deletion
		var existingTemplate api.DashboardTemplate
		err := database.DB.First(&existingTemplate, templateID).Error
		assert.NoError(t, err, "Template should exist in database before deletion")

		// Perform the DELETE request
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%d", templateID), nil)
		req = withIdentityContext(req)
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
		mockDashboard := test_util.MockDashboardTemplate()
		mockDashboard.UserId = "different-user-id"
		result := database.DB.Create(&mockDashboard)
		assert.NoError(t, result.Error, "Should be able to create mock dashboard in DB")

		templateID := int64(mockDashboard.ID)

		// Try to delete with different user identity
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/%d", templateID), nil)
		req = withIdentityContext(req) // This uses "user-123" identity
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

func TestCopyWidgetLayoutById(t *testing.T) {
	t.Run("should successfully copy existing dashboard template", func(t *testing.T) {
		server := setupRouter()

		// Create an original dashboard template in the database
		originalDashboard := test_util.MockDashboardTemplate()
		originalDashboard.UserId = "different-user" // Different user - to verify copying works across users
		originalDashboard.TemplateBase.Name = "Original Template"
		originalDashboard.TemplateBase.DisplayName = "Original Display Name"
		result := database.DB.Create(&originalDashboard)
		assert.NoError(t, result.Error, "Should be able to create original dashboard in DB")

		templateID := int64(originalDashboard.ID)

		// Perform the COPY request
		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/copy", templateID), nil)
		req = withIdentityContext(req)
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
		assert.Equal(t, "user-123", copiedTemplate.UserId, "Copied template should belong to copying user")
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
		assert.Equal(t, "different-user", originalFromDB.UserId, "Original template should belong to original user")
		assert.Equal(t, "user-123", copiedFromDB.UserId, "Copied template should belong to copying user")
	})

	t.Run("should return 404 for non-existent template", func(t *testing.T) {
		server := setupRouter()

		nonExistentID := int64(test_util.NonExistentID)

		req, _ := http.NewRequest("POST", fmt.Sprintf("/%d/copy", nonExistentID), nil)
		req = withIdentityContext(req)
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
			X:          1,
			WidgetType: "special-widget",
			Y:          2,
			Static:     true,
			Title:      "Special Test Widget",
			MaxHeight:  6,
			MinHeight:  2,
		}
		tm := datatypes.NewJSONType([]api.WidgetItem{testWidget})

		templateID := test_util.GetUniqueID()
		originalDashboard := api.DashboardTemplate{
			ID: templateID,
			TemplateBase: api.DashboardTemplateBase{
				Name:        "Widget Config Test",
				DisplayName: "Widget Configuration Test Template",
			},
			UserId: "some-other-user",
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
		req = withIdentityContext(req)
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
			assert.Equal(t, originalWidgets[0].Title, copiedWidgets[0].Title, "Widget title should match")
			assert.Equal(t, originalWidgets[0].Static, copiedWidgets[0].Static, "Widget static property should match")
			assert.Equal(t, originalWidgets[0].MaxHeight, copiedWidgets[0].MaxHeight, "Widget max height should match")
			assert.Equal(t, originalWidgets[0].MinHeight, copiedWidgets[0].MinHeight, "Widget min height should match")
		}
	})
}
