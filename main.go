package main

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config server.cfg.yaml spec/openapi.yaml

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/RedHatInsights/widget-layout-backend/api"
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/logger"
	"github.com/RedHatInsights/widget-layout-backend/pkg/server"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	chi "github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"github.com/oasdiff/yaml"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func loadSpec() (*openapi3.T, error) {
	specb, err := os.ReadFile("./spec/openapi.yaml")
	if err != nil {
		return nil, err
	}

	spec, err := openapi3.NewLoader().LoadFromData(specb)

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)

	}

	specJson, err := yaml.YAMLToJSON(specb)

	if err != nil {
		return nil, fmt.Errorf("failed to convert OpenAPI spec to JSON: %w", err)
	}

	err = os.WriteFile("./spec/openapi.json", specJson, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write OpenAPI spec to JSON file: %w", err)
	}
	return spec, err
}

func init() {
	godotenv.Load()
	database.InitDb()
	fmt.Println("Loading environment variables from .env file")
}

func main() {
	spec, err := loadSpec()
	cfg := config.GetConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load OpenAPI spec: %w", err))
	}

	spec.Servers = nil
	validatorMiddleware := middleware.OapiRequestValidatorWithOptions(spec, &middleware.Options{
		Options: openapi3filter.Options{},
	})

	r := chi.NewRouter()
	r.Use(
		chiMiddleware.RequestLogger(logger.NewLogger(logrus.New())))
	server := server.NewServer(r)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	apiPrefix := "/api/widget-layout/v1"

	api.HandlerWithOptions(server, api.ChiServerOptions{
		BaseURL:     apiPrefix,
		BaseRouter:  r,
		Middlewares: []api.MiddlewareFunc{},
	})

	r.Route(apiPrefix+"/", func(r chi.Router) {
		r.Use(validatorMiddleware)
	})

	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "/spec"))
	SpecServer(r, apiPrefix, filesDir)

	metricsRouter := chi.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())
	go func() {
		metricsStringAddr := fmt.Sprintf(":%s", strconv.Itoa(cfg.MetricsPort))
		if err := http.ListenAndServe(metricsStringAddr, metricsRouter); err != nil {
			log.Fatalf("Metrics server stopped %v", err)
		}
	}()

	err = http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(cfg.WebPort)), r)
	if err != nil {
		log.Fatalf("Widget layout backend has stopped due to %v", err)
	}
}

func SpecServer(r chi.Router, apiPrefix string, root http.FileSystem) {
	if strings.ContainsAny(apiPrefix, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if apiPrefix != "/" && apiPrefix[len(apiPrefix)-1] != '/' {
		r.Get(apiPrefix, http.RedirectHandler(apiPrefix+"/", 301).ServeHTTP)
		apiPrefix += "/"
	}
	apiPrefix += "*"

	r.Get(apiPrefix, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
