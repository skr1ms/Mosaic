package stats

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/internal/partner"
)

type CouponRepositoryInterface interface {
	CountTotal(ctx context.Context) (int64, error)
	CountActivated(ctx context.Context) (int64, error)
	CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error)
	GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error)
	CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error)
	GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetSizeCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetStyleCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error)
	GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]interface{}, error)
	HealthCheck(ctx context.Context) error
}

type PartnerRepositoryInterface interface {
	CountActive(ctx context.Context) (int64, error)
	CountTotal(ctx context.Context) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error)
}

type RedisClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Ping(ctx context.Context) *redis.StatusCmd
	LLen(ctx context.Context, key string) *redis.IntCmd
}
