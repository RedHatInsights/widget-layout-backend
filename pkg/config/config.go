package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
)

type DatabaseConfig struct {
	DBHost        string
	DBUser        string
	DBPassword    string
	DBPort        int
	DBName        string
	DBSSLMode     string
	DBSSLRootCert string
	DBDNS         string
}

type WidgetLayoutConfig struct {
	LogLevel       string
	WebPort        int
	MetricsPort    int
	DatabaseConfig DatabaseConfig
	TestMode       bool
}

var config *WidgetLayoutConfig

// The location of certificates is dictated by Clowder
const RdsCaLocation = "/app/rdsca.cert"

func (c *WidgetLayoutConfig) getCert(cfg *clowder.AppConfig) string {
	cert := ""
	if cfg.Database.SslMode != "verify-full" {
		return cert
	}
	if cfg.Database.RdsCa != nil {
		err := os.WriteFile(RdsCaLocation, []byte(*cfg.Database.RdsCa), 0644)
		if err != nil {
			panic(err)
		}
		cert = RdsCaLocation
	}
	return cert
}

type IdentityContextKeyType string

const (
	IdentityContextKey IdentityContextKeyType = "identity"
)

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
		config.DatabaseConfig = DatabaseConfig{
			DBHost:        cfg.Database.Hostname,
			DBUser:        cfg.Database.Username,
			DBPassword:    cfg.Database.Password,
			DBPort:        cfg.Database.Port,
			DBName:        cfg.Database.Name,
			DBSSLMode:     cfg.Database.SslMode,
			DBSSLRootCert: config.getCert(cfg),
		}
		config.DatabaseConfig.DBDNS = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", cfg.Database.Hostname, cfg.Database.Username, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port, cfg.Database.SslMode)
	} else {
		config.WebPort = 8000
		config.MetricsPort = 9000

		// Make sure the .env file is loaded and has correct values!
		config.DatabaseConfig.DBUser = os.Getenv("PGSQL_USER")
		config.DatabaseConfig.DBPassword = os.Getenv("PGSQL_PASSWORD")
		config.DatabaseConfig.DBHost = os.Getenv("PGSQL_HOSTNAME")
		port, _ := strconv.Atoi(os.Getenv("PGSQL_PORT"))
		config.DatabaseConfig.DBPort = port
		config.DatabaseConfig.DBName = os.Getenv("PGSQL_DATABASE")
		// Disable SSL mode for local development
		config.DatabaseConfig.DBSSLMode = "disable"
		config.DatabaseConfig.DBDNS = fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v", config.DatabaseConfig.DBHost, config.DatabaseConfig.DBUser, config.DatabaseConfig.DBPassword, config.DatabaseConfig.DBName, config.DatabaseConfig.DBPort, config.DatabaseConfig.DBSSLMode)
	}
}

func GetConfig() *WidgetLayoutConfig {
	return config
}
