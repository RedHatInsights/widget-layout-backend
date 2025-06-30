package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// handleServiceError is a generic error handler to reduce repeated error handling code in service methods.
func handleServiceError[T any](err error, notFoundMsg, generalMsg string, notFoundStatus int, notFoundReturn, generalReturn T) (T, int, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(notFoundMsg)
		return notFoundReturn, notFoundStatus, err
	}
	if err != nil {
		logrus.Errorf(generalMsg, err)
		return generalReturn, http.StatusInternalServerError, err
	}
	return *new(T), 0, nil
}

func GetTemplateByID(templateID int64, id identity.XRHID) (api.DashboardTemplate, int, error) {
	var template api.DashboardTemplate
	err := database.DB.Where(api.DashboardTemplate{ID: uint(templateID), UserId: id.Identity.User.UserID}).First(&template).Error
	if ret, status, err := handleServiceError(
		err,
		// notFoundMsg
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		// generalMsg
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		api.DashboardTemplate{}, api.DashboardTemplate{},
	); err != nil {
		return ret, status, err
	}
	return template, http.StatusOK, nil
}

func GetUserTemplates(id identity.XRHID) ([]api.DashboardTemplate, int, error) {
	var templates []api.DashboardTemplate
	err := database.DB.Where(api.DashboardTemplate{UserId: id.Identity.User.UserID}).Find(&templates).Error
	if _, status, err := handleServiceError(
		err,
		fmt.Sprintf("No dashboard templates found for user %s", id.Identity.User.UserID),
		"Failed to retrieve dashboard templates for user %s: %v", http.StatusNotFound,
		nil, []api.DashboardTemplate{},
	); err != nil {
		return nil, status, err
	}
	return templates, http.StatusOK, nil
}

func UpdateDashboardTemplate(templateID int64, newConfig api.DashboardTemplateConfig, id identity.XRHID) (api.DashboardTemplate, int, error) {
	var originalTemplate api.DashboardTemplate
	err := database.DB.First(&originalTemplate, templateID).Error
	if ret, status, err := handleServiceError(
		err,
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		api.DashboardTemplate{}, api.DashboardTemplate{},
	); err != nil {
		return ret, status, err
	}
	if !originalTemplate.IsAuthorized(id.Identity.User.UserID) {
		return api.DashboardTemplate{}, http.StatusForbidden, errors.New("unauthorized")
	}
	logrus.Infof("Updating dashboard template with ID: %d", templateID)
	originalTemplate.TemplateConfig = newConfig
	err = database.DB.Save(&originalTemplate).Error
	if err != nil {
		logrus.Errorf("Failed to update dashboard template with ID %d: %v", templateID, err)
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	return originalTemplate, http.StatusOK, nil
}

func DeleteDashboardTemplate(templateID int64, id identity.XRHID) (int, error) {
	var template api.DashboardTemplate
	err := database.DB.First(&template, templateID).Error
	if _, status, err := handleServiceError(
		err,
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		0, 0,
	); err != nil {
		return status, err
	}
	if !template.IsAuthorized(id.Identity.User.UserID) {
		return http.StatusForbidden, errors.New("unauthorized")
	}
	logrus.Infof("Deleting dashboard template with ID: %d", templateID)
	// delete permanently, there is no restore feature implemented
	err = database.DB.Unscoped().Delete(&template).Error
	if err != nil {
		logrus.Errorf("Failed to delete dashboard template with ID %d: %v", templateID, err)
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

func CopyDashboardTemplate(templateID int64, id identity.XRHID) (api.DashboardTemplate, int, error) {
	var dashboardTemplate api.DashboardTemplate
	err := database.DB.First(&dashboardTemplate, templateID).Error
	if ret, status, err := handleServiceError(
		err,
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		api.DashboardTemplate{}, api.DashboardTemplate{},
	); err != nil {
		return ret, status, err
	}
	// For copying, we don't check authorization - any user can copy any template
	// The new template will belong to the copying user
	newTemplate := api.DashboardTemplate{
		TemplateBase:   dashboardTemplate.TemplateBase,
		UserId:         id.Identity.User.UserID,
		TemplateConfig: dashboardTemplate.TemplateConfig,
	}
	err = database.DB.Create(&newTemplate).Error
	if err != nil {
		logrus.Errorf("Failed to create dashboard template: %v", err)
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	return newTemplate, http.StatusOK, nil
}

func ChangeDefaultTemplate(templateID int64, id identity.XRHID) (api.DashboardTemplate, int, error) {
	var template api.DashboardTemplate
	err := database.DB.First(&template, templateID).Error
	if ret, status, err := handleServiceError(
		err,
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		api.DashboardTemplate{}, api.DashboardTemplate{},
	); err != nil {
		return ret, status, err
	}
	if !template.IsAuthorized(id.Identity.User.UserID) {
		logrus.Errorf("User %s is not authorized to change default template with ID %d", id.Identity.User.UserID, templateID)
		return api.DashboardTemplate{}, http.StatusForbidden, errors.New("unauthorized")
	}
	tx := database.DB.Begin()
	// Update all other templates with the same base to unset their default status
	err = tx.Where(&api.DashboardTemplate{TemplateBase: api.DashboardTemplateBase{
		Name: template.TemplateBase.Name,
	}}).Updates(api.DashboardTemplate{Default: false}).Error
	if err != nil {
		logrus.Errorf("Failed to unset default dashboard template with ID %d: %v", templateID, err)
		tx.Rollback()
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	// Set the specified template as the default
	template.Default = true
	err = tx.Save(&template).Error
	if err != nil {
		logrus.Errorf("Failed to change default dashboard template with ID %d: %v", templateID, err)
		tx.Rollback()
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	err = tx.Commit().Error
	if err != nil {
		logrus.Errorf("Failed to commit transaction for changing default dashboard template with ID %d: %v", templateID, err)
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	return template, http.StatusOK, nil
}

func ResetDashboardTemplate(templateID int64, id identity.XRHID) (api.DashboardTemplate, int, error) {
	var template api.DashboardTemplate
	err := database.DB.First(&template, templateID).Error
	if ret, status, err := handleServiceError(
		err,
		fmt.Sprintf("Dashboard template with ID %d not found", templateID),
		"Failed to retrieve dashboard template with ID %d: %v", http.StatusNotFound,
		api.DashboardTemplate{}, api.DashboardTemplate{},
	); err != nil {
		return ret, status, err
	}
	if !template.IsAuthorized(id.Identity.User.UserID) {
		return api.DashboardTemplate{}, http.StatusForbidden, errors.New("unauthorized")
	}
	templateName := template.TemplateBase.Name
	baseTC, exists := BaseTemplateRegistry.GetBase(templateName)
	if !exists {
		logrus.Errorf("Base template %s not found for resetting dashboard template with ID %d", templateName, templateID)
		return template, http.StatusNotFound, fmt.Errorf("base template %s not found", templateName)
	}

	template.TemplateConfig = baseTC.TemplateConfig
	err = database.DB.Save(&template).Error
	if err != nil {
		logrus.Errorf("Failed to reset dashboard template with ID %d: %v", templateID, err)
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}
	logrus.Infof("Dashboard template with ID %d reset to base template %s", templateID, templateName)
	return template, http.StatusOK, nil
}

func ForkBaseTemplate(baseTemplateName string, id identity.XRHID) (api.DashboardTemplate, int, error) {
	baseTemplate, exists := BaseTemplateRegistry.GetBase(baseTemplateName)
	if !exists {
		logrus.Errorf("Base template %s not found for forking", baseTemplateName)
		return api.DashboardTemplate{}, http.StatusNotFound, fmt.Errorf("base template %s not found", baseTemplateName)
	}
	// Create a new dashboard template using the base template's ToDashboardTemplate method
	newTemplate := baseTemplate.ToDashboardTemplate()
	// Set the user ID for the forked template
	newTemplate.UserId = id.Identity.User.UserID

	err := database.DB.Create(&newTemplate).Error
	if err != nil {
		logrus.Errorf("Failed to create forked dashboard template from base %s: %v", baseTemplateName, err)
		return api.DashboardTemplate{}, http.StatusInternalServerError, err
	}

	logrus.Infof("Successfully forked base template %s to dashboard template with ID %d for user %s", baseTemplateName, newTemplate.ID, id.Identity.User.UserID)
	return newTemplate, http.StatusOK, nil
}
