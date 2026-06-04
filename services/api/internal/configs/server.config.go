package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type ServerConfig struct {
	Mode      string
	Port      string
	AllowUrls []string
}

func loadServerConfig() ServerConfig {
	feURL := utils.GetEnv("FE_URL", "")
	speechURL := utils.GetEnv("SPEECH_SERVICE_URL", "")

	var allowUrls []string
	if feURL != "" {
		allowUrls = append(allowUrls, feURL)
	}
	if speechURL != "" {
		allowUrls = append(allowUrls, speechURL)
	}

	return ServerConfig{
		Mode:      utils.GetEnv("MODE", "debug"),
		Port:      utils.GetEnv("PORT", "8080"),
		AllowUrls: allowUrls,
	}
}
