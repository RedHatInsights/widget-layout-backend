package api

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	minDimension  = 1
	maxCoordinate = 3 // for X Coordinate
	minCoordinate = 0
)

type GridSizes string

const (
	Sm GridSizes = "sm"
	Md GridSizes = "md"
	Lg GridSizes = "lg"
	Xl GridSizes = "xl"
)

func (gs GridSizes) IsValid() error {
	switch gs {
	case Sm, Md, Lg, Xl:
		return nil
	default:
		return fmt.Errorf("invalid grid size, expected one of %s, %s, %s, %s, got %s", Sm, Md, Lg, Xl, gs)
	}
}

func (gs GridSizes) GetMaxWidth() (int, error) {
	if err := gs.IsValid(); err != nil {
		return 0, err
	}

	switch gs {
	case Sm:
		return 1, nil
	case Md:
		return 2, nil
	case Lg:
		return 3, nil
	case Xl:
		return 4, nil
	default:
		return 0, errors.New("invalid grid size")
	}
}

// IsValid validates the entire dashboard template
func (dt *DashboardTemplate) IsValid() error {
	if dt.TemplateBase.Name == "" {
		return errors.New("template name is required")
	}

	if dt.TemplateBase.DisplayName == "" {
		return errors.New("template displayName is required")
	}

	if err := dt.TemplateConfig.IsValid(); err != nil {
		return err
	}

	return nil
}

// IsValid validates the template configuration
func (tc *DashboardTemplateConfig) IsValid() error {
	configs := reflect.ValueOf(*tc)
	typeOfS := configs.Type()

	for i := 0; i < configs.NumField(); i++ {
		field := configs.Field(i)
		layoutSize := typeOfS.Field(i).Tag.Get("json")
		gridSize := GridSizes(layoutSize)

		if err := gridSize.IsValid(); err != nil {
			return err
		}

		jsonType := field.Interface()

		dataMethod := reflect.ValueOf(jsonType).MethodByName("Data")
		if !dataMethod.IsValid() {
			return fmt.Errorf("field %s does not have Data() method", typeOfS.Field(i).Name)
		}

		results := dataMethod.Call(nil)
		if len(results) == 0 {
			return fmt.Errorf("Data() method returned no results for %s", layoutSize)
		}

		// Check if data is nil
		itemsValue := results[0]
		if itemsValue.IsNil() {
			return fmt.Errorf("grid size %s cannot be null", layoutSize)
		}

		items := itemsValue.Interface().([]WidgetItem)

		// Validate each widget
		for idx, widget := range items {
			// Initialize coordinates if needed
			if widget.X == nil {
				zero := 0
				widget.X = &zero
			}
			if widget.Y == nil {
				zero := 0
				widget.Y = &zero
			}

			// Validate the widget
			if err := widget.IsValid(gridSize, idx); err != nil {
				return err
			}
		}
	}

	return nil
}

// IsValid validates a single widget item
func (wi *WidgetItem) IsValid(variant GridSizes, index int) error {
	if err := variant.IsValid(); err != nil {
		return err
	}

	if wi.WidgetType == "" {
		return fmt.Errorf("widget[%d] in %s: widgetType is required", index, variant)
	}

	if wi.Height < minDimension {
		return fmt.Errorf("widget[%d] in %s: height must be at least 1", index, variant)
	}

	if wi.MaxHeight != nil && *wi.MaxHeight < minDimension {
		return fmt.Errorf("widget[%d] in %s: maxHeight must be at least 1", index, variant)
	}

	if wi.MinHeight != nil && *wi.MinHeight < minDimension {
		return fmt.Errorf("widget[%d] in %s: minHeight must be at least 1", index, variant)
	}

	if wi.MaxHeight != nil && wi.Height > *wi.MaxHeight {
		return fmt.Errorf("widget[%d] in %s: height %d exceeds maxHeight %d", index, variant, wi.Height, *wi.MaxHeight)
	}

	if wi.Width < minDimension {
		return fmt.Errorf("widget[%d] in %s: width must be at least 1", index, variant)
	}

	if wi.MinHeight != nil && wi.Height < *wi.MinHeight {
		return fmt.Errorf("widget[%d] in %s: height %d is less than minHeight %d", index, variant, wi.Height, *wi.MinHeight)
	}

	maxWidth, err := variant.GetMaxWidth()
	if err != nil {
		return err
	}
	if wi.Width > maxWidth {
		return fmt.Errorf("widget[%d] in %s: width %d exceeds maximum %d", index, variant, wi.Width, maxWidth)
	}

	if wi.X != nil && (*wi.X < minCoordinate || *wi.X > maxWidth) {
		return fmt.Errorf("widget[%d] in %s: x position %d is out of bounds", index, variant, *wi.X)
	}

	if wi.Y != nil && *wi.Y < minCoordinate {
		return fmt.Errorf("widget[%d] in %s: y position cannot be negative", index, variant)
	}

	return nil
}
