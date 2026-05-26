package zapLogger

import (
	"context"

	"github.com/getsentry/sentry-go/zap"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init(cfg configs.LoggerConfig, sentryCfg configs.SentryConfig) {
	var err error
	if cfg.Mode == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil || logger == nil {
		panic(err)
	}

	if sentryCfg.Dsn != "" {
		sentryCore := sentryzap.NewSentryCore(context.Background(), sentryzap.Option{
			Level: []zapcore.Level{
				zapcore.InfoLevel,
				zapcore.WarnLevel,
				zapcore.ErrorLevel,
			},
			AddCaller: true,
		})
		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, sentryCore)
		}))
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
