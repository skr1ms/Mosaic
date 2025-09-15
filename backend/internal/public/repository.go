package public

import (
	"context"
	"database/sql"
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
	query := `
		INSERT INTO previews (id, url, style, contrast, size, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	if preview.ID == "" {
		preview.ID = uuid.New().String()
	}

	createdAt := time.Now()
	expiresAt := createdAt.Add(24 * time.Hour) // Previews expire after 24 hours

	_, err := r.db.ExecContext(ctx, query,
		preview.ID,
		preview.URL,
		preview.Style,
		preview.Contrast,
		preview.Size,
		createdAt,
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create preview: %w", err)
	}

	return nil
}

func (r *PublicRepository) GetByID(ctx context.Context, id string) (*PreviewData, error) {
	query := `
		SELECT id, url, style, contrast, size
		FROM previews
		WHERE id = $1 AND expires_at > NOW()
	`

	var preview PreviewData
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&preview.ID,
		&preview.URL,
		&preview.Style,
		&preview.Contrast,
		&preview.Size,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("preview not found")
		}
		return nil, fmt.Errorf("failed to get preview: %w", err)
	}

	return &preview, nil
}

func (r *PublicRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM previews WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
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

func (r *PublicRepository) GetByUserSession(ctx context.Context, sessionID string) ([]*PreviewData, error) {
	query := `
		SELECT id, url, style, contrast, size
		FROM previews
		WHERE session_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get previews: %w", err)
	}
	defer rows.Close()

	var previews []*PreviewData
	for rows.Next() {
		var preview PreviewData
		err := rows.Scan(
			&preview.ID,
			&preview.URL,
			&preview.Style,
			&preview.Contrast,
			&preview.Size,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan preview: %w", err)
		}
		previews = append(previews, &preview)
	}

	return previews, nil
}

func (r *PublicRepository) CleanupExpired(ctx context.Context) error {
	query := `DELETE FROM previews WHERE expires_at < NOW()`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired previews: %w", err)
	}

	return nil
}
