package configs

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port                 string
	Host                 string
	Env                  string
	AppURL               string
	SecretKey            string
	GoogleClientID       string
	GoogleClientSecret   string
	GithubClientID       string
	GithubClientSecret   string
	GoogleLoginAPI       string
	GoogleCallbackAPI    string
	GithubLoginAPI       string
	GithubCallbackAPI    string
	LoginAPI             string
	SignupAPI            string
	LogoutAPI            string
	ProjectAPI           string
	UpdateProjectAPI     string
	StreamLogsAPI        string
	TriggerDeploymentAPI string
	ListDeploymentsAPI   string
	GetDeploymentAPI     string
	DeleteDeploymentAPI  string
	DbConnectionString   string
	ArtifactsDir         string
	RelayDomain          string
	ProxyPort            string
}

var serverConfig *ServerConfig

func InitServerConfig() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Error loading .env file, using system environment variables")
	}

	serverConfig = &ServerConfig{
		Port:                 os.Getenv("PORT"),
		Host:                 os.Getenv("HOST"),
		Env:                  os.Getenv("ENV"),
		AppURL:               os.Getenv("APP_URL"),
		SecretKey:            os.Getenv("JWT_SECRET"),
		GoogleClientID:       os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:   os.Getenv("GOOGLE_CLIENT_SECRET"),
		GithubClientID:       os.Getenv("GITHUB_CLIENT_ID"),
		GithubClientSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		GoogleLoginAPI:       os.Getenv("GOOGLE_LOGIN_API"),
		GoogleCallbackAPI:    os.Getenv("GOOGLE_CALLBACK_API"),
		GithubLoginAPI:       os.Getenv("GITHUB_LOGIN_API"),
		GithubCallbackAPI:    os.Getenv("GITHUB_CALLBACK_API"),
		LoginAPI:             os.Getenv("LOGIN_API"),
		SignupAPI:            os.Getenv("SIGNUP_API"),
		LogoutAPI:            os.Getenv("LOGOUT_API"),
		ProjectAPI:           os.Getenv("PROJECT_API"),
		UpdateProjectAPI:     os.Getenv("UPDATE_PROJECT_API"),
		StreamLogsAPI:        os.Getenv("STREAM_LOGS_API"),
		TriggerDeploymentAPI: os.Getenv("TRIGGER_DEPLOYMENT_API"),
		ListDeploymentsAPI:   os.Getenv("LIST_DEPLOYMENTS_API"),
		GetDeploymentAPI:     os.Getenv("GET_DEPLOYMENT_API"),
		DeleteDeploymentAPI:  os.Getenv("DELETE_DEPLOYMENT_API"),
		DbConnectionString:   os.Getenv("DATABASE_URL"),
		ArtifactsDir:         os.Getenv("ARTIFACTS_DIR"),
		RelayDomain:          os.Getenv("RELAY_DOMAIN"),
		ProxyPort:            os.Getenv("PROXY_PORT"),
	}
}

func GetServerConfig() *ServerConfig {
	if serverConfig == nil {
		InitServerConfig()
	}
	return serverConfig
}
