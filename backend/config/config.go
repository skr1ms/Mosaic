package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	MosaicGeneratorConfig MosaicGeneratorConfig
	DefaultAdminConfig    DefaultAdminConfig
	DefaultPartnerConfig  DefaultPartnerConfig
	GitLabConfig          GitLabConfig
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
	WebhookURL    string
	WebhookSecret string
}

type S3MinioConfig struct {
	Endpoint          string
	AccessKeyID       string
	SecretAccessKey   string
	UseSSL            bool
	BucketName        string
	Region            string
	LogosBucketName   string
	ChatBucketName    string
	PreviewBucketName string
	PublicURL         string
}

type StableDiffusionConfig struct {
	BaseURL string
}

type MosaicGeneratorConfig struct {
	ScriptPath    string
	PalettePath   string
	OutputDir     string
	PythonCommand string
}

type MetricsConfig struct {
	Port string
}

type DefaultAdminConfig struct {
	DefaultLogin    string
	DefaultEmail    string
	DefaultPassword string
}

type DefaultPartnerConfig struct {
	DefaultBrandName       string
	DefaultPartnerCode     string
	DefaultDomain          string
	DefaultLogo            string
	DefaultLogin           string
	DefaultEmail           string
	DefaultPassword        string
	DefaultAddress         string
	DefaultPhone           string
	DefaultContactTelegram string
	DefaultWhatsapp        string
	DefaultTelegramLink    string
	DefaultWhatsappLink    string
	DefaultOzonLink        string
	DefaultWildberriesLink string
}

type GitLabConfig struct {
	BaseURL      string
	APIToken     string
	TriggerToken string
	ProjectID    string
}

func NewConfig() (*Config, error) {
	envPath := filepath.Join("..", ".env")
	err := godotenv.Load(envPath)
	if err != nil {
		log.Printf("Warning: .env file not found: %s", envPath)
	}

	config := &Config{
		ServerConfig: ServerConfig{
			Port:              "8080",
			FrontendURL:       os.Getenv("FRONTEND_URL"),
			PaymentSuccessURL: "http://localhost:3000/payment/success",
			PaymentFailureURL: "http://localhost:3000/payment/failure",
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
			Port:     getSMTPPort(),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			SSL:      getSMTPSSL(),
		},
		RecaptchaConfig: RecaptchaConfig{
			SecretKey:   os.Getenv("RECAPTCHA_SECRET_KEY"),
			SiteKey:     os.Getenv("RECAPTCHA_SITE_KEY"),
			Environment: os.Getenv("ENVIRONMENT"),
		},
		AlphaBankConfig: AlphaBankConfig{
			Url:           os.Getenv("ALFA_BANK_PROD_URL"),
			Username:      os.Getenv("ALFA_BANK_USERNAME"),
			Password:      os.Getenv("ALFA_BANK_PASSWORD"),
			WebhookURL:    os.Getenv("ALFA_BANK_WEBHOOK_URL"),
			WebhookSecret: os.Getenv("ALFA_BANK_WEBHOOK_SECRET"),
		},
		MetricsConfig: MetricsConfig{
			Port: "8091",
		},
		S3MinioConfig: S3MinioConfig{
			Endpoint:          os.Getenv("MINIO_ENDPOINT"),
			AccessKeyID:       os.Getenv("MINIO_ROOT_USER"),
			SecretAccessKey:   os.Getenv("MINIO_ROOT_PASSWORD"),
			UseSSL:            false,
			Region:            os.Getenv("MINIO_REGION"),
			BucketName:        os.Getenv("MINIO_IMAGE_BUCKET"),
			LogosBucketName:   os.Getenv("MINIO_LOGOS_BUCKET"),
			ChatBucketName:    os.Getenv("MINIO_CHAT_BUCKET"),
			PreviewBucketName: os.Getenv("MINIO_PREVIEW_BUCKET"),
			PublicURL:         os.Getenv("MINIO_PUBLIC_URL"),
		},
		StableDiffusionConfig: StableDiffusionConfig{
			BaseURL: os.Getenv("STABLE_DIFFUSION_URL"),
		},
		MosaicGeneratorConfig: MosaicGeneratorConfig{
			ScriptPath:    "/app/scripts/mosaic_cli.py",
			PalettePath:   "/app/scripts/",
			OutputDir:     "/tmp/mosaic_output/",
			PythonCommand: "python3",
		},
		DefaultAdminConfig: DefaultAdminConfig{
			DefaultLogin:    "admin",
			DefaultEmail:    os.Getenv("DEFAULT_ADMIN_EMAIL"),
			DefaultPassword: os.Getenv("DEFAULT_ADMIN_PASSWORD"),
		},
		DefaultPartnerConfig: DefaultPartnerConfig{
			DefaultBrandName:       "Живопись по номерам",
			DefaultPartnerCode:     "0000",
			DefaultDomain:          "photo.doyoupaint.com",
			DefaultLogo:            "",
			DefaultLogin:           "doyoupaint",
			DefaultEmail:           "info@mosaic.ru",
			DefaultPassword:        os.Getenv("DEFAULT_PASSWORD"),
			DefaultAddress:         "Псковская обл., г.Великие Луки, ул.Запрудная д.4в3",
			DefaultPhone:           "74951234567",
			DefaultContactTelegram: "@mosaic_support",
			DefaultWhatsapp:        "79991234567",
			DefaultTelegramLink:    "https://t.me/mosaic_support",
			DefaultWhatsappLink:    "https://wa.me/79991234567",
			DefaultOzonLink:        "https://ozon.ru/seller/mosaic",
			DefaultWildberriesLink: "https://wildberries.ru/seller/mosaic",
		},
		GitLabConfig: GitLabConfig{
			BaseURL:      os.Getenv("GITLAB_BASE_URL"),
			APIToken:     os.Getenv("GITLAB_API_TOKEN"),
			TriggerToken: os.Getenv("GITLAB_TRIGGER_TOKEN"),
			ProjectID:    os.Getenv("GITLAB_PROJECT_ID"),
		},
	}

	if err := validateConfig(config); err != nil {
		log.Fatal().Err(err).Msg("Configuration validation failed")
		return nil, err
	}

	return config, nil
}

func getSMTPPort() int {
	portStr := os.Getenv("SMTP_PORT")
	if portStr == "" {
		return 465 // default SMTP SSL port
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Warning: Invalid SMTP_PORT value '%s', using default 465", portStr)
		return 465
	}
	return port
}

func getSMTPSSL() bool {
	sslStr := os.Getenv("SMTP_SSL")
	if sslStr == "" {
		return true // default to SSL enabled
	}
	ssl, err := strconv.ParseBool(sslStr)
	if err != nil {
		log.Printf("Warning: Invalid SMTP_SSL value '%s', using default true", sslStr)
		return true
	}
	return ssl
}

func validateConfig(config *Config) error {
	var missingVars []string

	// JWT configuration validation
	if config.AuthConfig.AccessTokenSecret == "" {
		missingVars = append(missingVars, "ACCESS_TOKEN_SECRET")
	}
	if config.AuthConfig.RefreshTokenSecret == "" {
		missingVars = append(missingVars, "REFRESH_TOKEN_SECRET")
	}

	// AlphaBank acquiring configuration validation
	if config.AlphaBankConfig.Url == "" {
		missingVars = append(missingVars, "ALFA_BANK_PROD_URL")
	}
	if config.AlphaBankConfig.Username == "" {
		missingVars = append(missingVars, "ALFA_BANK_USERNAME")
	}
	if config.AlphaBankConfig.Password == "" {
		missingVars = append(missingVars, "ALFA_BANK_PASSWORD")
	}
	if config.AlphaBankConfig.WebhookSecret == "" {
		missingVars = append(missingVars, "ALFA_BANK_WEBHOOK_SECRET")
	}
	if config.AlphaBankConfig.WebhookURL == "" {
		missingVars = append(missingVars, "ALFA_BANK_WEBHOOK_URL")
	}

	// Recaptcha v2 configuration validation
	if config.RecaptchaConfig.SecretKey == "" {
		missingVars = append(missingVars, "RECAPTCHA_SECRET_KEY")
	}
	if config.RecaptchaConfig.SiteKey == "" {
		missingVars = append(missingVars, "RECAPTCHA_SITE_KEY")
	}

	// S3/MinIO configuration validation
	if config.S3MinioConfig.Endpoint == "" {
		missingVars = append(missingVars, "MINIO_ENDPOINT")
	}
	if config.S3MinioConfig.AccessKeyID == "" {
		missingVars = append(missingVars, "MINIO_ROOT_USER")
	}
	if config.S3MinioConfig.SecretAccessKey == "" {
		missingVars = append(missingVars, "MINIO_ROOT_PASSWORD")
	}
	if config.S3MinioConfig.BucketName == "" {
		missingVars = append(missingVars, "MINIO_IMAGE_BUCKET")
	}
	if config.S3MinioConfig.LogosBucketName == "" {
		missingVars = append(missingVars, "MINIO_LOGOS_BUCKET")
	}
	if config.S3MinioConfig.ChatBucketName == "" {
		missingVars = append(missingVars, "MINIO_CHAT_BUCKET")
	}
	if config.S3MinioConfig.PreviewBucketName == "" {
		missingVars = append(missingVars, "MINIO_PREVIEW_BUCKET")
	}

	// StableDiffusion configuration validation
	if os.Getenv("STABLE_DIFFUSION_URL") == "" {
		missingVars = append(missingVars, "STABLE_DIFFUSION_URL")
	}

	// Default admin and partner configuration validation
	if os.Getenv("DEFAULT_ADMIN_PASSWORD") == "" {
		missingVars = append(missingVars, "DEFAULT_ADMIN_PASSWORD")
	}
	if os.Getenv("DEFAULT_ADMIN_EMAIL") == "" {
		missingVars = append(missingVars, "DEFAULT_ADMIN_EMAIL")
	}
	if os.Getenv("DEFAULT_PASSWORD") == "" {
		missingVars = append(missingVars, "DEFAULT_PASSWORD")
	}
	if os.Getenv("FRONTEND_URL") == "" {
		missingVars = append(missingVars, "FRONTEND_URL")
	}

	// SMTP configuration validation
	if config.SMTPConfig.Host == "" {
		missingVars = append(missingVars, "SMTP_HOST")
	}
	if config.SMTPConfig.Password == "" {
		missingVars = append(missingVars, "SMTP_PASSWORD")
	}
	if config.SMTPConfig.Username == "" {
		missingVars = append(missingVars, "SMTP_USERNAME")
	}
	if config.SMTPConfig.From == "" {
		missingVars = append(missingVars, "SMTP_FROM")
	}

	// S3/MinIO configuration validation
	if config.S3MinioConfig.Endpoint == "" {
		missingVars = append(missingVars, "MINIO_ENDPOINT")
	}
	if config.S3MinioConfig.SecretAccessKey == "" {
		missingVars = append(missingVars, "MINIO_ROOT_PASSWORD")
	}
	if config.S3MinioConfig.AccessKeyID == "" {
		missingVars = append(missingVars, "MINIO_ROOT_USER")
	}
	if config.S3MinioConfig.Region == "" {
		missingVars = append(missingVars, "MINIO_REGION")
	}
	if config.S3MinioConfig.BucketName == "" {
		missingVars = append(missingVars, "MINIO_IMAGE_BUCKET")
	}
	if config.S3MinioConfig.LogosBucketName == "" {
		missingVars = append(missingVars, "MINIO_LOGOS_BUCKET")
	}
	if config.S3MinioConfig.ChatBucketName == "" {
		missingVars = append(missingVars, "MINIO_CHAT_BUCKET")
	}
	if config.S3MinioConfig.PreviewBucketName == "" {
		missingVars = append(missingVars, "MINIO_PREVIEW_BUCKET")
	}
	if config.S3MinioConfig.PublicURL == "" {
		missingVars = append(missingVars, "MINIO_PUBLIC_URL")
	}

	// Database and Redis URL validation
	if config.PostgresConfig.URL == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}
	if os.Getenv("POSTGRES_PASSWORD") == "" {
		missingVars = append(missingVars, "POSTGRES_PASSWORD")
	}
	if config.RedisConfig.URL == "" {
		missingVars = append(missingVars, "REDIS_URL")
	}
	if os.Getenv("REDIS_PASSWORD") == "" {
		missingVars = append(missingVars, "REDIS_PASSWORD")
	}

	// GitLab configuration validation
	if config.GitLabConfig.APIToken == "" {
		missingVars = append(missingVars, "GITLAB_API_TOKEN")
	}
	if config.GitLabConfig.TriggerToken == "" {
		missingVars = append(missingVars, "GITLAB_TRIGGER_TOKEN")
	}
	if config.GitLabConfig.ProjectID == "" {
		missingVars = append(missingVars, "GITLAB_PROJECT_ID")
	}

	if len(missingVars) > 0 {
		log.Error().
			Strs("missing_vars", missingVars).
			Msg("Missing required environment variables")
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	return nil
}
