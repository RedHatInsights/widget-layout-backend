package models

import (
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestDashboardTemplateModel(t *testing.T) {
	t.Run("should create a dashboard template entity in database", func(t *testing.T) {
		genericTemplate := datatypes.NewJSONType([]api.WidgetItem{
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
		entity := DashboardTemplate{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "test-template",
				DisplayName: "Test Template",
			},
			TemplateConfig: api.DashboardTemplateConfig{
				Xl: genericTemplate,
				Lg: genericTemplate,
				Md: genericTemplate,
				Sm: genericTemplate,
			},
		}
		var emptyId uint = 0
		assert.NotNil(t, entity)
		assert.Equal(t, emptyId, entity.ID, "ID should be set after creation")
		err := database.DB.Create(&entity).Error
		assert.NoError(t, err)
		assert.NotEqual(t, emptyId, entity.ID, "ID should be set after creation")
		assert.NotEmpty(t, entity.TemplateConfig.Xl, "Xl config should not be empty")
		assert.NotEmpty(t, entity.TemplateConfig.Lg, "Lg config should not be empty")
		assert.NotEmpty(t, entity.TemplateConfig.Md, "Md config should not be empty")
		assert.NotEmpty(t, entity.TemplateConfig.Sm, "Sm config should not be empty")
		assert.NotEmpty(t, entity.CreatedAt, "CreatedAt should not be empty")
		assert.NotEmpty(t, entity.UpdatedAt, "UpdatedAt should not be empty")
		assert.Empty(t, entity.DeletedAt, "DeletedAt should be empty for new entity")
	})
}
