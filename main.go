package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config server.cfg.yaml openapi.yaml

import (
	"fmt"
	"log"
	"net/http"
	"os"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/logger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/sirupsen/logrus"
)

func loadSpec() (*openapi3.T, error) {
	specb, err := os.ReadFile("./openapi.yaml")
	if err != nil {
		return nil, err
	}

	spec, err := openapi3.NewLoader().LoadFromData(specb)

	return spec, err
}

func main() {
	spec, err := loadSpec()
	if err != nil {
		panic(fmt.Errorf("failed to load OpenAPI spec: %w", err))
	}

	spec.Servers = nil
	mw := middleware.OapiRequestValidatorWithOptions(spec, &middleware.Options{
		Options: openapi3filter.Options{},
	})

	r := chi.NewMux()
	r.Use(chiMiddleware.RequestLogger(logger.NewLogger(logrus.New())))
	server := api.NewServer(r, mw)

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	s := &http.Server{
		Handler: h,
		Addr:    "0.0.0.0:8000",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
