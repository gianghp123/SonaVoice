package database

import (
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"github.com/gianghp123/SonaVoice/api/internal/core/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(cfg configs.DatabaseConfig) *gorm.DB {
	dsn := cfg.DatabaseUrl
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	logger.S().Info("Connected to database")
	return db
}
