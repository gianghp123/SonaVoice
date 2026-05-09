package database

import (
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(cfg configs.DatabaseConfig) *gorm.DB {
	dsn := cfg.DatabaseUrl
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	zapLogger.S().Info("Connected to database")
	return db
}
