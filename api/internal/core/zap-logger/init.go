package zapLogger

import (
	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"go.uber.org/zap"
)

var logger *zap.Logger

func Init(cfg configs.LoggerConfig) {
	var err error
	if cfg.Mode == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil || logger == nil {
		panic(err)
	}
}

func S() *zap.SugaredLogger {
	if logger == nil {
		panic("logger not initialized")
	}

	return logger.Sugar()
}

func Sync() {
	if logger == nil {
		panic("logger not initialized")
	}
	logger.Sync()
}
