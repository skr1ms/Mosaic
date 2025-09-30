package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockAdminRepository struct {
	mock.Mock
}

func (m *MockAdminRepository) Create(admin *Admin) error {
	args := m.Called(admin)
	return args.Error(0)
}

func (m *MockAdminRepository) GetByID(id uuid.UUID) (*Admin, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Admin), args.Error(1)
}

func (m *MockAdminRepository) GetByLogin(login string) (*Admin, error) {
	args := m.Called(login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Admin), args.Error(1)
}

func (m *MockAdminRepository) GetByEmail(email string) (*Admin, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Admin), args.Error(1)
}

func (m *MockAdminRepository) Update(admin *Admin) error {
	args := m.Called(admin)
	return args.Error(0)
}

func (m *MockAdminRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAdminRepository) GetAll() ([]*Admin, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Admin), args.Error(1)
}

func (m *MockAdminRepository) UpdateLastLogin(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAdminRepository) CreateProfileChangeLog(log *ProfileChangeLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockAdminRepository) GetProfileChangesByPartnerID(partnerID uuid.UUID) ([]*ProfileChangeLog, error) {
	args := m.Called(partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ProfileChangeLog), args.Error(1)
}

func (m *MockAdminRepository) UpdateEmail(id uuid.UUID, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}

func (m *MockAdminRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	args := m.Called(id, hashedPassword)
	return args.Error(0)
}

// Mock PartnerRepository
type MockPartnerRepository struct {
	mock.Mock
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error) {
	args := m.Called(ctx, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByLogin(ctx context.Context, login string) (*partner.Partner, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByDomain(ctx context.Context, domain string) (*partner.Partner, error) {
	args := m.Called(ctx, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetByPartnerCode(ctx context.Context, code string) (*partner.Partner, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetNextPartnerCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockPartnerRepository) Create(ctx context.Context, p *partner.Partner) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPartnerRepository) Update(ctx context.Context, p *partner.Partner) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}

func (m *MockPartnerRepository) DeleteWithCoupons(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockPartnerRepository) Search(ctx context.Context, search, status, sortBy, order string) ([]*partner.Partner, error) {
	args := m.Called(ctx, search, status, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) GetActivePartners(ctx context.Context) ([]*partner.Partner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

func (m *MockPartnerRepository) InitializeArticleGrid(ctx context.Context, partnerID uuid.UUID) error {
	args := m.Called(ctx, partnerID)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetArticleGrid(ctx context.Context, partnerID uuid.UUID) (map[string]map[string]map[string]string, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]map[string]map[string]string), args.Error(1)
}

func (m *MockPartnerRepository) UpdateArticleSKU(ctx context.Context, partnerID uuid.UUID, size, style, marketplace, sku string) error {
	args := m.Called(ctx, partnerID, size, style, marketplace, sku)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetArticleBySizeStyle(ctx context.Context, partnerID uuid.UUID, size, style, marketplace string) (*partner.PartnerArticle, error) {
	args := m.Called(ctx, partnerID, size, style, marketplace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*partner.PartnerArticle), args.Error(1)
}

func (m *MockPartnerRepository) GetAllArticlesByPartner(ctx context.Context, partnerID uuid.UUID) ([]*partner.PartnerArticle, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.PartnerArticle), args.Error(1)
}

func (m *MockPartnerRepository) DeleteArticleGrid(ctx context.Context, partnerID uuid.UUID) error {
	args := m.Called(ctx, partnerID)
	return args.Error(0)
}

// Mock CouponRepository
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) Create(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) CreateBatch(ctx context.Context, coupons []*coupon.Coupon) error {
	args := m.Called(ctx, coupons)
	return args.Error(0)
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*coupon.Coupon, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetAll(ctx context.Context) ([]*coupon.Coupon, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetRecentActivated(ctx context.Context, limit int) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) BatchDelete(ctx context.Context, ids []uuid.UUID) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) Search(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, code, status, size, style, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) SearchWithPagination(ctx context.Context, code, status, size, style string, partnerID *uuid.UUID, page, limit int, createdFrom, createdTo, usedFrom, usedTo *time.Time, sortBy, sortDir string) ([]*coupon.Coupon, int, error) {
	args := m.Called(ctx, code, status, size, style, partnerID, page, limit, createdFrom, createdTo, usedFrom, usedTo, sortBy, sortDir)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*coupon.Coupon), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) UpdateStatusByPartnerID(ctx context.Context, partnerID uuid.UUID, blocked bool) error {
	args := m.Called(ctx, partnerID, blocked)
	return args.Error(0)
}

func (m *MockCouponRepository) GetFiltered(ctx context.Context, filters map[string]any) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) CountByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountPurchasedByPartnerID(ctx context.Context, partnerID uuid.UUID) (int, error) {
	args := m.Called(ctx, partnerID)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountTotal(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountByPartnerAndStatus(ctx context.Context, partnerID uuid.UUID, status string) (int64, error) {
	args := m.Called(ctx, partnerID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivated(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountBrandedPurchasesByPartner(ctx context.Context, partnerID uuid.UUID) (int64, error) {
	args := m.Called(ctx, partnerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedInTimeRange(ctx context.Context, from, to time.Time) (int64, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCouponRepository) GetStatistics(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerSalesStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerUsageStatistics(ctx context.Context, partnerID uuid.UUID) (map[string]any, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockCouponRepository) GetStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetExtendedStatusCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) FindAvailableCoupon(ctx context.Context, size, style string, partnerID *uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, size, style, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetSizeCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetStyleCounts(ctx context.Context, partnerID *uuid.UUID) (map[string]int64, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockCouponRepository) GetTopActivatedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coupon.PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) GetTopPurchasedByPartner(ctx context.Context, limit int) ([]coupon.PartnerCount, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]coupon.PartnerCount), args.Error(1)
}

func (m *MockCouponRepository) GetLastActivityByPartner(ctx context.Context, partnerID uuid.UUID) (*time.Time, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*time.Time), args.Error(1)
}

func (m *MockCouponRepository) GetTimeSeriesData(ctx context.Context, from, to time.Time, period string, partnerID *uuid.UUID) ([]map[string]any, error) {
	args := m.Called(ctx, from, to, period, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]map[string]any), args.Error(1)
}

func (m *MockCouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockCouponRepository) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCouponRepository) GetCouponsWithAdvancedFilter(ctx context.Context, filter coupon.CouponFilterRequest) ([]*coupon.CouponInfo, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*coupon.CouponInfo), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetCouponsForDeletion(ctx context.Context, ids []uuid.UUID) ([]*coupon.CouponDeletePreview, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.CouponDeletePreview), args.Error(1)
}

func (m *MockCouponRepository) GetCouponsForExport(ctx context.Context, options coupon.ExportOptionsRequest) (any, error) {
	args := m.Called(ctx, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}

func (m *MockCouponRepository) CountCreatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	args := m.Called(ctx, partnerID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountActivatedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	args := m.Called(ctx, partnerID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) CountPurchasedByPartnerInRange(ctx context.Context, partnerID uuid.UUID, from, to *time.Time) (int, error) {
	args := m.Called(ctx, partnerID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockCouponRepository) ActivateCoupon(ctx context.Context, id uuid.UUID, req coupon.ActivateCouponRequest) error {
	args := m.Called(ctx, id, req)
	return args.Error(0)
}

func (m *MockCouponRepository) MarkAsPurchased(ctx context.Context, id uuid.UUID, purchaseEmail string) error {
	args := m.Called(ctx, id, purchaseEmail)
	return args.Error(0)
}

func (m *MockCouponRepository) SendSchema(ctx context.Context, id uuid.UUID, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockCouponRepository) ResetCoupon(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) Reset(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) BatchReset(ctx context.Context, ids []uuid.UUID) ([]uuid.UUID, []uuid.UUID, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]uuid.UUID), args.Get(1).([]uuid.UUID), args.Error(2)
}

func (m *MockCouponRepository) GetPartnerCouponsWithFilter(ctx context.Context, partnerID uuid.UUID, filters map[string]any, page, limit int, sortBy, order string) ([]*coupon.Coupon, int, error) {
	args := m.Called(ctx, partnerID, filters, page, limit, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*coupon.Coupon), args.Int(1), args.Error(2)
}

func (m *MockCouponRepository) GetPartnerCouponByCode(ctx context.Context, partnerID uuid.UUID, code string) (*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerCouponDetail(ctx context.Context, partnerID uuid.UUID, couponID uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetPartnerRecentActivity(ctx context.Context, partnerID uuid.UUID, limit int) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

// Mock ImageRepository
type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) GetAll(ctx context.Context) ([]*image.Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*image.Image), args.Error(1)
}

func (m *MockImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) GetByStatus(ctx context.Context, status string) ([]*image.Image, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*image.Image), args.Error(1)
}

func (m *MockImageRepository) Update(ctx context.Context, img *image.Image) error {
	args := m.Called(ctx, img)
	return args.Error(0)
}

func (m *MockImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock S3Client
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) DownloadFile(ctx context.Context, key string) (any, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockS3Client) UploadLogo(ctx context.Context, reader any, size int64, contentType string, partnerID string) (string, error) {
	args := m.Called(ctx, reader, size, contentType, partnerID)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) GetLogoURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, objectKey, expiry)
	return args.String(0), args.Error(1)
}

// Mock RedisClient
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) LLen(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func TestAdminService_CreateAdmin(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateAdminRequest
		mockSetup     func(*MockAdminRepository)
		expectedError bool
	}{
		{
			name: "successful_creation",
			request: CreateAdminRequest{
				Login:    "admin@example.com",
				Email:    "admin@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *MockAdminRepository) {
				repo.On("GetByLogin", "admin@example.com").Return(nil, errors.New("not found"))
				repo.On("GetByEmail", "admin@example.com").Return(nil, errors.New("not found"))
				repo.On("Create", mock.AnythingOfType("*admin.Admin")).Return(nil)
			},
			expectedError: false,
		},
		{
			name: "admin_already_exists",
			request: CreateAdminRequest{
				Login:    "existing@example.com",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *MockAdminRepository) {
				existingAdmin := &Admin{
					ID:    uuid.New(),
					Login: "existing@example.com",
				}
				repo.On("GetByLogin", "existing@example.com").Return(existingAdmin, nil)
			},
			expectedError: true,
		},
		{
			name: "missing_login",
			request: CreateAdminRequest{
				Email:    "admin@example.com",
				Password: "password123",
			},
			mockSetup: func(repo *MockAdminRepository) {
				repo.On("GetByLogin", "").Return(nil, errors.New("not found"))
				repo.On("GetByEmail", "admin@example.com").Return(nil, errors.New("not found"))
				repo.On("Create", mock.AnythingOfType("*admin.Admin")).Return(errors.New("validation error: login required"))
			},
			expectedError: true,
		},
		{
			name: "missing_password",
			request: CreateAdminRequest{
				Login: "admin@example.com",
				Email: "admin@example.com",
			},
			mockSetup: func(repo *MockAdminRepository) {
				repo.On("GetByLogin", "admin@example.com").Return(nil, errors.New("not found"))
				repo.On("GetByEmail", "admin@example.com").Return(nil, errors.New("not found"))
				repo.On("Create", mock.AnythingOfType("*admin.Admin")).Return(errors.New("validation error: password required"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockAdminRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			deps := &AdminServiceDeps{
				AdminRepository: mockRepo,
			}
			service := NewAdminService(deps)

			admin, err := service.CreateAdmin(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, admin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, admin)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetDashboardData(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository, *MockImageRepository)
		expectedError bool
	}{
		{
			name: "successful_get_dashboard",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, imageRepo *MockImageRepository) {
				coupons := []*coupon.Coupon{
					{ID: uuid.New(), Status: "active"},
					{ID: uuid.New(), Status: "used"},
					{ID: uuid.New(), Status: "purchased"},
				}
				couponRepo.On("GetAll", mock.Anything).Return(coupons, nil)
				couponRepo.On("GetRecentActivated", mock.Anything, 10).Return([]*coupon.Coupon{}, nil)

				partners := []*partner.Partner{
					{ID: uuid.New(), Login: "partner1@example.com"},
					{ID: uuid.New(), Login: "partner2@example.com"},
				}
				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

				images := []*image.Image{
					{ID: uuid.New(), Status: "processing"},
					{ID: uuid.New(), Status: "completed"},
					{ID: uuid.New(), Status: "failed"},
				}
				imageRepo.On("GetAll", mock.Anything).Return(images, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockImageRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo, mockImageRepo)
			}

			deps := &AdminServiceDeps{
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
				ImageRepository:   mockImageRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetDashboardData()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetAnalytics(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository)
		expectedError bool
	}{
		{
			name: "get_analytics",
			mockSetup: func(couponRepo *MockCouponRepository) {
				coupons := []*coupon.Coupon{
					{ID: uuid.New(), Status: "active", Size: "30x40", Style: "default"},
					{ID: uuid.New(), Status: "used", Size: "40x50", Style: "contrast"},
					{ID: uuid.New(), Status: "purchased", Size: "50x70", Style: "soft"},
				}
				couponRepo.On("GetAll", mock.Anything).Return(coupons, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo)
			}

			deps := &AdminServiceDeps{
				CouponRepository: mockCouponRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetAnalytics()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCouponRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetPartners(t *testing.T) {
	tests := []struct {
		name          string
		search        string
		status        string
		sortBy        string
		order         string
		mockSetup     func(*MockPartnerRepository)
		expectedError bool
	}{
		{
			name:   "successful_get_partners",
			search: "",
			status: "",
			sortBy: "created_at",
			order:  "desc",
			mockSetup: func(repo *MockPartnerRepository) {
				partners := []*partner.Partner{
					{
						ID:    uuid.New(),
						Login: "partner1@example.com",
					},
					{
						ID:    uuid.New(),
						Login: "partner2@example.com",
					},
				}
				repo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)
			},
			expectedError: false,
		},
		{
			name:   "repository_error",
			search: "",
			status: "",
			sortBy: "created_at",
			order:  "desc",
			mockSetup: func(repo *MockPartnerRepository) {
				repo.On("GetAll", mock.Anything, "created_at", "desc").Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo)
			}

			deps := &AdminServiceDeps{
				PartnerRepository: mockPartnerRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetPartners(tt.search, tt.status, tt.sortBy, tt.order)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockPartnerRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetPartnerDetail(t *testing.T) {
	partnerID := uuid.New()

	tests := []struct {
		name          string
		partnerID     uuid.UUID
		mockSetup     func(*MockPartnerRepository, *MockCouponRepository, *MockAdminRepository)
		expectedError bool
	}{
		{
			name:      "successful_get_partner_detail",
			partnerID: partnerID,
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository, adminRepo *MockAdminRepository) {
				partner := &partner.Partner{
					ID:          partnerID,
					Login:       "partner@example.com",
					BrandName:   "Test Brand",
					PartnerCode: "TEST001",
				}
				partnerRepo.On("GetByID", mock.Anything, partnerID).Return(partner, nil)
				couponRepo.On("CountByPartnerID", mock.Anything, partnerID).Return(10, nil)
				couponRepo.On("CountActivatedByPartnerID", mock.Anything, partnerID).Return(5, nil)

				lastActivity := time.Now()
				couponRepo.On("GetLastActivityByPartner", mock.Anything, partnerID).Return(&lastActivity, nil)

				adminRepo.On("GetProfileChangesByPartnerID", partnerID).Return([]*ProfileChangeLog{}, nil)
			},
			expectedError: false,
		},
		{
			name:      "partner_not_found",
			partnerID: partnerID,
			mockSetup: func(partnerRepo *MockPartnerRepository, couponRepo *MockCouponRepository, adminRepo *MockAdminRepository) {
				partnerRepo.On("GetByID", mock.Anything, partnerID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPartnerRepo := new(MockPartnerRepository)
			mockCouponRepo := new(MockCouponRepository)
			mockAdminRepo := new(MockAdminRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockPartnerRepo, mockCouponRepo, mockAdminRepo)
			}

			deps := &AdminServiceDeps{
				PartnerRepository: mockPartnerRepo,
				CouponRepository:  mockCouponRepo,
				AdminRepository:   mockAdminRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetPartnerDetail(tt.partnerID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.partnerID, result.ID)
			}

			mockPartnerRepo.AssertExpectations(t)
			mockCouponRepo.AssertExpectations(t)
			mockAdminRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetStatistics(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository)
		expectedError bool
	}{
		{
			name: "successful_get_statistics",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository) {
				stats := map[string]int64{
					"total":     100,
					"active":    80,
					"used":      60,
					"purchased": 70,
				}
				couponRepo.On("GetStatistics", mock.Anything, (*uuid.UUID)(nil)).Return(stats, nil)
			},
			expectedError: false,
		},
		{
			name: "repository_error",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository) {
				couponRepo.On("GetStatistics", mock.Anything, (*uuid.UUID)(nil)).Return(nil, errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo)
			}

			deps := &AdminServiceDeps{
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetStatistics()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetSystemStatistics(t *testing.T) {
	tests := []struct {
		name          string
		mockSetup     func(*MockCouponRepository, *MockPartnerRepository, *MockImageRepository)
		expectedError bool
	}{
		{
			name: "successful_get_system_statistics",
			mockSetup: func(couponRepo *MockCouponRepository, partnerRepo *MockPartnerRepository, imageRepo *MockImageRepository) {
				stats := map[string]int64{
					"total":     100,
					"active":    80,
					"used":      60,
					"purchased": 70,
				}
				couponRepo.On("GetStatistics", mock.Anything, (*uuid.UUID)(nil)).Return(stats, nil)

				partnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return([]*partner.Partner{}, nil)
				imageRepo.On("GetAll", mock.Anything).Return([]*image.Image{}, nil)
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)
			mockPartnerRepo := new(MockPartnerRepository)
			mockImageRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockPartnerRepo, mockImageRepo)
			}

			deps := &AdminServiceDeps{
				CouponRepository:  mockCouponRepo,
				PartnerRepository: mockPartnerRepo,
				ImageRepository:   mockImageRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetSystemStatistics()

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCouponRepo.AssertExpectations(t)
			mockPartnerRepo.AssertExpectations(t)
			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_GetImageDetails(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		imageID       uuid.UUID
		mockSetup     func(*MockImageRepository)
		expectedError bool
	}{
		{
			name:    "successful_get_image_details",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				img := &image.Image{
					ID:     imageID,
					Status: "completed",
				}
				repo.On("GetByID", mock.Anything, imageID).Return(img, nil)
			},
			expectedError: false,
		},
		{
			name:    "image_not_found",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				repo.On("GetByID", mock.Anything, imageID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockImageRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockImageRepo)
			}

			deps := &AdminServiceDeps{
				ImageRepository: mockImageRepo,
			}
			service := NewAdminService(deps)

			result, err := service.GetImageDetails(tt.imageID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.imageID, result.ID)
			}

			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_DeleteImageTask(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		imageID       uuid.UUID
		mockSetup     func(*MockImageRepository)
		expectedError bool
	}{
		{
			name:    "successful_delete_image_task",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				repo.On("Delete", mock.Anything, imageID).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "repository_error",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				repo.On("Delete", mock.Anything, imageID).Return(errors.New("database error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockImageRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockImageRepo)
			}

			deps := &AdminServiceDeps{
				ImageRepository: mockImageRepo,
			}
			service := NewAdminService(deps)

			err := service.DeleteImageTask(tt.imageID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_RetryImageTask(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name          string
		imageID       uuid.UUID
		mockSetup     func(*MockImageRepository)
		expectedError bool
	}{
		{
			name:    "successful_retry_image_task",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				img := &image.Image{
					ID:     imageID,
					Status: "failed",
				}
				repo.On("GetByID", mock.Anything, imageID).Return(img, nil)
				repo.On("Update", mock.Anything, mock.MatchedBy(func(img *image.Image) bool {
					return img.ID == imageID && img.Status == "queued"
				})).Return(nil)
			},
			expectedError: false,
		},
		{
			name:    "image_not_found",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository) {
				repo.On("GetByID", mock.Anything, imageID).Return(nil, errors.New("not found"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockImageRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockImageRepo)
			}

			deps := &AdminServiceDeps{
				ImageRepository: mockImageRepo,
			}
			service := NewAdminService(deps)

			err := service.RetryImageTask(tt.imageID)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockImageRepo.AssertExpectations(t)
		})
	}
}

func TestAdminService_BatchResetCoupons(t *testing.T) {
	couponID1 := uuid.New()
	couponID2 := uuid.New()

	tests := []struct {
		name          string
		couponIDs     []string
		mockSetup     func(*MockCouponRepository, *MockRedisClient)
		expectedError bool
	}{
		{
			name:      "successful_batch_reset",
			couponIDs: []string{couponID1.String(), couponID2.String()},
			mockSetup: func(couponRepo *MockCouponRepository, redisClient *MockRedisClient) {
				couponRepo.On("BatchReset", mock.Anything, []uuid.UUID{couponID1, couponID2}).Return([]uuid.UUID{couponID1, couponID2}, []uuid.UUID{}, nil)
			},
			expectedError: false,
		},
		{
			name:          "invalid_uuid",
			couponIDs:     []string{"invalid-uuid"},
			mockSetup:     func(couponRepo *MockCouponRepository, redisClient *MockRedisClient) {},
			expectedError: false, // Invalid UUID doesn't cause an error, but gets added to Failed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCouponRepo := new(MockCouponRepository)
			mockRedisClient := new(MockRedisClient)

			if tt.mockSetup != nil {
				tt.mockSetup(mockCouponRepo, mockRedisClient)
			}

			deps := &AdminServiceDeps{
				CouponRepository: mockCouponRepo,
				RedisClient:      mockRedisClient,
			}
			service := NewAdminService(deps)

			result, err := service.BatchResetCoupons(tt.couponIDs)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockCouponRepo.AssertExpectations(t)
			mockRedisClient.AssertExpectations(t)
		})
	}
}
