package server

import (
	"encoding/json"
	"net/http"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
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
	id := middlewares.GetUserIdentity(r.Context())
	logrus.Debugln("GetWidgetLayout called with identity:", id.Identity.User.UserID)
	w.Header().Set("Content-Type", "application/json")
	i := "widget1"
	h := 2
	width := 2
	x := 0
	y := 0
	title := "Sample Widget"
	static := false
	minH := 1
	maxH := 4
	widget := api.WidgetItem{
		Height:     h,
		Width:      width,
		X:          x,
		WidgetType: i,
		Y:          y,
		Static:     static,
		Title:      title,
		MaxHeight:  maxH,
		MinHeight:  minH,
	}
	tm := datatypes.NewJSONType([]api.WidgetItem{widget})
	templateConfig := api.DashboardTemplateConfig{
		Lg: tm,
		Xl: tm,
		Md: tm,
		Sm: tm,
	}

	dashboardTemplate := api.DashboardTemplate{
		ID:             123456,
		TemplateConfig: templateConfig,
	}

	resp := api.DashboardTemplateList{
		dashboardTemplate,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
