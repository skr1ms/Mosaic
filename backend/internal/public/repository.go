package public

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PublicRepository struct {
	db *bun.DB
}

func NewPublicRepository(db *bun.DB) *PublicRepository {
	return &PublicRepository{db: db}
}

func (r *PublicRepository) Create(ctx context.Context, preview *PreviewData) error {
	// Set TTL to 24 hours from now for previews
	if preview.ExpiresAt == nil {
		expiresAt := time.Now().Add(24 * time.Hour)
		preview.ExpiresAt = &expiresAt
	}

	_, err := r.db.NewInsert().Model(preview).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create preview: %w", err)
	}

	return nil
}

func (r *PublicRepository) GetByID(ctx context.Context, id uuid.UUID) (*PreviewData, error) {
	preview := &PreviewData{}
	err := r.db.NewSelect().
		Model(preview).
		Where("id = ? AND (expires_at IS NULL OR expires_at > NOW())", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get preview: %w", err)
	}

	return preview, nil
}

func (r *PublicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.NewDelete().
		Model((*PreviewData)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete preview: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("preview not found")
	}

	return nil
}

func (r *PublicRepository) GetByImageID(ctx context.Context, imageID uuid.UUID) ([]*PreviewData, error) {
	var previews []*PreviewData
	err := r.db.NewSelect().
		Model(&previews).
		Where("image_id = ? AND (expires_at IS NULL OR expires_at > NOW())", imageID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get previews by image ID: %w", err)
	}

	return previews, nil
}

func (r *PublicRepository) GetByStyle(ctx context.Context, style, size string) ([]*PreviewData, error) {
	var previews []*PreviewData
	err := r.db.NewSelect().
		Model(&previews).
		Where("style = ? AND size = ? AND (expires_at IS NULL OR expires_at > NOW())", style, size).
		Order("created_at DESC").
		Limit(10). // Limit to avoid too many results
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get previews by style: %w", err)
	}

	return previews, nil
}

func (r *PublicRepository) CleanupExpired(ctx context.Context) error {
	result, err := r.db.NewDelete().
		Model((*PreviewData)(nil)).
		Where("expires_at < NOW()").
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to cleanup expired previews: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		// Log cleanup activity if needed
		fmt.Printf("Cleaned up %d expired previews\n", rowsAffected)
	}

	return nil
}

func (r *PublicRepository) SetTTL(ctx context.Context, id uuid.UUID, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)

	_, err := r.db.NewUpdate().
		Model((*PreviewData)(nil)).
		Set("expires_at = ?, updated_at = NOW()", expiresAt).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to set TTL for preview: %w", err)
	}

	return nil
}
