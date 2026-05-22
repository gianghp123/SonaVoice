package tests

import (
	"os"
	"testing"

	"github.com/gianghp123/SonaVoice/api/internal/configs"
	zapLogger "github.com/gianghp123/SonaVoice/api/internal/core/zap-logger"
)

func TestMain(m *testing.M) {
	zapLogger.Init(configs.LoggerConfig{Mode: "production"})
	os.Exit(m.Run())
}
