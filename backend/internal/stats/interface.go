package stats

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/pkg/jwt"
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
	GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error)
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
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Ping(ctx context.Context) *redis.StatusCmd
	LLen(ctx context.Context, key string) *redis.IntCmd
}

type StatsServiceInterface interface {
	GetGeneralStats(ctx context.Context) (*GeneralStatsResponse, error)
	GetPartnerStats(ctx context.Context, partnerID uuid.UUID) (*PartnerStatsResponse, error)
	GetAllPartnersStats(ctx context.Context) (*PartnerListStatsResponse, error)
	GetTimeSeriesStats(ctx context.Context, filters *StatsFiltersRequest) (*TimeSeriesStatsResponse, error)
	GetSystemHealth(ctx context.Context) (*SystemHealthResponse, error)
	GetCouponsByStatus(ctx context.Context, partnerID *uuid.UUID) (*CouponsByStatusResponse, error)
	GetCouponsBySize(ctx context.Context, partnerID *uuid.UUID) (*CouponsBySizeResponse, error)
	GetCouponsByStyle(ctx context.Context, partnerID *uuid.UUID) (*CouponsByStyleResponse, error)
	GetTopPartners(ctx context.Context, limit int, sortBy ...string) (*TopPartnersResponse, error)
	GetRealTimeStats(ctx context.Context) (*RealTimeStatsResponse, error)
}

type MetricsCollectorInterface interface {
	IncrementCouponsCreated(partnerID, size, style string)
	IncrementCouponsActivated(partnerID, size, style string)
	IncrementCouponsPurchased(partnerID string)
	ObserveImageProcessingDuration(operationType, status string, duration float64)
	SetImageProcessingQueueSize(size float64)
	SetPartnersCount(total, active float64)
	IncrementHTTPRequests(method, endpoint, status string)
	ObserveHTTPRequestDuration(method, endpoint string, duration float64)
	SetDatabaseConnections(count float64)
	SetRedisConnections(count float64)
	IncrementErrors(errorType, component string)
	SetActiveUsers(count float64)
	SetSystemMetrics(memoryBytes, cpuPercent float64)
}

type JWTServiceInterface interface {
	CreateAccessToken(userID uuid.UUID, login, role string) (string, error)
	CreateRefreshToken(userID uuid.UUID, login, role string) (string, error)
	ValidateAccessToken(tokenString string) (*jwt.Claims, error)
	ValidateRefreshToken(tokenString string) (*jwt.Claims, error)
	RefreshTokens(refreshTokenString string) (*jwt.TokenPair, error)
}
