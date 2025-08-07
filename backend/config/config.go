package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type Config struct {
	ServerConfig          ServerConfig
	PostgresConfig        PostgresConfig
	RedisConfig           RedisConfig
	AuthConfig            AuthConfig
	SMTPConfig            SMTPConfig
	RecaptchaConfig       RecaptchaConfig
	AlphaBankConfig       AlphaBankConfig
	MetricsConfig         MetricsConfig
	S3MinioConfig         S3MinioConfig
	StableDiffusionConfig StableDiffusionConfig
	BrandingConfig        BrandingConfig
}

type ServerConfig struct {
	Port              string
	FrontendURL       string
	PaymentSuccessURL string
	PaymentFailureURL string
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

type AlphaBankConfig struct {
	Url           string
	Username      string
	Password      string
	WebhookURL    string // URL для получения webhook уведомлений
	WebhookSecret string // Секретный ключ для валидации webhook'ов
}

type S3MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

type StableDiffusionConfig struct {
	BaseURL string
}

type MetricsConfig struct {
	Port string
}

type BrandingConfig struct {
	DefaultBrandName       string
	DefaultLogoURL         string
	DefaultContactEmail    string
	DefaultContactAddress  string
	DefaultContactPhone    string
	DefaultContactTelegram string
	DefaultContactWhatsapp string
	DefaultTelegramLink    string
	DefaultWhatsappLink    string
	DefaultOzonLink        string
	DefaultWildberriesLink string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	config := &Config{
		ServerConfig: ServerConfig{
			Port:              os.Getenv("SERVER_PORT"),
			FrontendURL:       os.Getenv("FRONTEND_URL"),
			PaymentSuccessURL: getEnvOrDefault("PAYMENT_SUCCESS_URL", "http://localhost:3000/payment/success"),
			PaymentFailureURL: getEnvOrDefault("PAYMENT_FAILURE_URL", "http://localhost:3000/payment/failure"),
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
		AlphaBankConfig: AlphaBankConfig{
			Url:           getAlphaBankUrl(),
			Username:      os.Getenv("ALFA_BANK_USERNAME"),
			Password:      os.Getenv("ALFA_BANK_PASSWORD"),
			WebhookURL:    os.Getenv("ALFA_BANK_WEBHOOK_URL"),
			WebhookSecret: os.Getenv("ALFA_BANK_WEBHOOK_SECRET"),
		},
		MetricsConfig: MetricsConfig{
			Port: os.Getenv("METRICS_PORT"),
		},
		S3MinioConfig: S3MinioConfig{
			Endpoint:        os.Getenv("MINIO_ENDPOINT"),
			AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY"),
			SecretAccessKey: os.Getenv("MINIO_SECRET_KEY"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
			BucketName:      getEnvOrDefault("MINIO_BUCKET", "mosaic-images"),
			Region:          getEnvOrDefault("MINIO_REGION", "us-east-1"),
		},
		StableDiffusionConfig: StableDiffusionConfig{
			BaseURL: os.Getenv("STABLE_DIFFUSION_URL"),
		},
		BrandingConfig: BrandingConfig{
			DefaultBrandName:       getEnvOrDefault("DEFAULT_BRAND_NAME", "Мозаика"),
			DefaultLogoURL:         getEnvOrDefault("DEFAULT_LOGO_URL", ""),
			DefaultContactEmail:    getEnvOrDefault("DEFAULT_CONTACT_EMAIL", "info@mosaic.ru"),
			DefaultContactAddress:  getEnvOrDefault("DEFAULT_CONTACT_ADDRESS", ""),
			DefaultContactPhone:    getEnvOrDefault("DEFAULT_CONTACT_PHONE", ""),
			DefaultContactTelegram: getEnvOrDefault("DEFAULT_CONTACT_TELEGRAM", ""),
			DefaultContactWhatsapp: getEnvOrDefault("DEFAULT_CONTACT_WHATSAPP", ""),
			DefaultTelegramLink:    getEnvOrDefault("DEFAULT_TELEGRAM_LINK", ""),
			DefaultWhatsappLink:    getEnvOrDefault("DEFAULT_WHATSAPP_LINK", ""),
			DefaultOzonLink:        getEnvOrDefault("DEFAULT_OZON_LINK", ""),
			DefaultWildberriesLink: getEnvOrDefault("DEFAULT_WILDBERRIES_LINK", ""),
		},
	}

	// Валидация обязательных переменных окружения
	if err := validateConfig(config); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
		return nil, err
	}

	return config, nil
}

// validateConfig проверяет обязательные поля конфигурации
func validateConfig(config *Config) error {
	var missingVars []string

	// Критически важные переменные для работы системы
	if config.ServerConfig.Port == "" {
		missingVars = append(missingVars, "SERVER_PORT")
	}
	if config.PostgresConfig.URL == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}
	if config.RedisConfig.URL == "" {
		missingVars = append(missingVars, "REDIS_URL")
	}
	if config.AuthConfig.AccessTokenSecret == "" {
		missingVars = append(missingVars, "ACCESS_TOKEN_SECRET")
	}
	if config.AuthConfig.RefreshTokenSecret == "" {
		missingVars = append(missingVars, "REFRESH_TOKEN_SECRET")
	}

	// Переменные для платежной системы (критично для безопасности)
	if config.AlphaBankConfig.Username == "" {
		missingVars = append(missingVars, "ALFA_BANK_USERNAME")
	}
	if config.AlphaBankConfig.Password == "" {
		missingVars = append(missingVars, "ALFA_BANK_PASSWORD")
	}
	if config.AlphaBankConfig.WebhookSecret == "" {
		missingVars = append(missingVars, "ALFA_BANK_WEBHOOK_SECRET")
	}

	// Переменные для S3 (важно для хранения изображений)
	if config.S3MinioConfig.Endpoint == "" {
		missingVars = append(missingVars, "MINIO_ENDPOINT")
	}
	if config.S3MinioConfig.AccessKeyID == "" {
		missingVars = append(missingVars, "MINIO_ACCESS_KEY")
	}
	if config.S3MinioConfig.SecretAccessKey == "" {
		missingVars = append(missingVars, "MINIO_SECRET_KEY")
	}

	// Переменные для SMTP (важно для отправки email)
	if config.SMTPConfig.Host == "" {
		missingVars = append(missingVars, "SMTP_HOST")
	}
	if config.SMTPConfig.Username == "" {
		missingVars = append(missingVars, "SMTP_USERNAME")
	}
	if config.SMTPConfig.Password == "" {
		missingVars = append(missingVars, "SMTP_PASSWORD")
	}
	if config.SMTPConfig.From == "" {
		missingVars = append(missingVars, "SMTP_FROM")
	}

	// Переменные для reCAPTCHA (важно для безопасности)
	if config.RecaptchaConfig.SecretKey == "" {
		missingVars = append(missingVars, "RECAPTCHA_SECRET_KEY")
	}

	// Переменные для Stable Diffusion (важно для основной функциональности)
	if config.StableDiffusionConfig.BaseURL == "" {
		missingVars = append(missingVars, "STABLE_DIFFUSION_URL")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}

func getAlphaBankUrl() string {
	environment := os.Getenv("ENVIRONMENT")
	if environment == "production" {
		return os.Getenv("ALFA_BANK_PROD_URL")
	}
	// Для тестовой среды
	testUrl := os.Getenv("ALFA_BANK_TEST_URL")
	if testUrl == "" {
		testUrl = "https://alfa.rbsuat.com" // URL тестовой среды по умолчанию
	}
	return testUrl
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

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
