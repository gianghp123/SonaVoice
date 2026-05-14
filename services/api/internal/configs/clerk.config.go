package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type ClerkConfig struct {
	ClerkSecret string
}

func loadClerkConfig() ClerkConfig {
	return ClerkConfig{
		ClerkSecret: utils.GetEnv("CLERK_SECRET_KEY", ""),
	}
}
