package server_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/RedHatInsights/widget-layout-backend/pkg/server"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/go-chi/chi/v5"
	"github.com/subpop/xrhidgen"
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
	test_util.ResetUserIDGenerator()

	// Reserve hardcoded IDs that are still used in some tests that don't create DB records
	test_util.ReserveID(test_util.NoDBTestID)    // Used in update tests that don't create DB records
	test_util.ReserveID(test_util.NonExistentID) // Used for non-existent ID tests

	// Reserve commonly used user IDs to prevent conflicts
	test_util.ReserveUserID("user-123") // Used in test utilities (MockDashboardTemplate, GenerateIdentityStruct)

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

func withCustomIdentityContext(req *http.Request, identity interface{}) *http.Request {
	ctx := context.Background()
	ctx = context.WithValue(ctx, config.IdentityContextKey, identity)
	return req.WithContext(ctx)
}

func withUniqueUserIdentityContext(req *http.Request) (*http.Request, string) {
	userID := test_util.GetUniqueUserID()
	identity := test_util.GenerateIdentityStructFromTemplate(
		xrhidgen.Identity{},
		xrhidgen.User{UserID: stringPtr(userID)},
		xrhidgen.Entitlements{},
	)
	ctx := context.Background()
	ctx = context.WithValue(ctx, config.IdentityContextKey, identity)
	return req.WithContext(ctx), userID
}
