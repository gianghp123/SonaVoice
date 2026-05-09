package configs

import (
	"log"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	ClerkAuth ClerkConfig
	Logger    LoggerConfig
}

func Load() *Config {

	err := dotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		Server:    loadServerConfig(),
		Database:  loadDatabaseConfig(),
		ClerkAuth: loadClerkConfig(),
		Logger:    loadLoggerConfig(),
	}
}
