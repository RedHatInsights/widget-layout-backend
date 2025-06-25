package database

import (
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDb() {
	var dialector gorm.Dialector
	godotenv.Load()
	cfg := config.GetConfig()
	if cfg.TestMode {
		dialector = sqlite.Open(cfg.DatabaseConfig.DBName)
	} else {
		dns := cfg.DatabaseConfig.DBDNS
		dialector = postgres.Open(dns)
	}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	DB = db
}
