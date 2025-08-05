package service_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/RedHatInsights/widget-layout-backend/pkg/test_util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subpop/xrhidgen"
	"gorm.io/datatypes"
)

func TestMain(m *testing.M) {
	cfg := config.GetConfig()
	now := time.Now().UnixNano()
	dbName := fmt.Sprintf("%d-service-dashboard-template.db", now)
	cfg.TestMode = true
	cfg.DatabaseConfig.DBName = dbName

	database.InitDb()
	// Load the models into the tmp database
	database.DB.AutoMigrate(
		&models.DashboardTemplate{},
	)

	// Reset the unique ID generator for clean tests
	test_util.ResetIDGenerator()
	test_util.ResetUserIDGenerator()

	// Reserve hardcoded IDs that are still used in some tests
	test_util.ReserveID(test_util.NoDBTestID)
	test_util.ReserveID(test_util.NonExistentID)

	exitCode := m.Run()

	err := os.Remove(dbName)
	if err != nil {
		fmt.Printf("Error removing test database file %s: %v\n", dbName, err)
	}

	os.Exit(exitCode)
}

func TestForkBaseTemplate(t *testing.T) {
	t.Run("should successfully fork existing base template", func(t *testing.T) {
		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "fork-service-test",
			DisplayName: "Fork Service Test",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{
					{
						Width:      1,
						Height:     2,
						MaxHeight:  5,
						MinHeight:  1,
						X:          test_util.IntPTR(0),
						Y:          test_util.IntPTR(0),
						WidgetType: "service-fork-widget",
					},
				}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Fork the base template
		forkedTemplate, status, err := service.ForkBaseTemplate("fork-service-test", testIdentity)

		// Verify success
		assert.NoError(t, err, "ForkBaseTemplate should not return an error")
		assert.Equal(t, http.StatusOK, status, "Status should be 200 OK")

		// Verify forked template data
		assert.NotZero(t, forkedTemplate.ID, "Forked template should have a new ID")
		assert.Equal(t, testUserID, forkedTemplate.UserId, "Forked template should belong to requesting user")
		assert.Equal(t, "fork-service-test", forkedTemplate.TemplateBase.Name, "Template base name should match")
		assert.Equal(t, "Fork Service Test", forkedTemplate.TemplateBase.DisplayName, "Template display name should match")

		// Verify template config was copied
		widgets := forkedTemplate.TemplateConfig.Sm.Data()
		require.Len(t, widgets, 1, "Should have one widget from base template")
		assert.Equal(t, "service-fork-widget", widgets[0].WidgetType, "Widget should match base template")

		// Verify template was saved to database
		var dbTemplate api.DashboardTemplate
		err = database.DB.First(&dbTemplate, forkedTemplate.ID).Error
		assert.NoError(t, err, "Forked template should be saved in database")
		assert.Equal(t, testUserID, dbTemplate.UserId, "Database template should belong to requesting user")
	})

	t.Run("should return 404 for non-existent base template", func(t *testing.T) {
		// Reset registry to ensure no templates exist
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Try to fork non-existent template
		forkedTemplate, status, err := service.ForkBaseTemplate("non-existent-template", testIdentity)

		// Verify error response
		assert.Error(t, err, "ForkBaseTemplate should return an error for non-existent template")
		assert.Equal(t, http.StatusNotFound, status, "Status should be 404 Not Found")
		assert.Equal(t, api.DashboardTemplate{}, forkedTemplate, "Should return empty template on error")
		assert.Contains(t, err.Error(), "base template", "Error message should mention base template")
		assert.Contains(t, err.Error(), "not found", "Error message should mention not found")
	})

	t.Run("should create separate templates for different users", func(t *testing.T) {
		// Reset registry and add a base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "shared-fork-test",
			DisplayName: "Shared Fork Test",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		user1ID := test_util.GetUniqueUserID()
		user2ID := test_util.GetUniqueUserID()
		identity1 := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(user1ID)},
			xrhidgen.Entitlements{},
		)
		identity2 := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(user2ID)},
			xrhidgen.Entitlements{},
		)

		// Fork template as first user
		template1, status1, err1 := service.ForkBaseTemplate("shared-fork-test", identity1)
		assert.NoError(t, err1, "First fork should succeed")
		assert.Equal(t, http.StatusOK, status1, "First fork status should be 200")

		// Fork same template as second user
		template2, status2, err2 := service.ForkBaseTemplate("shared-fork-test", identity2)
		assert.NoError(t, err2, "Second fork should succeed")
		assert.Equal(t, http.StatusOK, status2, "Second fork status should be 200")

		// Verify templates are separate
		assert.NotEqual(t, template1.ID, template2.ID, "Templates should have different IDs")
		assert.Equal(t, user1ID, template1.UserId, "First template should belong to first user")
		assert.Equal(t, user2ID, template2.UserId, "Second template should belong to second user")
		assert.Equal(t, template1.TemplateBase.Name, template2.TemplateBase.Name, "Both should have same base name")
	})

	t.Run("should preserve complex template configuration", func(t *testing.T) {
		// Reset registry and add a complex base template
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}

		complexWidgets := []api.WidgetItem{
			{
				Width:      2,
				Height:     3,
				MaxHeight:  6,
				MinHeight:  2,
				X:          test_util.IntPTR(0),
				Y:          test_util.IntPTR(0),
				WidgetType: "complex-widget-1",
				Static:     true,
			},
			{
				Width:      1,
				Height:     4,
				MaxHeight:  8,
				MinHeight:  1,
				X:          test_util.IntPTR(2),
				Y:          test_util.IntPTR(0),
				WidgetType: "complex-widget-2",
				Static:     false,
			},
		}

		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "complex-fork-test",
			DisplayName: "Complex Fork Test",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType(complexWidgets[:1]), // Only first widget for small screens
				Md: datatypes.NewJSONType(complexWidgets),     // Both widgets for medium and above
				Lg: datatypes.NewJSONType(complexWidgets),
				Xl: datatypes.NewJSONType(complexWidgets),
			},
		}

		service.BaseTemplateRegistry.AddBase(baseTemplate)

		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Fork the complex template
		forkedTemplate, status, err := service.ForkBaseTemplate("complex-fork-test", testIdentity)

		assert.NoError(t, err, "Complex fork should succeed")
		assert.Equal(t, http.StatusOK, status, "Status should be 200")

		// Verify small screen config (1 widget)
		smWidgets := forkedTemplate.TemplateConfig.Sm.Data()
		require.Len(t, smWidgets, 1, "Small screen should have 1 widget")
		assert.Equal(t, "complex-widget-1", smWidgets[0].WidgetType, "Small screen widget should match")

		// Verify large screen config (2 widgets)
		lgWidgets := forkedTemplate.TemplateConfig.Lg.Data()
		require.Len(t, lgWidgets, 2, "Large screen should have 2 widgets")
		assert.Equal(t, "complex-widget-1", lgWidgets[0].WidgetType, "First widget should match")
		assert.Equal(t, "complex-widget-2", lgWidgets[1].WidgetType, "Second widget should match")
		assert.Equal(t, true, lgWidgets[0].Static, "First widget static property should be preserved")
		assert.Equal(t, false, lgWidgets[1].Static, "Second widget static property should be preserved")
	})
}

// Helper function to create a test template with specific base name
func createTestTemplate(userID, baseName, displayName string) api.DashboardTemplate {
	template := test_util.MockDashboardTemplateWithSpecificUser(userID)
	template.TemplateBase.Name = baseName
	template.TemplateBase.DisplayName = displayName
	return template
}

func TestGetUserTemplates(t *testing.T) {
	t.Run("should filter templates by templateBase.Name when DashboardType is provided", func(t *testing.T) {
		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Create templates with different base names
		template1 := createTestTemplate(testUserID, "dashboard-type-1", "Dashboard Type 1")
		template2 := createTestTemplate(testUserID, "dashboard-type-2", "Dashboard Type 2")
		template3 := createTestTemplate(testUserID, "dashboard-type-1", "Dashboard Type 1 Copy")

		// Save templates to database
		require.NoError(t, database.DB.Create(&template1).Error)
		require.NoError(t, database.DB.Create(&template2).Error)
		require.NoError(t, database.DB.Create(&template3).Error)

		// Test filtering by dashboard-type-1
		params := api.GetWidgetLayoutParams{DashboardType: stringPtr("dashboard-type-1")}
		templates, status, err := service.GetUserTemplates(testIdentity, params)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Len(t, templates, 2, "Should return 2 templates with dashboard-type-1")
		
		// Verify both returned templates have the correct base name
		for _, template := range templates {
			assert.Equal(t, "dashboard-type-1", template.TemplateBase.Name)
			assert.Equal(t, testUserID, template.UserId)
		}
	})

	t.Run("should auto-create template from base when user has none", func(t *testing.T) {
		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Reset and add a base template to the registry
		service.BaseTemplateRegistry = api.BaseWidgetDashboardTemplateRegistry{}
		baseTemplate := api.BaseWidgetDashboardTemplate{
			Name:        "auto-create-test",
			DisplayName: "Auto Create Test",
			TemplateConfig: api.DashboardTemplateConfig{
				Sm: datatypes.NewJSONType([]api.WidgetItem{}),
				Md: datatypes.NewJSONType([]api.WidgetItem{}),
				Lg: datatypes.NewJSONType([]api.WidgetItem{}),
				Xl: datatypes.NewJSONType([]api.WidgetItem{}),
			},
		}
		service.BaseTemplateRegistry.AddBase(baseTemplate)

		// Test filtering by a dashboard type that exists as base template but user has no templates
		params := api.GetWidgetLayoutParams{DashboardType: stringPtr("auto-create-test")}
		templates, status, err := service.GetUserTemplates(testIdentity, params)

		assert.NoError(t, err, "Should not return error when auto-creating from base template")
		assert.Equal(t, http.StatusNotFound, status, "Should return 404 status (but with created template)")
		assert.Len(t, templates, 1, "Should return the newly created template")
		assert.Equal(t, "auto-create-test", templates[0].TemplateBase.Name)
		assert.Equal(t, testUserID, templates[0].UserId)
		assert.True(t, templates[0].Default, "Auto-created template should be set as default")

		// Verify template was actually saved to database
		var dbTemplate api.DashboardTemplate
		err = database.DB.Where("user_id = ? AND name = ?", testUserID, "auto-create-test").First(&dbTemplate).Error
		assert.NoError(t, err, "Auto-created template should be saved in database")
	})

	t.Run("should return 404 error when filtering by non-existent base template", func(t *testing.T) {
		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Create a template with a different base name for this user
		template := createTestTemplate(testUserID, "existing-dashboard", "Existing Dashboard")
		require.NoError(t, database.DB.Create(&template).Error)

		// Test filtering by non-existent dashboard type (no base template exists)
		params := api.GetWidgetLayoutParams{DashboardType: stringPtr("non-existent-dashboard")}
		templates, status, err := service.GetUserTemplates(testIdentity, params)

		assert.Error(t, err, "Should return error when base template doesn't exist")
		assert.Equal(t, http.StatusNotFound, status, "Should return 404 when base template not found")
		assert.Nil(t, templates, "Should return nil templates on error")
		assert.Contains(t, err.Error(), "base template", "Error should mention base template")
	})

	t.Run("should return all user templates when DashboardType is not provided", func(t *testing.T) {
		testUserID := test_util.GetUniqueUserID()
		testIdentity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(testUserID)},
			xrhidgen.Entitlements{},
		)

		// Create multiple templates with different base names
		templates := []api.DashboardTemplate{
			createTestTemplate(testUserID, "type-a", "Type A"),
			createTestTemplate(testUserID, "type-b", "Type B"),
			createTestTemplate(testUserID, "type-c", "Type C"),
		}

		// Save templates to database
		for _, template := range templates {
			require.NoError(t, database.DB.Create(&template).Error)
		}

		// Test without filtering (DashboardType is nil)
		params := api.GetWidgetLayoutParams{DashboardType: nil}
		result, status, err := service.GetUserTemplates(testIdentity, params)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Len(t, result, 3, "Should return all 3 templates when no filter is applied")
		
		// Verify all templates belong to the test user and extract names
		names := make([]string, len(result))
		for i, template := range result {
			assert.Equal(t, testUserID, template.UserId)
			names[i] = template.TemplateBase.Name
		}
		
		// Verify we have all expected template names
		assert.Contains(t, names, "type-a")
		assert.Contains(t, names, "type-b")
		assert.Contains(t, names, "type-c")
	})

	t.Run("should only return templates for the requesting user when filtering", func(t *testing.T) {
		user1ID := test_util.GetUniqueUserID()
		user2ID := test_util.GetUniqueUserID()
		user1Identity := test_util.GenerateIdentityStructFromTemplate(
			xrhidgen.Identity{},
			xrhidgen.User{UserID: stringPtr(user1ID)},
			xrhidgen.Entitlements{},
		)

		// Create templates for both users with the same base name
		template1 := createTestTemplate(user1ID, "shared-dashboard-type", "User 1 Dashboard")
		template2 := createTestTemplate(user2ID, "shared-dashboard-type", "User 2 Dashboard")

		// Save templates to database
		require.NoError(t, database.DB.Create(&template1).Error)
		require.NoError(t, database.DB.Create(&template2).Error)

		// Test filtering as user1 - should only get user1's template
		params := api.GetWidgetLayoutParams{DashboardType: stringPtr("shared-dashboard-type")}
		templates, status, err := service.GetUserTemplates(user1Identity, params)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, status)
		assert.Len(t, templates, 1, "Should return only 1 template for user1")
		assert.Equal(t, "shared-dashboard-type", templates[0].TemplateBase.Name)
		assert.Equal(t, user1ID, templates[0].UserId)
		assert.Equal(t, "User 1 Dashboard", templates[0].TemplateBase.DisplayName)
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
