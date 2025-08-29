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
		(*admin.Admin)(nil),
		(*admin.ProfileChangeLog)(nil),
		(*coupon.Coupon)(nil),
		(*image.Image)(nil),
		(*payment.Order)(nil),
		(*chat.Message)(nil),
		(*chat.SupportChat)(nil),
		(*chat.SupportMessage)(nil),
	}

	for _, model := range models {
		_, err := database.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create table")
		}
	}

	// Создаем ограничения внешнего ключа с каскадным удалением
	if err := createForeignKeys(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create foreign keys")
	}

	// Создаем индексы для таблиц (колонка уже существует)
	if err := createIndexes(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create indexes")
	}

	// Создаем дефолтного администратора, если администраторов нет
	if err := createDefaultAdmin(cfg, database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default admin")
	}

	// Создаем дефолтного партнера, если партнеров нет
	if err := createDefaultPartner(cfg, database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default partner")
	}
}

func createEnumTypes(db *bun.DB, ctx context.Context) error {
	enumQueries := []string{
		// ENUM для статуса партнера
		`DO $$ BEGIN
			CREATE TYPE partner_status AS ENUM ('active', 'blocked');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM для размеров купонов
		`DO $$ BEGIN
			CREATE TYPE coupon_size AS ENUM ('21x30', '30x40', '40x40', '40x50', '40x60', '50x70');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM для стилей купонов
		`DO $$ BEGIN
			CREATE TYPE coupon_style AS ENUM ('grayscale', 'skin_tones', 'pop_art', 'max_colors');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM для статуса купонов
		`DO $$ BEGIN
			CREATE TYPE coupon_status AS ENUM ('new', 'activated', 'used', 'completed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM для статуса обработки изображений
		`DO $$ BEGIN
			CREATE TYPE processing_status AS ENUM ('queued', 'uploaded', 'edited', 'processing', 'processed', 'completed', 'failed');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,
	}

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		for _, query := range enumQueries {
			if _, err := tx.Exec(query); err != nil {
				return fmt.Errorf("error creating enum type: %w", err)
			}
		}
		return nil
	})
}

// createForeignKeys создает ограничения внешнего ключа с каскадным удалением
func createForeignKeys(db *bun.DB, ctx context.Context) error {
	foreignKeyQueries := []string{
		// Ограничение между coupons и partners
		`DO $$ BEGIN
			ALTER TABLE coupons 
			ADD CONSTRAINT fk_coupons_partner_id 
			FOREIGN KEY (partner_id) REFERENCES partners(id) 
			ON DELETE CASCADE;
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// Ограничение между images и coupons
		`DO $$ BEGIN
			ALTER TABLE images 
			ADD CONSTRAINT fk_images_coupon_id 
			FOREIGN KEY (coupon_id) REFERENCES coupons(id) 
			ON DELETE CASCADE;
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// Ограничение между profile_changes и partners
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

// createIndexes создает индексы для таблиц
func createIndexes(db *bun.DB, ctx context.Context) error {
	partnerModel := &partner.Partner{}
	if _, err := db.ExecContext(ctx, partnerModel.CreateIndex()); err != nil {
		return fmt.Errorf("error creating index for partners: %w", err)
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

// createDefaultAdmin создает дефолтного администратора, если его нет
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

// Создает дефолтного партнера (собсвенный 0000)
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
		PartnerCode:     cfg.DefaultPartnerConfig.DefaultPartnerCode,
		Domain:          cfg.DefaultPartnerConfig.DefaultDomain,
		Login:           cfg.DefaultPartnerConfig.DefaultLogin,
		Email:           cfg.DefaultPartnerConfig.DefaultEmail,
		Password:        hashedPassword,
		BrandName:       cfg.DefaultPartnerConfig.DefaultBrandName,
		LogoURL:         cfg.DefaultPartnerConfig.DefaultLogo,
		Address:         cfg.DefaultPartnerConfig.DefaultAddress,
		Phone:           cfg.DefaultPartnerConfig.DefaultPhone,
		Telegram:        cfg.DefaultPartnerConfig.DefaultContactTelegram,
		Whatsapp:        cfg.DefaultPartnerConfig.DefaultWhatsapp,
		TelegramLink:    cfg.DefaultPartnerConfig.DefaultTelegramLink,
		WhatsappLink:    cfg.DefaultPartnerConfig.DefaultWhatsappLink,
		OzonLink:        cfg.DefaultPartnerConfig.DefaultOzonLink,
		WildberriesLink: cfg.DefaultPartnerConfig.DefaultWildberriesLink,
		Status:          "active",
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
