package config

import (
	"os"

	"github.com/joho/godotenv"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

type WidgetLayoutConfig struct {
	LogLevel    string
	WebPort     int
	MetricsPort int
}

var config *WidgetLayoutConfig

func init() {
	godotenv.Load()
	config = &WidgetLayoutConfig{}
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "warn"
	}

	config.LogLevel = level
	if clowder.IsClowderEnabled() {
		cfg := clowder.LoadedConfig
		config.WebPort = *cfg.PublicPort
		config.MetricsPort = cfg.MetricsPort
	} else {
		config.WebPort = 8000
		config.MetricsPort = 9000
	}
}

func GetConfig() *WidgetLayoutConfig {
	return config
}
