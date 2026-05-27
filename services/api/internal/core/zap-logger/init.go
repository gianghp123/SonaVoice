package zapLogger

import (
	"context"
	"os"

	sentryzap "github.com/getsentry/sentry-go/zap"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func Init(cfg configs.LoggerConfig, sentryCfg configs.SentryConfig) {
	encoderConfig := zap.NewProductionEncoderConfig()

	if cfg.Mode == "development" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// stdout: debug/info/warn
	lowPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level < zapcore.ErrorLevel
	})

	// stderr: error/fatal/panic
	highPriority := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			lowPriority,
		),
		zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stderr),
			highPriority,
		),
	)

	logger = zap.New(core, zap.AddCaller())

	if sentryCfg.Dsn != "" {
		sentryCore := sentryzap.NewSentryCore(
			context.Background(),
			sentryzap.Option{
				Level: []zapcore.Level{
					zapcore.ErrorLevel,
					zapcore.PanicLevel,
					zapcore.FatalLevel,
				},
				AddCaller: true,
			},
		)

		logger = logger.WithOptions(
			zap.WrapCore(func(c zapcore.Core) zapcore.Core {
				return zapcore.NewTee(c, sentryCore)
			}),
		)
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
