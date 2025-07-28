package main

//	@title			Mosaic API
//	@version		1.0
//	@description	API для системы мозаичных купонов
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:3000
//	@BasePath	/api
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
	_ "github.com/skr1ms/mosaic/docs" // Swagger docs
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/auth"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/public"
	"github.com/skr1ms/mosaic/migrations"
	"github.com/skr1ms/mosaic/pkg/db"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/recaptcha"
)

// generateRequestID создает уникальный ID для запроса
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	migrations.Init(cfg)
	database := db.NewDb(cfg)

	// Настройка логгера с более детальной конфигурацией
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "mosaic-api").
		Caller().
		Logger()

	// Устанавливаем уровень логирования в зависимости от окружения
	if os.Getenv("APP_MODE") == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Логируем все HTTP ошибки
			logger.Error().
				Err(err).
				Int("status_code", code).
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("ip", c.IP()).
				Msg("HTTP Error")

			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Language",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

	app.Use(recover.New())

	// Middleware для логирования запросов с дополнительной информацией
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
		Fields: []string{"ip", "method", "url", "status", "latency", "user_agent"},
	}))

	// Middleware для добавления request_id в контекст
	app.Use(func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Добавляем логгер с request_id в контекст
		loggerWithRequestID := logger.With().Str("request_id", requestID).Logger()
		c.SetUserContext(loggerWithRequestID.WithContext(c.UserContext()))

		return c.Next()
	})

	// swagger ui middleware
	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))

	// API
	api := app.Group("/api")

	// repository
	adminRepo := admin.NewAdminRepository(database.DB)
	couponRepo := coupon.NewCouponRepository(database.DB)
	partnerRepo := partner.NewPartnerRepository(database.DB)
	imageRepo := image.NewRepository(database.DB)

	// service
	mailSender := email.NewMailer(cfg, &logger)
	recaptchService := recaptcha.NewVerifier(cfg.RecaptchaConfig.SecretKey, 0.5)
	jwtService := jwt.NewJWT(cfg.AuthConfig.AccessTokenSecret, cfg.AuthConfig.RefreshTokenSecret)
	authService := auth.NewAuthService(&auth.AuthServiceDeps{
		PartnerRepository: partnerRepo,
		AdminRepository:   adminRepo,
		JwtService:        jwtService,
	})

	adminService := admin.NewAdminService(&admin.AdminServiceDeps{
		AdminRepository:   adminRepo,
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
	})

	partnerService := partner.NewPartnerService(&partner.PartnerServiceDeps{
		PartnerRepository: partnerRepo,
		Recaptcha:         recaptchService,
		JwtService:        jwtService,
		MailSender:        mailSender,
		Config:            cfg,
	})

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: couponRepo,
	})

	imageService := image.NewImageService(&image.ImageServiceDeps{
		ImageRepository:  imageRepo,
		CouponRepository: couponRepo,
	})

	publicService := public.NewPublicService(&public.PublicServiceDeps{
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		PartnerRepository: partnerRepo,
		ImageService:      imageService,
	})

	// handlers
	admin.NewAdminHandler(api, &admin.AdminHandlerDeps{
		AdminService: adminService,
		JwtService:   jwtService,
	})

	auth.NewAuthHandler(api, &auth.AuthHandlerDeps{
		AuthService: authService,
	})

	partner.NewPartnerHandler(api, &partner.PartnerHandlerDeps{
		Config:           cfg,
		PartnerService:   partnerService,
		CouponRepository: couponRepo,
		JwtService:       jwtService,
		MailSender:       mailSender,
	})

	coupon.NewCouponHandler(api, &coupon.CouponHandlerDeps{
		CouponService: couponService,
	})

	image.NewImageProcessingHandler(api, &image.ImageHandlerDeps{
		ImageRepository:  imageRepo,
		CouponRepository: couponRepo,
	})

	public.NewPublicHandler(app, &public.PublicHandlerDeps{
		PublicService: publicService,
	})

	log.Info().Msg("Server is running on port 3000")
	app.Listen(":3000")
}
