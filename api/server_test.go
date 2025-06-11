package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *api.Server {
	r := chi.NewRouter()
	server := api.NewServer(r)
	return server
}

type MockWidget struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
}

func prepareWidgetPayload(resp MockWidget) api.Widget {
	return api.Widget{
		Id:       &resp.Id,
		Name:     &resp.Name,
		Position: &resp.Position,
	}
}

func TestGetWidgets(t *testing.T) {
	t.Run("should return list of all widgets", func(t *testing.T) {
		server := setupRouter()
		expectedResponse := []api.Widget{
			prepareWidgetPayload(MockWidget{
				Id:       "id",
				Name:     "name",
				Position: 1,
			})}

		// Simulate a request to the /widgets endpoint
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		server.GetWidgetLayout(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		resp := w.Body.Bytes()
		var parsedResp []api.Widget
		json.Unmarshal(resp, &parsedResp)
		assert.Equal(t, len(expectedResponse), len(parsedResp), "Expected one widget in response")
		assert.EqualValues(t, expectedResponse, parsedResp, "Expected widget data to match")
	})
}
