package domain

import (
	"os"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
)

func TestMain(m *testing.M) {
	zapLogger.Init(configs.LoggerConfig{Mode: "production"}, configs.SentryConfig{})
	os.Exit(m.Run())
}
