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
	"os"

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

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	migrations.Init(cfg)
	database := db.NewDB(cfg)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Language",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS, PATCH",
	}))

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
	// repository
	adminRepo := admin.NewAdminRepository(database.DB)
	couponRepo := coupon.NewCouponRepository(database.DB)
	imageRepo := image.NewRepository(database.DB)
	partnerRepo := partner.NewPartnerRepository(database.DB)
	imageRepo = image.NewRepository(database.DB)

	// service
	mailSender := email.NewMailer(cfg, &logger)
	recaptchService := recaptcha.NewVerifier(cfg.RecaptchaConfig.SecretKey, 0.5)
	jwtService := jwt.NewJWT(cfg.AuthConfig.AccessTokenSecret, cfg.AuthConfig.RefreshTokenSecret)
	authService := auth.NewAuthService(
		partner.NewPartnerRepository(database.DB),
		admin.NewAdminRepository(database.DB),
		jwtService,
		&logger,
	)

	adminService := admin.NewAdminService(&admin.AdminServiceDeps{
		AdminRepository:   adminRepo,
		PartnerRepository: partnerRepo,
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		Logger:            &logger,
	})

	partnerService := partner.NewPartnerService(
		partnerRepo,
		recaptchService,
		jwtService,
		mailSender,
		cfg,
		&logger,
	)

	imageService := image.NewImageService(
		imageRepo,
		couponRepo,
		&logger,
	)

	// handlers
	admin.NewAdminHandler(api, &admin.AdminHandlerDeps{
		AdminService:      adminService,
		JwtService:        jwtService,
	})

	auth.NewAuthHandler(api, &auth.AuthHandlerDeps{
		AuthService: authService,
		Logger:      &logger,
	})

	partner.NewPartnerHandler(api, &partner.PartnerHandlerDeps{
		Config:           cfg,
		PartnerService:   partnerService,
		CouponRepository: couponRepo,
		JwtService:       jwtService,
		MailSender:       mailSender,
		Logger:           &logger,
	})

	coupon.NewCouponHandler(api, &coupon.CouponHandlerDeps{
		CouponRepository: couponRepo,
	})

	image.NewImageProcessingHandler(api, &image.ImageHandlerDeps{
		ImageRepository:  imageRepo,
		CouponRepository: couponRepo,
	})

	public.NewPublicHandler(app, &public.PublicService{
		CouponRepository:  couponRepo,
		ImageRepository:   imageRepo,
		PartnerRepository: partnerRepo,
		ImageService:      imageService,
		Logger:            &logger,
	})

	log.Info().Msg("Server is running on port 3000")
	app.Listen(":3000")
}
