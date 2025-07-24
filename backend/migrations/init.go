package migrations

import (
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/admin"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image_processing"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/db"
	password "github.com/skr1ms/mosaic/pkg/bcrypt"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) {
	database := db.NewDB(cfg)

	if err := createEnumTypes(database.DB); err != nil {
		panic("Failed to create enum types: " + err.Error())
	}

	err := database.AutoMigrate(
		&partner.Partner{},
		&admin.Admin{},
		&coupon.Coupon{},
		&image_processing.ImageProcessingQueue{},
	)
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	// Создаем дефолтного администратора, если администраторов нет
	if err := createDefaultAdmin(database.DB); err != nil {
		panic("Failed to create default admin: " + err.Error())
	}
}

func createEnumTypes(db *gorm.DB) error {
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

	return db.Transaction(func(tx *gorm.DB) error {
		for _, query := range enumQueries {
			if err := tx.Exec(query).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// createDefaultAdmin создает дефолтного администратора, если его нет
func createDefaultAdmin(db *gorm.DB) error {
	var count int64
	if err := db.Model(&admin.Admin{}).Count(&count).Error; err != nil {
		return err
	}

	// Если есть хотя бы один админ, ничего не делаем
	if count > 0 {
		return nil
	}

	// Создаем дефолтного админа
	defaultPassword := "admin123"
	hashedPassword, err := password.HashPassword(defaultPassword)
	if err != nil {
		return err
	}

	defaultAdmin := &admin.Admin{
		Login:    "admin",
		Password: hashedPassword,
	}

	if err := db.Create(defaultAdmin).Error; err != nil {
		return err
	}

	return nil
}
