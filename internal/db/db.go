package db

import (
	"fmt"

	"github.com/codesentinel/codesentinel/internal/config"
	"github.com/codesentinel/codesentinel/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	var dialector gorm.Dialector

	if cfg.DBType == "postgres" && cfg.DSN != "" {
		dialector = postgres.Open(cfg.DSN)
	} else {
		dbPath := cfg.DBPath
		if dbPath == "" {
			dbPath = "codesentinel.db"
		}
		dialector = sqlite.Open(dbPath)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migration for all core models
	err = db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Repository{},
		&models.PullRequest{},
		&models.Commit{},
		&models.Review{},
		&models.Finding{},
		&models.Policy{},
		&models.CustomRule{},
		&models.Patch{},
		&models.Metric{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}
