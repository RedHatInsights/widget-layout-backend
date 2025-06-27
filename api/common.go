package api

import "github.com/sirupsen/logrus"

func (t DashboardTemplate) IsAuthorized(userID string) bool {
	// This method checks if the user is authorized to access the template.
	// For this example, we will assume that the user is authorized if their ID matches the UserId in the template.
	if t.UserId != userID {
		logrus.Errorf("User %s is not authorized to access template with ID %d", userID, t.ID)
	}
	return t.UserId == userID
}
