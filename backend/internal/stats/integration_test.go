package stats

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/internal/partner"
)

type IntegrationTestSuite struct {
	app      *fiber.App
	handler  *StatsHandler
	service  *StatsService
	mocks    *IntegrationMocks
	testData *IntegrationTestData
}

type IntegrationMocks struct {
	CouponRepo  *MockCouponRepository
	PartnerRepo *MockPartnerRepository
	RedisClient *MockRedisClient
}

type IntegrationTestData struct {
	Partner *partner.Partner
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		mocks: &IntegrationMocks{
			CouponRepo:  &MockCouponRepository{},
			PartnerRepo: &MockPartnerRepository{},
			RedisClient: &MockRedisClient{},
		},
		testData: &IntegrationTestData{
			Partner: createTestPartner(),
		},
	}

	deps := &StatsServiceDeps{
		CouponRepository:  suite.mocks.CouponRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		RedisClient:       suite.mocks.RedisClient,
	}
	suite.service = NewStatsService(deps)

	handlerDeps := &StatsHandlerDeps{
		StatsService: suite.service,
	}
	suite.handler = &StatsHandler{
		Router: suite.app,
		deps:   handlerDeps,
	}

	adminStats := suite.app.Group("/admin/stats")
	adminStats.Get("/general", suite.handler.GetGeneralStats)
	adminStats.Get("/partners/:partner_id", suite.handler.GetPartnerStats)
	adminStats.Get("/partners", suite.handler.GetAllPartnersStats)
	adminStats.Get("/time-series", suite.handler.GetTimeSeriesStats)
	adminStats.Get("/system-health", suite.handler.GetSystemHealth)
	adminStats.Get("/coupons-by-status", suite.handler.GetCouponsByStatus)
	adminStats.Get("/coupons-by-size", suite.handler.GetCouponsBySizes)
	adminStats.Get("/coupons-by-style", suite.handler.GetCouponsByStyles)
	adminStats.Get("/top-partners", suite.handler.GetTopPartners)

	return suite
}

func TestIntegration_GetGeneralStats(t *testing.T) {
	suite := SetupIntegrationTest(t)

	suite.mocks.RedisClient.On("Get", mock.Anything, "general_stats").Return("", errors.New("not found"))

	suite.mocks.CouponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
	suite.mocks.CouponRepo.On("CountActivated", mock.Anything).Return(int64(750), nil)
	suite.mocks.PartnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
	suite.mocks.PartnerRepo.On("CountTotal", mock.Anything).Return(int64(60), nil)

	suite.mocks.RedisClient.On("Set", mock.Anything, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/general", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result GeneralStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(1000), result.TotalCouponsCreated)
	assert.Equal(t, int64(750), result.TotalCouponsActivated)
	assert.Equal(t, float64(75), result.ActivationRate)
	assert.Equal(t, int64(50), result.ActivePartnersCount)
	assert.Equal(t, int64(60), result.TotalPartnersCount)

	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.RedisClient.AssertExpectations(t)
}

func TestIntegration_GetPartnerStats(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := suite.testData.Partner
	cacheKey := "partner_stats:" + testPartner.ID.String()

	suite.mocks.RedisClient.On("Get", mock.Anything, cacheKey).Return("", errors.New("not found"))
	suite.mocks.PartnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)
	suite.mocks.CouponRepo.On("CountByPartner", mock.Anything, testPartner.ID).Return(int64(100), nil)
	suite.mocks.CouponRepo.On("CountActivatedByPartner", mock.Anything, testPartner.ID).Return(int64(80), nil)
	suite.mocks.CouponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, testPartner.ID).Return(int64(20), nil)

	lastActivity := time.Now()
	suite.mocks.CouponRepo.On("GetLastActivityByPartner", mock.Anything, testPartner.ID).Return(&lastActivity, nil)
	suite.mocks.RedisClient.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/partners/"+testPartner.ID.String(), nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result PartnerStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, testPartner.ID, result.PartnerID)
	assert.Equal(t, testPartner.BrandName, result.PartnerName)
	assert.Equal(t, int64(100), result.TotalCoupons)
	assert.Equal(t, int64(80), result.ActivatedCoupons)
	assert.Equal(t, int64(20), result.UnusedCoupons)
	assert.Equal(t, float64(80), result.ActivationRate)

	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.RedisClient.AssertExpectations(t)
}

func TestIntegration_GetPartnerStats_InvalidID(t *testing.T) {
	suite := SetupIntegrationTest(t)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/partners/invalid-uuid", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_GetAllPartnersStats(t *testing.T) {
	suite := SetupIntegrationTest(t)

	partners := []*partner.Partner{createTestPartner(), createTestPartner()}
	suite.mocks.PartnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

	for _, p := range partners {
		cacheKey := "partner_stats:" + p.ID.String()
		suite.mocks.RedisClient.On("Get", mock.Anything, cacheKey).Return("", errors.New("not found"))
		suite.mocks.PartnerRepo.On("GetByID", mock.Anything, p.ID).Return(p, nil)
		suite.mocks.CouponRepo.On("CountByPartner", mock.Anything, p.ID).Return(int64(50), nil)
		suite.mocks.CouponRepo.On("CountActivatedByPartner", mock.Anything, p.ID).Return(int64(40), nil)
		suite.mocks.CouponRepo.On("CountBrandedPurchasesByPartner", mock.Anything, p.ID).Return(int64(10), nil)
		suite.mocks.CouponRepo.On("GetLastActivityByPartner", mock.Anything, p.ID).Return(nil, nil)
		suite.mocks.RedisClient.On("Set", mock.Anything, cacheKey, mock.AnythingOfType("[]uint8"), 10*time.Minute).Return(nil)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/partners", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result PartnerListStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Partners, 2)
	assert.Equal(t, int64(2), result.Total)

	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.RedisClient.AssertExpectations(t)
}

func TestIntegration_GetTimeSeriesStats(t *testing.T) {
	suite := SetupIntegrationTest(t)

	rawData := []map[string]interface{}{
		{
			"date":               "2023-01-01",
			"coupons_created":    int64(10),
			"coupons_activated":  int64(8),
			"coupons_purchased":  int64(5),
			"new_partners_count": int64(1),
		},
	}

	suite.mocks.CouponRepo.On("GetTimeSeriesData", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), "day", (*uuid.UUID)(nil)).Return(rawData, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/time-series?period=day", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TimeSeriesStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "day", result.Period)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "2023-01-01", result.Data[0].Date)
	assert.Equal(t, int64(10), result.Data[0].CouponsCreated)

	suite.mocks.CouponRepo.AssertExpectations(t)
}

func TestIntegration_GetSystemHealth(t *testing.T) {
	suite := SetupIntegrationTest(t)

	suite.mocks.RedisClient.On("Get", mock.Anything, "system_health").Return("", errors.New("not found"))
	suite.mocks.CouponRepo.On("HealthCheck", mock.Anything).Return(nil)
	suite.mocks.RedisClient.On("Ping", mock.Anything).Return(nil)
	suite.mocks.RedisClient.On("LLen", mock.Anything, "image_processing_queue").Return(int64(5), nil)
	suite.mocks.RedisClient.On("Set", mock.Anything, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/system-health", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result SystemHealthResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "healthy", result.Status)
	assert.Equal(t, "healthy", result.DatabaseStatus)
	assert.Equal(t, "healthy", result.RedisStatus)
	assert.Equal(t, int64(5), result.ImageProcessingQueue)

	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.RedisClient.AssertExpectations(t)
}

func TestIntegration_GetCouponsByStatus(t *testing.T) {
	suite := SetupIntegrationTest(t)

	statusCounts := map[string]int64{
		"new":       100,
		"activated": 80,
		"used":      60,
		"completed": 40,
	}

	suite.mocks.CouponRepo.On("GetExtendedStatusCounts", mock.Anything, (*uuid.UUID)(nil)).Return(statusCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/coupons-by-status", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result CouponsByStatusResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(100), result.New)
	assert.Equal(t, int64(80), result.Activated)
	assert.Equal(t, int64(60), result.Used)
	assert.Equal(t, int64(40), result.Completed)

	suite.mocks.CouponRepo.AssertExpectations(t)
}

func TestIntegration_GetCouponsBySize(t *testing.T) {
	suite := SetupIntegrationTest(t)

	sizeCounts := map[string]int64{
		"21x30": 10,
		"30x40": 20,
		"40x40": 30,
		"40x50": 40,
		"40x60": 50,
		"50x70": 60,
	}

	suite.mocks.CouponRepo.On("GetSizeCounts", mock.Anything, (*uuid.UUID)(nil)).Return(sizeCounts, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/coupons-by-size", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result CouponsBySizeResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, int64(10), result.Size21x30)
	assert.Equal(t, int64(60), result.Size50x70)

	suite.mocks.CouponRepo.AssertExpectations(t)
}

func TestIntegration_GetTopPartners(t *testing.T) {
	suite := SetupIntegrationTest(t)

	partners := []*partner.Partner{createTestPartner(), createTestPartner()}
	suite.mocks.PartnerRepo.On("GetAll", mock.Anything, "created_at", "desc").Return(partners, nil)

	for i, p := range partners {
		suite.mocks.CouponRepo.On("CountByPartner", mock.Anything, p.ID).Return(int64(100-i*50), nil)
		suite.mocks.CouponRepo.On("CountActivatedByPartner", mock.Anything, p.ID).Return(int64(80-i*40), nil)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/top-partners?limit=5", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TopPartnersResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Len(t, result.Partners, 2)
	if len(result.Partners) > 1 {
		assert.True(t, result.Partners[0].ActivatedCoupons >= result.Partners[1].ActivatedCoupons)
	}

	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.CouponRepo.AssertExpectations(t)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCases := []struct {
		name           string
		endpoint       string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name:     "Database error during general stats",
			endpoint: "/admin/stats/general",
			setupMocks: func() {
				suite.mocks.RedisClient.On("Get", mock.Anything, "general_stats").Return("", errors.New("not found"))
				suite.mocks.CouponRepo.On("CountTotal", mock.Anything).Return(int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "Partner not found",
			endpoint: "/admin/stats/partners/" + uuid.New().String(),
			setupMocks: func() {
				suite.mocks.RedisClient.On("Get", mock.Anything, mock.AnythingOfType("string")).Return("", errors.New("not found"))
				suite.mocks.PartnerRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("partner not found"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "Redis error during system health",
			endpoint: "/admin/stats/system-health",
			setupMocks: func() {
				suite.mocks.RedisClient.On("Get", mock.Anything, "system_health").Return("", errors.New("not found"))
				suite.mocks.CouponRepo.On("HealthCheck", mock.Anything).Return(nil)
				suite.mocks.RedisClient.On("Ping", mock.Anything).Return(errors.New("redis error"))
				suite.mocks.RedisClient.On("LLen", mock.Anything, "image_processing_queue").Return(int64(0), nil)
				suite.mocks.RedisClient.On("Set", mock.Anything, "system_health", mock.AnythingOfType("[]uint8"), 1*time.Minute).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)
			tc.setupMocks()

			req := httptest.NewRequest(http.MethodGet, tc.endpoint, nil)
			resp, err := testSuite.app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTest(t)

	suite.mocks.RedisClient.On("Get", mock.Anything, "general_stats").Return("", errors.New("not found"))
	suite.mocks.CouponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
	suite.mocks.CouponRepo.On("CountActivated", mock.Anything).Return(int64(750), nil)
	suite.mocks.PartnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
	suite.mocks.PartnerRepo.On("CountTotal", mock.Anything).Return(int64(60), nil)
	suite.mocks.RedisClient.On("Set", mock.Anything, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	numRequests := 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/admin/stats/general", nil)
			resp, err := suite.app.Test(req)
			if err != nil {
				results <- err
				return
			}
			if resp.StatusCode != http.StatusOK {
				results <- errors.New("unexpected status code")
				return
			}
			results <- nil
		}()
	}

	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

func TestIntegration_QueryParameters(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := suite.testData.Partner
	partnerUUID := testPartner.ID

	rawData := []map[string]interface{}{
		{
			"date":               "2023-01-01",
			"coupons_created":    int64(5),
			"coupons_activated":  int64(4),
			"coupons_purchased":  int64(2),
			"new_partners_count": int64(0),
		},
	}

	suite.mocks.CouponRepo.On("GetTimeSeriesData", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"), "week", &partnerUUID).Return(rawData, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/stats/time-series?partner_id="+partnerUUID.String()+"&period=week&date_from=2023-01-01&date_to=2023-01-07", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result TimeSeriesStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "week", result.Period)
	assert.Len(t, result.Data, 1)

	suite.mocks.CouponRepo.AssertExpectations(t)
}

func BenchmarkIntegration_GetGeneralStats(b *testing.B) {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		mocks: &IntegrationMocks{
			CouponRepo:  &MockCouponRepository{},
			PartnerRepo: &MockPartnerRepository{},
			RedisClient: &MockRedisClient{},
		},
	}

	deps := &StatsServiceDeps{
		CouponRepository:  suite.mocks.CouponRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		RedisClient:       suite.mocks.RedisClient,
	}
	service := NewStatsService(deps)

	handlerDeps := &StatsHandlerDeps{
		StatsService: service,
	}
	handler := &StatsHandler{
		Router: suite.app,
		deps:   handlerDeps,
	}

	suite.app.Get("/admin/stats/general", handler.GetGeneralStats)

	suite.mocks.RedisClient.On("Get", mock.Anything, "general_stats").Return("", errors.New("not found"))
	suite.mocks.CouponRepo.On("CountTotal", mock.Anything).Return(int64(1000), nil)
	suite.mocks.CouponRepo.On("CountActivated", mock.Anything).Return(int64(750), nil)
	suite.mocks.PartnerRepo.On("CountActive", mock.Anything).Return(int64(50), nil)
	suite.mocks.PartnerRepo.On("CountTotal", mock.Anything).Return(int64(60), nil)
	suite.mocks.RedisClient.On("Set", mock.Anything, "general_stats", mock.AnythingOfType("[]uint8"), 5*time.Minute).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/admin/stats/general", nil)
		resp, err := suite.app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Fatal("unexpected status code")
		}
		resp.Body.Close()
	}
}
