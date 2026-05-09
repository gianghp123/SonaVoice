package main

import (
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
	"github.com/gianghp123/SonaVoice/api/internal/database"
)

func main() {
	cfg := configs.Load()
	zapLogger.Init(cfg.Logger)
	_ = database.Init(cfg.Database)

	// r := setupRouter()
	// registerRoutes(r, db)
	// r.Run(":" + cfg.Server.Port)
}
