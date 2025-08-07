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

//	@host		localhost:8080
//	@BasePath	/api
//	@schemes	http https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

import (
	"time"

	adaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/fiberi18n/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
	_ "github.com/skr1ms/mosaic/docs" // Swagger docs
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/auth"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/public"
	"github.com/skr1ms/mosaic/internal/stats"
	"github.com/skr1ms/mosaic/migrations"
	"github.com/skr1ms/mosaic/pkg/db"
	"github.com/skr1ms/mosaic/pkg/email"
	"github.com/skr1ms/mosaic/pkg/errors"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/queue"
	"github.com/skr1ms/mosaic/pkg/recaptcha"
	"github.com/skr1ms/mosaic/pkg/redis"
	"github.com/skr1ms/mosaic/pkg/s3"
	"github.com/skr1ms/mosaic/pkg/stablediffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
	"golang.org/x/text/language"
)

func InitializeApp() *fiber.App {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	migrations.Init(cfg)
	database, err := db.NewDb(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database")
	}

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	s3Client, err := s3.NewS3Client(cfg.S3MinioConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create S3 client")
	}

	stableDiffusionClient := stablediffusion.NewStableDiffusionClient(cfg.StableDiffusionConfig)

	queueManager := queue.NewQueueManager(redisClient)

	appLogger := middleware.NewLogger()

	app := fiber.New(fiber.Config{
		ErrorHandler: errors.ErrorHandler(),
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Language",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

	app.Use(recover.New())

	app.Use(appLogger.CombinedMiddleware())

	app.Use(appLogger.AnalyticsMiddleware())

	app.Use(middleware.GeneralRateLimiter())

	app.Use(middleware.AuditLogger())

	app.Use(fiberi18n.New(&fiberi18n.Config{
		RootPath:        "./locales",
		AcceptLanguages: []language.Tag{language.Russian, language.English, language.Spanish},
		DefaultLanguage: language.Russian,
	}))

	brandingMiddleware := middleware.BrandingMiddleware(database.DB, middleware.DefaultBranding{
		BrandName:       cfg.BrandingConfig.DefaultBrandName,
		LogoURL:         cfg.BrandingConfig.DefaultLogoURL,
		ContactEmail:    cfg.BrandingConfig.DefaultContactEmail,
		ContactAddress:  cfg.BrandingConfig.DefaultContactAddress,
		ContactPhone:    cfg.BrandingConfig.DefaultContactPhone,
		ContactTelegram: cfg.BrandingConfig.DefaultContactTelegram,
		ContactWhatsapp: cfg.BrandingConfig.DefaultContactWhatsapp,
		TelegramLink:    cfg.BrandingConfig.DefaultTelegramLink,
		WhatsappLink:    cfg.BrandingConfig.DefaultWhatsappLink,
		OzonLink:        cfg.BrandingConfig.DefaultOzonLink,
		WildberriesLink: cfg.BrandingConfig.DefaultWildberriesLink,
	})
	app.Use(brandingMiddleware)

	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.json",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))

	// API
	api := app.Group("/api")

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		})
	})

	// repository
	adminRepo := admin.NewAdminRepository(database.DB)
	couponRepo := coupon.NewCouponRepository(database.DB)
	partnerRepo := partner.NewPartnerRepository(database.DB)
	imageRepo := image.NewRepository(database.DB)
	paymentRepo := payment.NewPaymentRepository(database.DB)

	// service
	mailSender := email.NewMailer(cfg)
	recaptchService := recaptcha.NewVerifier(cfg.RecaptchaConfig.SecretKey, 0.5)
	jwtService := jwt.NewJWT(cfg.AuthConfig.AccessTokenSecret, cfg.AuthConfig.RefreshTokenSecret)
	zipService := zip.NewZipService()
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
		S3Client:          s3Client,
		RedisClient:       redisClient,
	})

	partnerService := partner.NewPartnerService(&partner.PartnerServiceDeps{
		PartnerRepository: partnerRepo,
		Recaptcha:         recaptchService,
		JwtService:        &partner.JWTAdapter{JWT: jwtService},
		MailSender:        mailSender,
		Config:            cfg,
	})

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: couponRepo,
		RedisClient:      redisClient,
	})

	imageService := image.NewImageService(&image.ImageServiceDeps{
		ImageRepository:       imageRepo,
		CouponRepository:      couponRepo,
		S3Client:              s3Client,
		StableDiffusionClient: stableDiffusionClient,
		EmailService:          mailSender,
		ZipService:            zipService,
	})

	paymentService := payment.NewPaymentService(&payment.PaymentServiceDeps{
		PaymentRepository: paymentRepo,
		CouponRepository:  couponRepo,
		PartnerRepository: partnerRepo,
		Config:            cfg,
	})

	publicService := public.NewPublicService(&public.PublicServiceDeps{
		Config:            cfg,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		PartnerRepository: partnerRepo,
		ImageService:      imageService,
		PaymentService:    paymentService,
		EmailService:      mailSender,
	})

	statsService := stats.NewStatsService(&stats.StatsServiceDeps{
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		RedisClient:       redisClient,
	})

	cronService := stats.NewCronService(statsService)
	cronService.Start()

	imageAdapter := queue.NewImageServiceAdapter(imageService)
	emailAdapter := queue.NewEmailServiceAdapter(mailSender)

	queueManager.StartAllWorkers(imageAdapter, emailAdapter)

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

	payment.NewPaymentHandler(api, &payment.PaymentHandlerDeps{
		PaymentService:   paymentService,
		CouponRepository: couponRepo,
	})

	image.NewImageProcessingHandler(api, &image.ImageHandlerDeps{
		ImageService:    imageService,
		ImageRepository: imageRepo,
	})

	public.NewPublicHandler(app, &public.PublicHandlerDeps{
		PublicService: publicService,
	})

	stats.NewStatsHandler(api, &stats.StatsHandlerDeps{
		StatsService: statsService,
	})

	// Prometheus metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Запускаем метрики сервер на отдельном порту
	go func() {
		metricsApp := fiber.New()
		metricsApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

		log.Info().Msgf("Metrics server is running on port %s", cfg.MetricsConfig.Port)
		metricsApp.Listen(":" + cfg.MetricsConfig.Port)
	}()

	log.Info().Msgf("Server is running on port %s", cfg.ServerConfig.Port)

	return app
}
