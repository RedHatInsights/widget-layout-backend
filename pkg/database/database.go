package database

import (
	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	godotenv.Load()
	cfg := config.GetConfig()
	dns := cfg.DatabaseConfig.DBDNS
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}
	DB = db
}
