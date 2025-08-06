package migrations

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
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

	// Создаем таблицы
	models := []interface{}{
		(*partner.Partner)(nil),
		(*admin.Admin)(nil),
		(*admin.ProfileChangeLog)(nil),
		(*coupon.Coupon)(nil),
		(*image.Image)(nil),
		(*payment.Order)(nil),
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

	// Создаем индексы для таблиц
	if err := createIndexes(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create indexes")
	}

	// Миграция кодов партнеров в строковый формат
	if err := migratePartnerCodes(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate partner codes")
	}

	// Создаем дефолтного администратора, если администраторов нет
	if err := createDefaultAdmin(database.DB, ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create default admin")
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
			CREATE TYPE coupon_status AS ENUM ('new', 'used');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$;`,

		// ENUM для статуса обработки изображений
		`DO $$ BEGIN
			CREATE TYPE processing_status AS ENUM ('queued', 'processing', 'completed', 'failed');
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
func createDefaultAdmin(db *bun.DB, ctx context.Context) error {
	count, err := db.NewSelect().Model((*admin.Admin)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("error getting admin count: %w", err)
	}

	if count > 0 {
		return nil
	}

	defaultPassword := "admin123"
	hashedPassword, err := bcrypt.HashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	defaultAdmin := &admin.Admin{
		Login:    "admin",
		Password: hashedPassword,
	}

	_, err = db.NewInsert().Model(defaultAdmin).Exec(ctx)
	if err != nil {
		return fmt.Errorf("error creating default admin: %w", err)
	}
	return nil
}

// migratePartnerCodes мигрирует коды партнеров из int16 в строковый формат с ведущими нулями
func migratePartnerCodes(db *bun.DB, ctx context.Context) error {
	var columnType string
	err := db.NewRaw(`
		SELECT data_type 
		FROM information_schema.columns 
		WHERE table_name = 'partners' AND column_name = 'partner_code'
	`).Scan(ctx, &columnType)

	if err != nil {
		return fmt.Errorf("error getting column type: %w", err)
	}

	if columnType == "character varying" {
		return nil
	}

	return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.Exec("ALTER TABLE partners ADD COLUMN partner_code_new VARCHAR(4)"); err != nil {
			return fmt.Errorf("error adding partner_code_new column: %w", err)
		}

		if _, err := tx.Exec(`
			UPDATE partners 
			SET partner_code_new = LPAD(CAST(partner_code AS TEXT), 4, '0')
		`); err != nil {
			return fmt.Errorf("error updating partners table: %w", err)
		}

		if _, err := tx.Exec("ALTER TABLE partners DROP COLUMN partner_code"); err != nil {
			return fmt.Errorf("error dropping partner_code column: %w", err)
		}

		if _, err := tx.Exec("ALTER TABLE partners RENAME COLUMN partner_code_new TO partner_code"); err != nil {
			return fmt.Errorf("error renaming partner_code_new column: %w", err)
		}

		if _, err := tx.Exec("ALTER TABLE partners ALTER COLUMN partner_code SET NOT NULL"); err != nil {
			return fmt.Errorf("error setting partner_code column to not null: %w", err)
		}

		if _, err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_partners_partner_code ON partners(partner_code)"); err != nil {
			return fmt.Errorf("error creating unique index for partners: %w", err)
		}

		return nil
	})
}
