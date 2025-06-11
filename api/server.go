package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// optional code omitted

type Server struct {
}

func NewServer(r *chi.Mux, middlewares ...func(next http.Handler) http.Handler) *Server {
	for _, mw := range middlewares {
		r.Use(mw)
	}
	server := &Server{}
	return server
}

// (GET /)
func (Server) GetWidgetLayout(w http.ResponseWriter, r *http.Request) {
	id := "id"
	name := "name"
	position := 1
	widget := Widget{
		Id:       &id,
		Name:     &name,
		Position: &position,
	}
	resp := WidgetList{widget}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
