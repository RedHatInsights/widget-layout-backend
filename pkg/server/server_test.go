package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/RedHatInsights/widget-layout-backend/pkg/server"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func setupRouter() *server.Server {
	r := chi.NewRouter()
	server := server.NewServer(r)
	return server
}

type MockWidget struct {
	H      int    `json:"h"`
	W      int    `json:"w"`
	X      int    `json:"x"`
	I      string `json:"i"`
	Y      int    `json:"y"`
	Static bool   `json:"static"`
	Title  string `json:"title"`
	MaxH   int    `json:"maxH"`
	MinH   int    `json:"minH"`
}

func withIdentityContext(req *http.Request) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, config.IdentityContextKey, test_util.GenerateIdentityStruct())
	return req.WithContext(ctx)
}

func TestGetWidgets(t *testing.T) {
	t.Run("should return list of all widgets", func(t *testing.T) {
		server := setupRouter()
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
			ID:             123456,
			TemplateConfig: testTemplateConfig,
		}
		expectedResponse := api.DashboardTemplateList{
			testTemplate,
		}

		// Simulate a request to the /widgets endpoint
		req, _ := http.NewRequest("GET", "/", nil)
		req = withIdentityContext(req)
		w := httptest.NewRecorder()

		server.GetWidgetLayout(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		resp := w.Body.Bytes()
		var parsedResp []models.DashboardTemplate
		json.Unmarshal(resp, &parsedResp)
		assert.Equal(t, len(expectedResponse), len(parsedResp), "Expected one widget in response")
		assert.EqualValues(t, expectedResponse, parsedResp, "Expected widget data to match")
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
		var js interface{}
		err := json.Unmarshal(w.Body.Bytes(), &js)
		assert.NoError(t, err, "Response should be valid JSON")
	})
}
