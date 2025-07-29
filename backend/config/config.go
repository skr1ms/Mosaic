package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	ServerConfig    ServerConfig
	PostgresConfig  PostgresConfig
	RedisConfig     RedisConfig
	AuthConfig      AuthConfig
	SMTPConfig      SMTPConfig
	RecaptchaConfig RecaptchaConfig
}

type ServerConfig struct {
	Port        string
	FrontendURL string
}

type PostgresConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type AuthConfig struct {
	AccessTokenSecret     string
	RefreshTokenSecret    string
	PasswordResetTokenTTL time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	SSL      bool
}

type RecaptchaConfig struct {
	SecretKey   string
	SiteKey     string
	Environment string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	return &Config{
		ServerConfig: ServerConfig{
			Port:        os.Getenv("SERVER_PORT"),
			FrontendURL: os.Getenv("FRONTEND_URL"),
		},
		PostgresConfig: PostgresConfig{
			URL: os.Getenv("DATABASE_URL"),
		},
		RedisConfig: RedisConfig{
			URL: os.Getenv("REDIS_URL"),
		},
		AuthConfig: AuthConfig{
			AccessTokenSecret:     os.Getenv("ACCESS_TOKEN_SECRET"),
			RefreshTokenSecret:    os.Getenv("REFRESH_TOKEN_SECRET"),
			PasswordResetTokenTTL: 1 * time.Hour,
		},
		SMTPConfig: SMTPConfig{
			Host:     os.Getenv("SMTP_HOST"),
			Port:     getEnvAsInt("SMTP_PORT", 465),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			SSL:      getEnvAsBool("SMTP_SSL", true),
		},
		RecaptchaConfig: RecaptchaConfig{
			SecretKey:   os.Getenv("RECAPTCHA_SECRET_KEY"),
			SiteKey:     os.Getenv("RECAPTCHA_SITE_KEY"),
			Environment: os.Getenv("ENVIRONMENT"),
		},
	}, nil
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
