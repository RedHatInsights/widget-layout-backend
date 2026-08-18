package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
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

func TestGetUserIdentity(t *testing.T) {
	t.Run("should panic when identity is not in context", func(t *testing.T) {
		ctx := context.Background()
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic but did not get one")
				return
			}
			err, ok := r.(error)
			if !ok {
				t.Error("Expected panic with error type")
				return
			}
			if err.Error() != "identity not found in context" {
				t.Errorf("Expected panic message 'identity not found in context', got '%s'", err.Error())
			}
		}()
		GetUserIdentity(ctx)
	})

	t.Run("should panic when context has wrong type for identity key", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), config.IdentityContextKey, "wrong-type")
		defer func() {
			r := recover()
			if r == nil {
				t.Error("Expected panic but did not get one")
				return
			}
		}()
		GetUserIdentity(ctx)
	})

	t.Run("should return identity when present in context", func(t *testing.T) {
		id := test_util.GenerateIdentityStruct()
		ctx := context.WithValue(context.Background(), config.IdentityContextKey, id)
		result := GetUserIdentity(ctx)
		if result.Identity.User.UserID != "user-123" {
			t.Errorf("Expected user ID 'user-123', got '%s'", result.Identity.User.UserID)
		}
	})
}
