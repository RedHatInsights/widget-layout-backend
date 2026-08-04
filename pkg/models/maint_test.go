package models

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/RedHatInsights/widget-layout-backend/pkg/config"
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
)

func TestMain(m *testing.M) {
	cfg := config.GetConfig()
	now := time.Now().UnixNano()
	dbName := fmt.Sprintf("%d-dashboard-template.db", now)
	cfg.TestMode = true
	cfg.DatabaseConfig.DBName = dbName

	database.InitDb()
	if err := database.DB.AutoMigrate(
		&DashboardTemplate{},
	); err != nil {
		log.Fatalf("Failed to migrate test database: %v", err)
	}

	exitCode := m.Run()

	err := os.Remove(dbName)

	if err != nil {
		fmt.Printf("Error removing test database file %s: %v\n", dbName, err)
	}

	os.Exit(exitCode)
}
