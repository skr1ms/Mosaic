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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	adaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/skr1ms/mosaic/config"

	// _ "github.com/skr1ms/mosaic/docs" // Swagger docs
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/auth"
	"github.com/skr1ms/mosaic/internal/chat"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/public"
	"github.com/skr1ms/mosaic/internal/stats"
	"github.com/skr1ms/mosaic/migrations"
	"github.com/skr1ms/mosaic/pkg/db"
	"github.com/skr1ms/mosaic/pkg/email"

	"github.com/skr1ms/mosaic/pkg/gitlab"
	"github.com/skr1ms/mosaic/pkg/goroutine"
	"github.com/skr1ms/mosaic/pkg/jwt"
	"github.com/skr1ms/mosaic/pkg/middleware"
	"github.com/skr1ms/mosaic/pkg/mosaic"
	"github.com/skr1ms/mosaic/pkg/palette"
	"github.com/skr1ms/mosaic/pkg/queue"
	"github.com/skr1ms/mosaic/pkg/recaptcha"
	"github.com/skr1ms/mosaic/pkg/redis"
	"github.com/skr1ms/mosaic/pkg/s3"
	"github.com/skr1ms/mosaic/pkg/stableDiffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
)

func InitializeApp() *fiber.App {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
	appLogger := middleware.NewLogger()

	// Creating temporary directories for image processing
	tempDirs := []string{"/tmp/originals", "/tmp/edited", "/tmp/processed", "/tmp/previews", "/tmp/mosaic_output"}
	for _, dir := range tempDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create temp directory %s: %v", dir, err))
		}
	}

	migrations.Init(cfg)
	database, err := db.NewDb(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create database: %v", err))
	}

	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	s3Client, err := s3.NewS3Client(cfg.S3MinioConfig, appLogger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create S3 client: %v", err))
	}

	gitlabClient := gitlab.NewClient(
		cfg.GitLabConfig.BaseURL,
		cfg.GitLabConfig.APIToken,
		cfg.GitLabConfig.ProjectID,
	)

	alfaBankClient := payment.NewAlfaBankClient(cfg)

	stableDiffusionClient := stableDiffusion.NewStableDiffusionClient(cfg.StableDiffusionConfig, appLogger)

	queueManager := queue.NewQueueManager(redisClient, appLogger)

	app := fiber.New(fiber.Config{
		ErrorHandler: appLogger.ErrorHandler(),
		BodyLimit:    50 * 1024 * 1024,
		ReadTimeout:  time.Second * 60,
		WriteTimeout: time.Second * 60,
		IdleTimeout:  time.Second * 60,
	})
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			// Allow empty origin for same-origin requests
			if origin == "" {
				return true
			}

			// Production domains
			if origin == "https://photo.doyoupaint.com" || origin == "https://adm.doyoupaint.com" {
				return true
			}

			// Development origins
			return strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
		},
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Authorization, Content-Type, Accept, Origin, X-Requested-With, X-CSRF-Token",
		AllowCredentials: true,
	}))

	app.Use(recover.New())
	app.Use(appLogger.RequestIDMiddleware())

	// Skip logging for WebSocket paths and health checks
	app.Use(appLogger.SkipLoggingMiddleware("/api/ws/chat", "/health", "/metrics"))

	app.Use(func(c *fiber.Ctx) error {
		// Skip WebSocket paths
		if c.Path() == "/api/ws/chat" {
			return c.Next()
		}
		return middleware.GeneralRateLimiter(appLogger)(c)
	})

	app.Use(func(c *fiber.Ctx) error {
		// Skip WebSocket paths
		if c.Path() == "/api/ws/chat" {
			return c.Next()
		}
		return middleware.AuditLogger(appLogger)(c)
	})

	brandingMiddleware := middleware.NewBrandingMiddleware(database.DB, middleware.DefaultBranding{
		BrandName:       cfg.DefaultPartnerConfig.DefaultBrandName,
		LogoURL:         cfg.DefaultPartnerConfig.DefaultLogo,
		ContactEmail:    cfg.DefaultPartnerConfig.DefaultEmail,
		ContactAddress:  cfg.DefaultPartnerConfig.DefaultAddress,
		ContactPhone:    cfg.DefaultPartnerConfig.DefaultPhone,
		ContactTelegram: cfg.DefaultPartnerConfig.DefaultContactTelegram,
		ContactWhatsapp: cfg.DefaultPartnerConfig.DefaultWhatsapp,
		TelegramLink:    cfg.DefaultPartnerConfig.DefaultTelegramLink,
		WhatsappLink:    cfg.DefaultPartnerConfig.DefaultWhatsappLink,
		OzonLink:        cfg.DefaultPartnerConfig.DefaultOzonLink,
		WildberriesLink: cfg.DefaultPartnerConfig.DefaultWildberriesLink,
	}, appLogger)

	app.Use(brandingMiddleware.BrandingMiddlewareHandler())

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
	chatRepo := chat.NewRepository(database.DB)

	// service
	mailSender := email.NewMailer(cfg, appLogger)
	recaptchService := recaptcha.NewVerifier(cfg.RecaptchaConfig.SecretKey, 0.5, appLogger)
	jwtService := jwt.NewJWT(cfg.AuthConfig.AccessTokenSecret, cfg.AuthConfig.RefreshTokenSecret)

	zipService := zip.NewZipService(appLogger)
	authService := auth.NewAuthService(&auth.AuthServiceDeps{
		PartnerRepository: partnerRepo,
		AdminRepository:   adminRepo,
		JwtService:        jwtService,
		Recaptcha:         recaptchService,
		MailSender:        mailSender,
		Config:            cfg,
	})

	// Initialize goroutine manager
	goroutineManager := goroutine.NewManager(context.Background())

	adminService := admin.NewAdminService(&admin.AdminServiceDeps{
		AdminRepository:   adminRepo,
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		S3Client:          s3Client,
		RedisClient:       redisClient,
		GitLabClient:      gitlabClient,
		GoroutineManager:  goroutineManager,
		Logger:            middleware.NewLogger().GetZerologLogger(),
	})

	partnerService := partner.NewPartnerService(&partner.PartnerServiceDeps{
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		S3Client:          s3Client,
		Recaptcha:         recaptchService,
		JwtService:        jwtService,
		MailSender:        mailSender,
		Config:            cfg,
	})

	couponService := coupon.NewCouponService(&coupon.CouponServiceDeps{
		CouponRepository: couponRepo,
		RedisClient:      redisClient,
		S3Client:         s3Client,
	})

	paletteService := palette.NewPaletteService(cfg.MosaicGeneratorConfig.PalettePath, appLogger)

	mosaicGenerator := mosaic.NewMosaicGenerator(
		cfg.MosaicGeneratorConfig.ScriptPath,
		cfg.MosaicGeneratorConfig.OutputDir,
		cfg.MosaicGeneratorConfig.PythonCommand,
		appLogger,
		goroutineManager,
	)

	imageService := image.NewImageService(&image.ImageServiceDeps{
		ImageRepository:       imageRepo,
		CouponRepository:      couponRepo,
		S3Client:              s3Client,
		StableDiffusionClient: stableDiffusionClient,
		EmailService:          mailSender,
		ZipService:            zipService,
		MosaicGenerator:       mosaicGenerator,
		PaletteService:        paletteService,
		WorkingDir:            "/tmp",
	})

	paymentService := payment.NewPaymentService(&payment.PaymentServiceDeps{
		PaymentRepository: paymentRepo,
		CouponRepository:  couponRepo,
		PartnerRepository: partnerRepo,
		Config:            cfg,
		AlfaBankClient:    alfaBankClient,
	})

	publicService := public.NewPublicService(&public.PublicServiceDeps{
		Config:            cfg,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		PartnerRepository: partnerRepo,
		ImageService:      imageService,
		PaymentService:    paymentService,
		EmailService:      mailSender,
		RecaptchaSiteKey:  cfg.RecaptchaConfig.SiteKey,
	})

	statsService := stats.NewStatsService(&stats.StatsServiceDeps{
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		RedisClient:       redisClient,
	})

	chatService := chat.NewChatService(chatRepo, s3Client)

	cronService := stats.NewCronService(statsService)
	cronService.Start()

	imageAdapter := queue.NewImageServiceAdapter(imageService)
	emailAdapter := queue.NewEmailServiceAdapter(mailSender)

	queueManager.StartAllWorkers(imageAdapter, emailAdapter)

	// handlers
	chat.NewChatHandler(api, &chat.ChatHandlerDeps{
		ChatService:      chatService,
		JwtService:       jwtService,
		Logger:           appLogger,
		GoroutineManager: goroutineManager,
	})

	admin.NewAdminHandler(api, &admin.AdminHandlerDeps{
		AdminService: adminService,
		JwtService:   jwtService,
		Logger:       appLogger,
	})

	auth.NewAuthHandler(api.Group("/auth"), &auth.AuthHandlerDeps{
		AuthService:    authService,
		PartnerService: partnerService,
		JwtService:     jwtService,
		Logger:         appLogger,
	})

	partner.NewPartnerHandler(api, &partner.PartnerHandlerDeps{
		Config:           cfg,
		PartnerService:   partnerService,
		CouponRepository: couponRepo,
		JwtService:       jwtService,
		MailSender:       mailSender,
		Logger:           appLogger,
	})

	payment.NewPaymentHandler(api, &payment.PaymentHandlerDeps{
		PaymentService:   paymentService,
		CouponRepository: couponRepo,
		Logger:           appLogger,
	})

	image.NewImageProcessingHandler(api, &image.ImageHandlerDeps{
		ImageService:    imageService,
		ImageRepository: imageRepo,
		Logger:          appLogger,
	})

	public.NewPublicHandler(api, &public.PublicHandlerDeps{
		PublicService: publicService,
		S3Client:      s3Client,
		Logger:        appLogger,
	})

	stats.NewStatsHandler(api, &stats.StatsHandlerDeps{
		StatsService: statsService,
		JwtService:   jwtService,
		Logger:       appLogger,
	})

	coupon.NewCouponHandler(api, &coupon.CouponHandlerDeps{
		CouponService: couponService,
		Logger:        appLogger,
	})

	// Prometheus metrics endpoint
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// Running the metrics server on a separate port - используем обычную горутину для долгоживущего процесса
	go func() {
		metricsApp := fiber.New()
		metricsApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

		// Metrics server is running on port
		metricsApp.Listen(":" + cfg.MetricsConfig.Port)
	}()

	// Server is running on port

	return app
}
