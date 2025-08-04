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
func (Server) GetWidgetLayout(w http.ResponseWriter, r *http.Request, params api.GetWidgetLayoutParams) {
	w.Header().Set("Content-Type", "application/json")
	id := middlewares.GetUserIdentity(r.Context())

	resp, status, err := service.GetUserTemplates(id, params)
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

	// Create the new list response format
	listResponse := api.DashboardTemplateListResponse{
		Data: resp,
		Meta: api.ListResponseMeta{
			Count: len(resp),
		},
	}

	// Use the status returned by the service (could be 200 or 404 when auto-creating)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(listResponse)
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
	resp, status, err := service.ResetDashboardTemplate(dashboardTemplateId, id)
	if err != nil {
		logrus.Errorf("Failed to reset dashboard template: %v", err)
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

func (Server) GetBaseWidgetDashboardTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	templateMap := service.BaseTemplateRegistry.GetAllBases()

	// Convert map to array to match API spec
	templates := make([]api.BaseWidgetDashboardTemplate, 0, len(templateMap))
	for _, template := range templateMap {
		templates = append(templates, template)
	}

	// Create the new list response format
	listResponse := api.BaseWidgetDashboardTemplateListResponse{
		Data: templates,
		Meta: api.ListResponseMeta{
			Count: len(templates),
		},
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(listResponse)
	if err != nil {
		logrus.Errorf("Failed to encode base widget dashboard templates: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (Server) GetBaseWidgetDashboardTemplateByName(w http.ResponseWriter, r *http.Request, baseTemplateName string) {
	w.Header().Set("Content-Type", "application/json")
	template, exists := service.BaseTemplateRegistry.GetBase(baseTemplateName)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(api.ErrorResponse{Errors: []api.ErrorPayload{
			{
				Code:    http.StatusNotFound,
				Message: "Base template not found",
			},
		}})
		return
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(template)
	if err != nil {
		logrus.Errorf("Failed to encode base widget dashboard template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (Server) ForkBaseWidgetDashboardTemplateByName(w http.ResponseWriter, r *http.Request, baseTemplateName string) {
	id := middlewares.GetUserIdentity(r.Context())
	w.Header().Set("Content-Type", "application/json")
	resp, status, err := service.ForkBaseTemplate(baseTemplateName, id)
	if err != nil {
		logrus.Errorf("Failed to fork base widget dashboard template: %v", err)
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

func (Server) GetWidgetMapping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	mappings := service.GetWidgetMappings()
	resp := api.WidgetMappingResponse{
		Data: mappings,
	}
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logrus.Errorf("Failed to encode widget mappings: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
