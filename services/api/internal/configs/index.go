package configs

import (
	"log"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	Database  DatabaseConfig
	ClerkAuth ClerkConfig
	Logger    LoggerConfig
	Redis     RedisConfig
	Sentry    SentryConfig
}

func Load() *Config {

	if err := dotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		Database:  loadDatabaseConfig(),
		ClerkAuth: loadClerkConfig(),
		Logger:    loadLoggerConfig(),
		Redis:     loadRedisConfig(),
		Sentry:    loadSentryConfig(),
	}
}
