package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
)

func TestImportWidgetLayoutById(t *testing.T) {
	t.Run("should successfully import widget", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "imported-dashboard",
				DisplayName: "Imported Dashboard",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code, "Expected status code 200 for successful import")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		// Decode response
		var importedTemplate api.DashboardTemplate
		err = json.NewDecoder(w.Body).Decode(&importedTemplate)
		require.NoError(t, err, "Should be able to decode imported template response")

		// Verify the imported template has correct data
		assert.NotZero(t, importedTemplate.ID, "Imported template should have a new ID")
		assert.Equal(t, testUserID, importedTemplate.UserId, "Imported template should belong to requesting user")
		assert.Equal(t, "imported-dashboard", importedTemplate.TemplateBase.Name, "Template name should match")
		assert.Equal(t, "Imported Dashboard", importedTemplate.TemplateBase.DisplayName, "Template display name should match")
		assert.NotEmpty(t, importedTemplate.CreatedAt, "Imported template should have creation timestamp")
		assert.NotEmpty(t, importedTemplate.UpdatedAt, "Imported template should have update timestamp")
		assert.False(t, importedTemplate.Default, "Imported template should not be default")

		// Verify the template config was imported correctly
		widgets := importedTemplate.TemplateConfig.Sm.Data()
		require.Len(t, widgets, 1, "Should have one widget from import data")
		assert.Equal(t, "import-widget", widgets[0].WidgetType, "Widget should match import data")
		assert.Equal(t, 1, widgets[0].Width, "Widget width should match import data")
		assert.Equal(t, 2, widgets[0].Height, "Widget height should match import data")

		// Verify the template was actually saved to the database
		var dbTemplate api.DashboardTemplate
		err = database.DB.First(&dbTemplate, importedTemplate.ID).Error
		require.NoError(t, err, "Imported template should be saved in database")
		assert.Equal(t, testUserID, dbTemplate.UserId, "Database template should belong to requesting user")
		assert.Equal(t, "imported-dashboard", dbTemplate.TemplateBase.Name, "Database template name should match")
	})

	t.Run("should return 400 for invalid template data - no name", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "",
				DisplayName: "Imported Dashboard",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - no DisplayName", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Width", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      0,
						Height:     2,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Width", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      5,
						Height:     2,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Height", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     0,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Height", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     6,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Height", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(0),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Height", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(0),
						MinHeight:  test_util.IntPTR(0),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Height", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(1),
						MinHeight:  test_util.IntPTR(5),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid X", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(4),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid X", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(-1),
						Y:          test_util.IntPTR(0),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid Y", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(-1),
						WidgetType: "import-widget",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for invalid template data - Invalid WidgetType", func(t *testing.T) {
		server := setupRouter()
		testUserID := test_util.GetUniqueUserID()

		importData := api.ExportWidgetDashboardTemplateResponse{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test",
				DisplayName: "Test",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     1,
						MaxHeight:  test_util.IntPTR(5),
						MinHeight:  test_util.IntPTR(1),
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "",
						Static:     false,
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		// Marshal import data to JSON
		requestBody, err := json.Marshal(importData)
		require.NoError(t, err, "Should be able to marshal import data")

		// Create POST request with import data
		req, _ := http.NewRequest("POST", "/import", bytes.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = withCustomIdentityContext(req, test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		))
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for invalid data")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"), "Content-Type should be application/json")

		var errorResponse api.ErrorResponse
		err = json.NewDecoder(w.Body).Decode(&errorResponse)
		assert.NoError(t, err, "Should be able to decode error response")
		assert.NotEmpty(t, errorResponse.Errors, "Error response should contain error messages")
	})

	t.Run("should return 400 for empty request body", func(t *testing.T) {
		server := setupRouter()

		req, _ := http.NewRequest("POST", "/import", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for empty body")
	})

	t.Run("should handle malformed JSON gracefully", func(t *testing.T) {
		server := setupRouter()

		// Truly malformed JSON that will fail to parse
		malformedJSON := `{"templateConfig": {"lg": [{"height":}]}}`

		req, _ := http.NewRequest("POST", "/import", strings.NewReader(malformedJSON))
		req.Header.Set("Content-Type", "application/json")
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.ImportWidgetLayout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400 for malformed JSON")
	})
}
