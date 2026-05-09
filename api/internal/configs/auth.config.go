package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type AuthConfig struct {
	JWTSecret        string
	JWTAccessExpiry  int
	JWTRefreshExpiry int
}

func loadAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:        utils.GetEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  utils.GetEnvInt("JWT_ACCESS_EXPIRY", 3600),
		JWTRefreshExpiry: utils.GetEnvInt("JWT_REFRESH_EXPIRY", 604800),
	}
}
