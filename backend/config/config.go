package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Auth     AuthConfig
}

type DatabaseConfig struct {
	DB_URL string
}

type AuthConfig struct {
	SecretKey        string
	RefreshSecretKey string
}

func InitConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Config{
		Database: DatabaseConfig{
			DB_URL: os.Getenv("DB_URL"),
		},
		Auth: AuthConfig{
			SecretKey:        os.Getenv("JWT_SECRET_KEY"),
			RefreshSecretKey: os.Getenv("JWT_REFRESH_SECRET_KEY"),
		},
	}, nil
}
