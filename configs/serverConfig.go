package configs

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
	Host string
	Env  string
}

func GetServerConfig() *ServerConfig {
	err := godotenv.Load()

	if err != nil {
		slog.Warn(
			"Error loading .env file, using system environment variables",
		)
	}

	serverConfig := &ServerConfig{
		Port: os.Getenv("PORT"),
		Host: os.Getenv("HOST"),
		Env:  os.Getenv("ENV"),
	}

	return serverConfig
}
