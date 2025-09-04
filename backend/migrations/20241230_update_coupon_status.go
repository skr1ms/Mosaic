package migrations

import (
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

// UpdateCouponStatus updates the coupon_status enum to only have 'new' and 'activated'
func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		// Start transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Update existing statuses to map to new ones
		// 'used' and 'completed' -> 'activated'
		_, err = tx.Exec(`
			UPDATE coupons 
			SET status = 'activated' 
			WHERE status IN ('used', 'completed')
		`)
		if err != nil {
			return fmt.Errorf("failed to update coupon statuses: %w", err)
		}

		// Drop the old enum type
		_, err = tx.Exec(`ALTER TYPE coupon_status RENAME TO coupon_status_old`)
		if err != nil {
			return fmt.Errorf("failed to rename old enum type: %w", err)
		}

		// Create new enum type with only 'new' and 'activated'
		_, err = tx.Exec(`CREATE TYPE coupon_status AS ENUM ('new', 'activated')`)
		if err != nil {
			return fmt.Errorf("failed to create new enum type: %w", err)
		}

		// Update the column to use the new enum type
		_, err = tx.Exec(`
			ALTER TABLE coupons 
			ALTER COLUMN status TYPE coupon_status 
			USING status::text::coupon_status
		`)
		if err != nil {
			return fmt.Errorf("failed to alter column type: %w", err)
		}

		// Drop the old enum type
		_, err = tx.Exec(`DROP TYPE coupon_status_old`)
		if err != nil {
			return fmt.Errorf("failed to drop old enum type: %w", err)
		}

		// Add new columns for enhanced coupon tracking
		_, err = tx.Exec(`
			ALTER TABLE coupons 
			ADD COLUMN IF NOT EXISTS preview_image_url TEXT,
			ADD COLUMN IF NOT EXISTS selected_preview_id TEXT,
			ADD COLUMN IF NOT EXISTS stones_count INTEGER,
			ADD COLUMN IF NOT EXISTS final_schema_url TEXT,
			ADD COLUMN IF NOT EXISTS page_count INTEGER DEFAULT 0
		`)
		if err != nil {
			return fmt.Errorf("failed to add new columns: %w", err)
		}

		// Commit transaction
		return tx.Commit()
	}, func(ctx context.Context, db *bun.DB) error {
		// Rollback migration
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Drop new columns
		_, err = tx.Exec(`
			ALTER TABLE coupons 
			DROP COLUMN IF EXISTS preview_image_url,
			DROP COLUMN IF EXISTS selected_preview_id,
			DROP COLUMN IF EXISTS stones_count,
			DROP COLUMN IF EXISTS final_schema_url,
			DROP COLUMN IF EXISTS page_count
		`)
		if err != nil {
			return fmt.Errorf("failed to drop new columns: %w", err)
		}

		// Rename current enum type
		_, err = tx.Exec(`ALTER TYPE coupon_status RENAME TO coupon_status_new`)
		if err != nil {
			return fmt.Errorf("failed to rename new enum type: %w", err)
		}

		// Recreate old enum type
		_, err = tx.Exec(`CREATE TYPE coupon_status AS ENUM ('new', 'activated', 'used', 'completed')`)
		if err != nil {
			return fmt.Errorf("failed to recreate old enum type: %w", err)
		}

		// Update column to use old enum type
		_, err = tx.Exec(`
			ALTER TABLE coupons 
			ALTER COLUMN status TYPE coupon_status 
			USING status::text::coupon_status
		`)
		if err != nil {
			return fmt.Errorf("failed to revert column type: %w", err)
		}

		// Drop the new enum type
		_, err = tx.Exec(`DROP TYPE coupon_status_new`)
		if err != nil {
			return fmt.Errorf("failed to drop new enum type: %w", err)
		}

		return tx.Commit()
	})
}