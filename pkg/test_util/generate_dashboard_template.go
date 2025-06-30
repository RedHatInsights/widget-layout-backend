package test_util

import (
	"github.com/RedHatInsights/widget-layout-backend/api"
	"gorm.io/datatypes"
)

func MockDashboardTemplateFromTemplate(template api.DashboardTemplate) api.DashboardTemplate {
	// This function generates a new DashboardTemplate based on the provided template.
	// It can be used to create a new template with the same structure but different data.
	newTemplate := api.DashboardTemplate{
		ID:             template.ID,
		TemplateConfig: template.TemplateConfig,
		UserId:         template.UserId,
	}

	return newTemplate
}

func MockDashboardTemplate() api.DashboardTemplate {
	items := datatypes.NewJSONType([]api.WidgetItem{
		{
			Height:     2,
			Width:      2,
			X:          0,
			WidgetType: "widget1",
			Y:          0,
			Static:     false,
			Title:      "Sample Widget",
			MaxHeight:  4,
			MinHeight:  1,
		},
	})
	// This function generates a mock DashboardTemplate with predefined values.
	tm := api.DashboardTemplateConfig{
		Lg: items,
		Xl: items,
		Md: items,
		Sm: items,
	}

	return MockDashboardTemplateFromTemplate(api.DashboardTemplate{
		ID: GetUniqueID(), // Use unique ID instead of hardcoded
		TemplateBase: api.DashboardTemplateBase{
			Name:        "mock-template",
			DisplayName: "Mock Template Display",
		},
		TemplateConfig: tm,
		UserId:         "user-123",
	})
}

// MockDashboardTemplateWithUniqueID creates a mock dashboard template with a unique ID
func MockDashboardTemplateWithUniqueID() api.DashboardTemplate {
	template := MockDashboardTemplate()
	// The MockDashboardTemplate() already uses GetUniqueID() internally, so this is redundant
	// but we keep this function for clarity and future-proofing
	return template
}

// MockDashboardTemplateWithSpecificUser creates a mock dashboard template with a specific user ID
func MockDashboardTemplateWithSpecificUser(userID string) api.DashboardTemplate {
	template := MockDashboardTemplate()
	template.UserId = userID
	return template
}
