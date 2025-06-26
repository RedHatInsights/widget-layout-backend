package server

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/middlewares"
	"github.com/RedHatInsights/widget-layout-backend/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// optional code omitted

type Server struct{}

func NewServer(r chi.Router, middlewares ...func(next http.Handler) http.Handler) *Server {
	for _, mw := range middlewares {
		r.Use(mw)
	}
	server := &Server{}
	return server
}

// (GET /)
func (Server) GetWidgetLayout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := middlewares.GetUserIdentity(r.Context())
	resp, status, err := service.GetUserTemplates(id)
	if err != nil {
		logrus.Errorf("Failed to get dashboard templates: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (GET /{dashboardTemplateId})
func (Server) GetWidgetLayoutById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	w.Header().Set("Content-Type", "application/json")
	id := middlewares.GetUserIdentity(r.Context())
	resp, status, err := service.GetTemplateByID(dashboardTemplateId, id)

	if err != nil {
		logrus.Errorf("Failed to get dashboard template: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// (PATCH /{dashboardTemplateId})

func (Server) UpdateWidgetLayoutById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	var template api.DashboardTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	id := middlewares.GetUserIdentity(r.Context())
	dr, status, err := service.UpdateDashboardTemplate(
		dashboardTemplateId,
		template.TemplateConfig,
		id,
	)

	if err != nil {
		logrus.Errorf("Failed to update dashboard template: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(dr)
	if err != nil {
		logrus.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (Server) DeleteWidgetLayoutById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	id := middlewares.GetUserIdentity(r.Context())
	status, err := service.DeleteDashboardTemplate(
		dashboardTemplateId,
		id,
	)
	if err != nil {
		logrus.Errorf("Failed to delete dashboard template: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}
	w.WriteHeader(status)
	w.Write(nil)
}

func (Server) CopyWidgetLayoutById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	id := middlewares.GetUserIdentity(r.Context())
	w.Header().Set("Content-Type", "application/json")
	resp, status, err := service.CopyDashboardTemplate(dashboardTemplateId, id)
	if err != nil {
		logrus.Errorf("Failed to copy dashboard template: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (Server) SetWidgetLayoutDefaultById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	id := middlewares.GetUserIdentity(r.Context())
	w.Header().Set("Content-Type", "application/json")
	resp, status, err := service.ChangeDefaultTemplate(dashboardTemplateId, id)
	if err != nil {
		logrus.Errorf("Failed to change default dashboard template: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}
	w.WriteHeader(status)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (Server) ResetWidgetLayoutById(w http.ResponseWriter, r *http.Request, dashboardTemplateId int64) {
	id := middlewares.GetUserIdentity(r.Context())
	w.Header().Set("Content-Type", "application/json")
	// TODO: Implement reset functionality
	// Reset functionality not yet implemented, will require collection of base templates to exist
	resp, status, err := service.GetTemplateByID(dashboardTemplateId, id)
	if err != nil {
		logrus.Errorf("Failed to get dashboard template for reset: %v", err)
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    status,
				Message: err.Error(),
			},
		}})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
