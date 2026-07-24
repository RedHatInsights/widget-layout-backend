package api_test

import (
	"testing"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func intPtr(i int) *int {
	return &i
}

func TestGridSizesIsValid(t *testing.T) {
	t.Run("should accept valid grid sizes", func(t *testing.T) {
		validSizes := []api.GridSizes{api.Sm, api.Md, api.Lg, api.Xl}
		for _, size := range validSizes {
			assert.NoError(t, size.IsValid(), "Grid size %s should be valid", size)
		}
	})

	t.Run("should reject invalid grid size", func(t *testing.T) {
		invalidSize := api.GridSizes("xxl")
		err := invalidSize.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid grid size")
		assert.Contains(t, err.Error(), "xxl")
	})

	t.Run("should reject empty grid size", func(t *testing.T) {
		emptySize := api.GridSizes("")
		err := emptySize.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid grid size")
	})
}

func TestGridSizesGetMaxWidth(t *testing.T) {
	t.Run("should return correct max width for each size", func(t *testing.T) {
		tests := []struct {
			size     api.GridSizes
			expected int
		}{
			{api.Sm, 1},
			{api.Md, 2},
			{api.Lg, 3},
			{api.Xl, 4},
		}

		for _, tt := range tests {
			maxWidth, err := tt.size.GetMaxWidth()
			assert.NoError(t, err, "GetMaxWidth should not error for %s", tt.size)
			assert.Equal(t, tt.expected, maxWidth, "Max width for %s", tt.size)
		}
	})

	t.Run("should return error for invalid grid size", func(t *testing.T) {
		invalidSize := api.GridSizes("invalid")
		maxWidth, err := invalidSize.GetMaxWidth()
		assert.Error(t, err)
		assert.Equal(t, 0, maxWidth)
	})
}

func TestDashboardTemplateIsValid(t *testing.T) {
	validConfig := api.DashboardTemplateConfig{
		Sm: datatypes.NewJSONType([]api.WidgetItem{}),
		Md: datatypes.NewJSONType([]api.WidgetItem{}),
		Lg: datatypes.NewJSONType([]api.WidgetItem{}),
		Xl: datatypes.NewJSONType([]api.WidgetItem{}),
	}

	t.Run("should accept valid template", func(t *testing.T) {
		dt := &api.DashboardTemplate{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "valid-template",
				DisplayName: "Valid Template",
			},
			TemplateConfig: validConfig,
		}
		assert.NoError(t, dt.IsValid())
	})

	t.Run("should reject template with empty name", func(t *testing.T) {
		dt := &api.DashboardTemplate{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "",
				DisplayName: "Valid Display",
			},
			TemplateConfig: validConfig,
		}
		err := dt.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("should reject template with empty displayName", func(t *testing.T) {
		dt := &api.DashboardTemplate{
			TemplateBase: api.DashboardTemplateBase{
				Name:        "valid-name",
				DisplayName: "",
			},
			TemplateConfig: validConfig,
		}
		err := dt.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "displayName is required")
	})
}

func TestDashboardTemplateConfigIsValid(t *testing.T) {
	t.Run("should accept valid config with widgets", func(t *testing.T) {
		config := &api.DashboardTemplateConfig{
			Sm: datatypes.NewJSONType([]api.WidgetItem{
				{Width: 1, Height: 2, WidgetType: "test-widget", X: intPtr(0), Y: intPtr(0), MaxHeight: intPtr(4), MinHeight: intPtr(1)},
			}),
			Md: datatypes.NewJSONType([]api.WidgetItem{}),
			Lg: datatypes.NewJSONType([]api.WidgetItem{}),
			Xl: datatypes.NewJSONType([]api.WidgetItem{}),
		}
		assert.NoError(t, config.IsValid())
	})

	t.Run("should accept valid config with empty widget arrays", func(t *testing.T) {
		config := &api.DashboardTemplateConfig{
			Sm: datatypes.NewJSONType([]api.WidgetItem{}),
			Md: datatypes.NewJSONType([]api.WidgetItem{}),
			Lg: datatypes.NewJSONType([]api.WidgetItem{}),
			Xl: datatypes.NewJSONType([]api.WidgetItem{}),
		}
		assert.NoError(t, config.IsValid())
	})

	t.Run("should reject config with invalid widget", func(t *testing.T) {
		config := &api.DashboardTemplateConfig{
			Sm: datatypes.NewJSONType([]api.WidgetItem{
				{Width: 1, Height: 0, WidgetType: "test-widget", X: intPtr(0), Y: intPtr(0)}, // height 0 invalid
			}),
			Md: datatypes.NewJSONType([]api.WidgetItem{}),
			Lg: datatypes.NewJSONType([]api.WidgetItem{}),
			Xl: datatypes.NewJSONType([]api.WidgetItem{}),
		}
		err := config.IsValid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "height must be at least 1")
	})
}

func TestWidgetItemIsValid(t *testing.T) {
	t.Run("should accept valid widget", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     2,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
			MaxHeight:  intPtr(4),
			MinHeight:  intPtr(1),
		}
		assert.NoError(t, wi.IsValid(api.Sm, 0))
	})

	t.Run("should reject empty widgetType", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     2,
			WidgetType: "",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "widgetType is required")
	})

	t.Run("should reject height less than 1", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     0,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "height must be at least 1")
	})

	t.Run("should reject width less than 1", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      0,
			Height:     2,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "width must be at least 1")
	})

	t.Run("should reject width exceeding max for grid size", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      2, // sm max is 1
			Height:     2,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "width 2 exceeds maximum 1")
	})

	t.Run("should accept max width for each grid size", func(t *testing.T) {
		tests := []struct {
			size     api.GridSizes
			maxWidth int
		}{
			{api.Sm, 1},
			{api.Md, 2},
			{api.Lg, 3},
			{api.Xl, 4},
		}

		for _, tt := range tests {
			wi := &api.WidgetItem{
				Width:      tt.maxWidth,
				Height:     1,
				WidgetType: "test-widget",
				X:          intPtr(0),
				Y:          intPtr(0),
			}
			assert.NoError(t, wi.IsValid(tt.size, 0), "Width %d should be valid for %s", tt.maxWidth, tt.size)
		}
	})

	t.Run("should reject height exceeding maxHeight", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     5,
			MaxHeight:  intPtr(4),
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "height 5 exceeds maxHeight 4")
	})

	t.Run("should reject height less than minHeight", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			MinHeight:  intPtr(2),
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "height 1 is less than minHeight 2")
	})

	t.Run("should reject maxHeight less than 1", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			MaxHeight:  intPtr(0),
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "maxHeight must be at least 1")
	})

	t.Run("should reject minHeight less than 1", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			MinHeight:  intPtr(0),
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minHeight must be at least 1")
	})

	t.Run("should reject x position out of bounds", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			WidgetType: "test-widget",
			X:          intPtr(5), // sm max width is 1, x can be 0 or 1
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "x position")
		assert.Contains(t, err.Error(), "out of bounds")
	})

	t.Run("should reject negative x position", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			WidgetType: "test-widget",
			X:          intPtr(-1),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "x position")
	})

	t.Run("should reject negative y position", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(-1),
		}
		err := wi.IsValid(api.Sm, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "y position cannot be negative")
	})

	t.Run("should reject invalid grid size variant", func(t *testing.T) {
		wi := &api.WidgetItem{
			Width:      1,
			Height:     1,
			WidgetType: "test-widget",
			X:          intPtr(0),
			Y:          intPtr(0),
		}
		err := wi.IsValid(api.GridSizes("invalid"), 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid grid size")
	})
}

func TestIsAuthorized(t *testing.T) {
	t.Run("should authorize matching user", func(t *testing.T) {
		dt := api.DashboardTemplate{
			UserId: "user-abc",
		}
		assert.True(t, dt.IsAuthorized("user-abc"))
	})

	t.Run("should reject non-matching user", func(t *testing.T) {
		dt := api.DashboardTemplate{
			UserId: "user-abc",
		}
		assert.False(t, dt.IsAuthorized("user-xyz"))
	})

	t.Run("should reject empty user ID", func(t *testing.T) {
		dt := api.DashboardTemplate{
			UserId: "user-abc",
		}
		assert.False(t, dt.IsAuthorized(""))
	})
}
