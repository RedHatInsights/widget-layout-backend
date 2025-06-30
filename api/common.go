package api

import (
	"encoding/json"
	"errors"

	"github.com/sirupsen/logrus"
)

func (t DashboardTemplate) IsAuthorized(userID string) bool {
	// This method checks if the user is authorized to access the template.
	// For this example, we will assume that the user is authorized if their ID matches the UserId in the template.
	if t.UserId != userID {
		logrus.Errorf("User %s is not authorized to access template with ID %d", userID, t.ID)
	}
	return t.UserId == userID
}

// We have to ensure the CX and CY attributes gets unmarshaled into x and y attributes
// There is an issue with the yaml parser that has the character "y" as a reserved character which translates into "true" value and it causes issues
// when unmarshaling the yaml file into the DashboardTemplateConfig struct.
func (wi *WidgetItem) UnmarshalJSON(data []byte) error {
	type alias WidgetItem

	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	tmp := struct {
		alias
		X  *int `json:"x"`
		Y  *int `json:"y"`
		CX *int `json:"cx"`
		CY *int `json:"cy"`
	}{}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	*wi = WidgetItem(tmp.alias)

	switch {
	case tmp.X != nil && tmp.Y != nil:
		wi.X, wi.Y = tmp.X, tmp.Y
	case tmp.CX != nil && tmp.CY != nil:
		wi.X, wi.Y = tmp.CX, tmp.CY
	default:
		return errors.New("WidgetItem must have either x/y or cx/cy attributes")
	}

	return nil
}
