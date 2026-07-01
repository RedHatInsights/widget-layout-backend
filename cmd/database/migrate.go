package main

import (
	"github.com/RedHatInsights/widget-layout-backend/pkg/database"
	"github.com/RedHatInsights/widget-layout-backend/pkg/models"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	godotenv.Load()
	database.InitDb()
	// migrate models
	tx := database.DB.Begin().Session(&gorm.Session{
		Logger: logger.Default.LogMode(logger.Info),
	})
	// rollback changes
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorln("Migration failed:", r)
			tx.Rollback()
		}
	}()

	// Rename "default" column to "is_default" to avoid SQL reserved keyword
	// that causes silent 0-row updates in PostgreSQL.
	// Check if the old column exists before renaming to make this idempotent.
	if tx.Migrator().HasColumn(&models.DashboardTemplate{}, "default") {
		if err := tx.Migrator().RenameColumn(&models.DashboardTemplate{}, "default", "is_default"); err != nil {
			logrus.Errorln("Failed to rename 'default' column to 'is_default':", err.Error())
			tx.Rollback()
			return
		}
		logrus.Infoln("Renamed column 'default' to 'is_default'")
	}

	// migrate the base models
	if err := tx.AutoMigrate(
		&models.DashboardTemplate{},
	); err != nil {
		logrus.Errorln("Failed to migrate models:", err.Error())
		tx.Rollback()
		return
	}

	// Populate dashboardName for existing records that don't have it
	result := tx.Model(&models.DashboardTemplate{}).
		Where("dashboard_name = ? OR dashboard_name IS NULL", "").
		Update("dashboard_name", gorm.Expr("display_name"))

	if result.Error != nil {
		logrus.Errorln("Failed to populate dashboardName:", result.Error.Error())
		tx.Rollback()
		return
	}

	err := tx.Commit().Error
	if err != nil {
		logrus.Errorln("Failed to commit migration transaction:", err.Error())
		return
	}
	logrus.Infoln("Database migration completed successfully")
}
