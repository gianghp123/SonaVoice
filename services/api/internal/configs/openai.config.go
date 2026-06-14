package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type OpenAIConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

func loadOpenAIConfig() OpenAIConfig {
	return OpenAIConfig{
		APIKey:  utils.GetEnv("OPENAI_API_KEY", ""),
		BaseURL: utils.GetEnv("OPENAI_BASE_URL", "https://opencode.ai/zen/go/v1"),
		Model:   utils.GetEnv("OPENAI_MODEL", "deepseek-v4-flash"),
	}
}
