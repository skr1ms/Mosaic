package main

// @title Mosaic API
// @version 1.0
// @description API для системы мозаичных купонов
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
	_ "github.com/skr1ms/mosaic/docs" // Swagger docs
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image_processing"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/migrations"
	"github.com/skr1ms/mosaic/pkg/db"
	"github.com/skr1ms/mosaic/pkg/jwt"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// Создаем JWT сервис
	jwtService := jwt.NewJWT(cfg.Auth.SecretKey, cfg.Auth.RefreshSecretKey)

	migrations.Init(cfg)
	database := db.NewDB(cfg)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Language",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

	app.Use(logger.New())
	app.Use(recover.New())

	// swagger ui middleware
	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))

	// API
	api := app.Group("/api")

	// handlers с передачей JWT сервиса
	admin.NewAdminHandler(api, database.DB, jwtService)
	partner.NewPartnerHandler(api, database.DB, jwtService)
	coupon.NewCouponHandler(api, database.DB)
	image_processing.NewImageProcessingHandler(api, database.DB)

	log.Info().Msg("Server is running on port 3000")
	app.Listen(":3000")
}
