package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config server.cfg.yaml openapi.yaml

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/logger"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func init() {
	godotenv.Load()
	fmt.Println("Loading environment variables from .env file")
}

func main() {
	spec, err := loadSpec()
	cfg := config.GetConfig()
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

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// get an `http.Handler` that we can use
	serverHandler := api.HandlerFromMux(server, r)

	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	go func() {
		metricsStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.MetricsPort))
		if err := http.ListenAndServe(metricsStringAddr, metricsRouter); err != nil {
			log.Fatalf("Metrics server stopped %v", err)
		}
	}()

	s := &http.Server{
		Handler: serverHandler,
		Addr:    fmt.Sprintf(":%s", strconv.Itoa(cfg.WebPort)),
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Widget layout backend has stopped due to %v", err)
	}
}
