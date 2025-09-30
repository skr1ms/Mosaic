package migrations

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/chat"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/public"
	"github.com/skr1ms/mosaic/pkg/bcrypt"
	"github.com/skr1ms/mosaic/pkg/db"
	"github.com/uptrace/bun"
)

func Init(cfg *config.Config) {
	database, err := db.NewDb(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create database")
	}
	ctx := context.Background()

	if err := createEnumTypes(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create enum types")
	}

	models := []any{
		(*partner.Partner)(nil),
		(*partner.PartnerArticle)(nil),
		(*admin.Admin)(nil),
		(*admin.ProfileChangeLog)(nil),
		(*coupon.Coupon)(nil),
		(*image.Image)(nil),
		(*payment.Order)(nil),
		(*chat.Message)(nil),
		(*chat.SupportChat)(nil),
		(*chat.SupportMessage)(nil),
		(*public.PreviewData)(nil),
	}

	for _, model := range models {
		_, err := database.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create table")
		}
	}

	// Create foreign key constraints with cascade deletion
	if err := createForeignKeys(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create foreign keys")
	}

	// Create indexes for tables (column already exists)
	if err := createIndexes(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create indexes")
	}

	// Create default administrator if no administrators exist
	if err := createDefaultAdmin(cfg, database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default admin")
	}

	// Create default partner if no partners exist
	if err := createDefaultPartner(cfg, database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default partner")
	}

	// Initialize articles for default partner
	if err := initializeDefaultPartnerArticles(cfg, database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize default partner articles")
	}
}

func createEnumTypes(db *bun.DB, ctx context.Context) error {
	enumQueries := []string{
		// ENUM for partner status
		`DO $$ BEGIN
			CREATE TYPE partner_status AS ENUM ('active', 'blocked');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM for coupon sizes
		`DO $$ BEGIN
			CREATE TYPE coupon_size AS ENUM ('21x30', '30x40', '40x40', '40x50', '40x60', '50x70');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM for coupon styles
		`DO $$ BEGIN
			CREATE TYPE coupon_style AS ENUM ('grayscale', 'skin_tones', 'pop_art', 'max_colors');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM for coupon status
		`DO $$ BEGIN
			CREATE TYPE coupon_status AS ENUM ('new', 'activated', 'used', 'completed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM for image processing status
		`DO $$ BEGIN
			CREATE TYPE processing_status AS ENUM ('queued', 'uploaded', 'edited', 'processing', 'processed', 'completed', 'failed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
	}

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, query := range enumQueries {
			if _, err := tx.Exec(query); err != nil {
				log.Error().
					Err(err).
					Str("query", query).
					Msg("Failed to create enum type")
				return fmt.Errorf("error creating enum type: %w", err)
			}
		}
		return nil
	})
}

// createForeignKeys creates foreign key constraints with cascade deletion
func createForeignKeys(db *bun.DB, ctx context.Context) error {
	foreignKeyQueries := []string{
		// Constraint between coupons and partners
		`DO $$ BEGIN
			ALTER TABLE coupons 
			ADD CONSTRAINT fk_coupons_partner_id 
			FOREIGN KEY (partner_id) REFERENCES partners(id) 
			ON DELETE CASCADE;
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// Constraint between images and coupons
		`DO $$ BEGIN
			ALTER TABLE images 
			ADD CONSTRAINT fk_images_coupon_id 
			FOREIGN KEY (coupon_id) REFERENCES coupons(id) 
			ON DELETE CASCADE;
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// Constraint between profile_changes and partners
		`DO $$ BEGIN
			ALTER TABLE profile_changes 
			ADD CONSTRAINT fk_profile_changes_partner_id 
			FOREIGN KEY (partner_id) REFERENCES partners(id) 
			ON DELETE CASCADE;
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
	}

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, query := range foreignKeyQueries {
			if _, err := tx.Exec(query); err != nil {
				return fmt.Errorf("error creating foreign key: %w", err)
			}
		}
		return nil
	})
}

// createIndexes creates indexes for tables
func createIndexes(db *bun.DB, ctx context.Context) error {
	partnerModel := &partner.Partner{}
	if _, err := db.ExecContext(ctx, partnerModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for partners: %w", err)
	}

	if _, err := db.ExecContext(ctx, partner.CreatePartnerArticlesIndex()); err != nil {
		return fmt.Errorf("error creating index for partner articles: %w", err)
	}

	adminModel := &admin.Admin{}
	if _, err := db.ExecContext(ctx, adminModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for admins: %w", err)
	}

	profileChangeModel := &admin.ProfileChangeLog{}
	if _, err := db.ExecContext(ctx, profileChangeModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for profile changes: %w", err)
	}

	couponModel := &coupon.Coupon{}
	if _, err := db.ExecContext(ctx, couponModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for coupons: %w", err)
	}

	imageModel := &image.Image{}
	if _, err := db.ExecContext(ctx, imageModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for images: %w", err)
	}

	orderModel := &payment.Order{}
	if _, err := db.ExecContext(ctx, orderModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for orders: %w", err)
	}

	return nil
}

// createDefaultAdmin creates default administrator if none exists
func createDefaultAdmin(cfg *config.Config, db *bun.DB, ctx context.Context) error {
	count, err := db.NewSelect().Model((*admin.Admin)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("error getting admin count: %w", err)
	}

	if count > 0 {
		log.Info().Msg("Default admin already exists, skipping creation")
		return nil
	}

	hashedPassword, err := bcrypt.HashPassword(cfg.DefaultAdminConfig.DefaultPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	defaultAdmin := &admin.Admin{
		Login:    cfg.DefaultAdminConfig.DefaultLogin,
		Email:    cfg.DefaultAdminConfig.DefaultEmail,
		Password: hashedPassword,
		Role:     "main_admin",
	}

	_, err = db.NewInsert().Model(defaultAdmin).Exec(ctx)
	if err != nil {
		return fmt.Errorf("error creating default admin: %w", err)
	}

	log.Info().Str("login", cfg.DefaultAdminConfig.DefaultLogin).Msg("Default admin created successfully")
	return nil
}

// Creates default partner (own 0000)
func createDefaultPartner(cfg *config.Config, db *bun.DB, ctx context.Context) error {
	count, err := db.NewSelect().Model((*partner.Partner)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("error getting partner count: %w", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.HashPassword(cfg.DefaultPartnerConfig.DefaultPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	defaultPartner := &partner.Partner{
		PartnerCode:             cfg.DefaultPartnerConfig.DefaultPartnerCode,
		Domain:                  cfg.DefaultPartnerConfig.DefaultDomain,
		Login:                   cfg.DefaultPartnerConfig.DefaultLogin,
		Email:                   cfg.DefaultPartnerConfig.DefaultEmail,
		Password:                hashedPassword,
		BrandName:               cfg.DefaultPartnerConfig.DefaultBrandName,
		LogoURL:                 cfg.DefaultPartnerConfig.DefaultLogo,
		Address:                 cfg.DefaultPartnerConfig.DefaultAddress,
		Phone:                   cfg.DefaultPartnerConfig.DefaultPhone,
		Telegram:                cfg.DefaultPartnerConfig.DefaultContactTelegram,
		Whatsapp:                cfg.DefaultPartnerConfig.DefaultWhatsapp,
		TelegramLink:            cfg.DefaultPartnerConfig.DefaultTelegramLink,
		WhatsappLink:            cfg.DefaultPartnerConfig.DefaultWhatsappLink,
		OzonLink:                cfg.DefaultPartnerConfig.DefaultOzonLink,
		WildberriesLink:         cfg.DefaultPartnerConfig.DefaultWildberriesLink,
		OzonLinkTemplate:        "https://www.ozon.ru/search/?text={sku}+{size}+{style}",
		WildberriesLinkTemplate: "https://www.wildberries.ru/catalog/{sku}/detail.aspx",
		Status:                  "active",
	}

	_, err = db.NewInsert().Model(defaultPartner).Exec(ctx)
	if err != nil {
		return fmt.Errorf("error creating default partner: %w", err)
	}

	partnerExists, err := db.NewSelect().Model((*partner.Partner)(nil)).Where("partner_code = ?", "0000").Exists(ctx)
	if err != nil {
		return fmt.Errorf("error checking if partner exists: %w", err)
	}
	if !partnerExists {
		_, err = db.NewInsert().Model(defaultPartner).Exec(ctx)
		if err != nil {
			return fmt.Errorf("error creating default partner: %w", err)
		}
	}

	return nil
}

// initializeDefaultPartnerArticles initializes empty article grid for default partner
func initializeDefaultPartnerArticles(cfg *config.Config, db *bun.DB, ctx context.Context) error {
	// Find default partner
	var defaultPartner partner.Partner
	err := db.NewSelect().
		Model(&defaultPartner).
		Where("partner_code = ?", cfg.DefaultPartnerConfig.DefaultPartnerCode).
		Scan(ctx)

	if err != nil {
		// If partner not found, maybe it already existed before - this is normal
		log.Info().Msg("Default partner not found or already processed, skipping article initialization")
		return nil
	}

	// Check if this partner already has articles
	count, err := db.NewSelect().
		Model((*partner.PartnerArticle)(nil)).
		Where("partner_id = ?", defaultPartner.ID).
		Count(ctx)

	if err != nil {
		return fmt.Errorf("error checking existing articles: %w", err)
	}

	if count > 0 {
		log.Info().Msg("Default partner articles already exist, skipping initialization")
		return nil
	}

	// Create empty article grid (48 cells: 4 styles × 6 sizes × 2 marketplaces)
	var articles []partner.PartnerArticle

	marketplaces := []string{"ozon", "wildberries"}
	styles := []string{"grayscale", "skin_tones", "pop_art", "max_colors"}
	sizes := []string{"21x30", "30x40", "40x40", "40x50", "40x60", "50x70"}

	for _, marketplace := range marketplaces {
		for _, style := range styles {
			for _, size := range sizes {
				articles = append(articles, partner.PartnerArticle{
					PartnerID:   defaultPartner.ID,
					Size:        size,
					Style:       style,
					Marketplace: marketplace,
					SKU:         "", // empty SKU
					IsActive:    true,
				})
			}
		}
	}

	// Insert all cells into DB
	_, err = db.NewInsert().Model(&articles).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize default partner articles: %w", err)
	}

	log.Info().
		Str("partner_id", defaultPartner.ID.String()).
		Int("article_count", len(articles)).
		Msg("Successfully initialized article grid for default partner")

	return nil
}
