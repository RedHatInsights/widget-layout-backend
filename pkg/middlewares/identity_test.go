package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
)

// Helper to run middleware test and return response recorder
func runMiddlewareTest(t *testing.T, headerValue *string, nextHandler http.HandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	if headerValue != nil {
		req.Header.Set("x-rh-identity", *headerValue)
	}
	rr := httptest.NewRecorder()
	middleware := InjectUserIdentity(nextHandler)
	middleware.ServeHTTP(rr, req)
	return rr
}

func TestInjectUserIdentityMiddleware(t *testing.T) {
	t.Run("should inject user identity into context", func(t *testing.T) {
		header := test_util.GenerateIdentityHeader()
		rr := runMiddlewareTest(t, &header, func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if GetUserIdentity(ctx).Identity.User.UserID != "user-123" {
				http.Error(w, "Identity not found in context", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		})
		if rr.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
		}
	})

	t.Run("should return 400 for invalid identity header", func(t *testing.T) {
		invalidHeader := "invalid-header"
		rr := runMiddlewareTest(t, &invalidHeader, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})

	t.Run("should return 400 for missing identity header", func(t *testing.T) {
		rr := runMiddlewareTest(t, nil, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rr.Code)
		}
	})
}
