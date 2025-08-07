package partner

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
)

type IntegrationTestSuite struct {
	partnerRepo   *MockPartnerRepository
	couponService *MockCouponService
	recaptcha     *MockRecaptcha
	jwtService    *MockJWT
	mailer        *MockMailer
	config        *config.Config
	service       *PartnerService
}

func SetupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		partnerRepo:   &MockPartnerRepository{},
		couponService: &MockCouponService{},
		recaptcha:     &MockRecaptcha{},
		jwtService:    &MockJWT{},
		mailer:        &MockMailer{},
		config:        createTestConfig(),
	}

	deps := &PartnerServiceDeps{
		PartnerRepository: suite.partnerRepo,
		CouponService:     suite.couponService,
		Recaptcha:         suite.recaptcha,
		JwtService:        suite.jwtService,
		MailSender:        suite.mailer,
		Config:            suite.config,
	}

	suite.service = NewPartnerService(deps)
	return suite
}

func TestIntegration_PartnerPasswordResetWorkflow(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testPartner := createTestPartner()
	email := testPartner.Email
	resetToken := "reset-token-123"
	newPassword := "newpassword123"
	resetLink := suite.config.ServerConfig.FrontendURL + "/reset?token=" + resetToken

	suite.partnerRepo.On("GetByEmail", mock.Anything, email).Return(testPartner, nil)
	suite.jwtService.On("CreatePasswordResetToken", testPartner.ID, testPartner.Email).Return(resetToken, nil)
	suite.mailer.On("SendResetPasswordEmail", email, resetLink).Return(nil)

	claims := &TokenClaims{
		UserID: testPartner.ID,
		Login:  testPartner.Email,
	}
	suite.jwtService.On("ValidatePasswordResetToken", resetToken).Return(claims, nil)
	suite.partnerRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)
	suite.partnerRepo.On("UpdatePassword", mock.Anything, testPartner.ID, mock.AnythingOfType("string")).Return(nil)

	err := suite.service.ForgotPassword(context.Background(), email)
	require.NoError(t, err)

	err = suite.service.ResetPassword(context.Background(), resetToken, newPassword)
	require.NoError(t, err)

	suite.partnerRepo.AssertExpectations(t)
	suite.jwtService.AssertExpectations(t)
	suite.mailer.AssertExpectations(t)
}

func TestIntegration_PartnerCouponExportWorkflow(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	partnerID := uuid.New()
	status := "new"
	format := "txt"

	exportOptions := coupon.ExportOptionsRequest{
		Format:        coupon.ExportFormatCodes,
		PartnerID:     stringPtr(partnerID.String()),
		Status:        status,
		FileFormat:    format,
		IncludeHeader: false,
	}

	expectedContent := []byte("COUPON001\nCOUPON002\nCOUPON003\n")
	expectedFilename := "partner_coupons.txt"
	expectedContentType := "text/plain"

	suite.couponService.On("ExportCouponsAdvanced", exportOptions).Return(
		expectedContent, expectedFilename, expectedContentType, nil)

	content, filename, contentType, err := suite.service.ExportCoupons(partnerID, status, format)

	require.NoError(t, err)
	assert.Equal(t, expectedContent, content)
	assert.Equal(t, expectedFilename, filename)
	assert.Equal(t, expectedContentType, contentType)

	suite.couponService.AssertExpectations(t)
}

func TestIntegration_PartnerPasswordUpdateWorkflow(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testPartner := createTestPartner()
	testPartner.Password = "$2a$10$uIAjQp4L.ABjHuGRFySv9uHg96dL3O4YL9ffIIAemlKvFgpkCcDYO"

	partnerID := testPartner.ID
	currentPassword := "currentpassword"
	newPassword := "newpassword123"

	suite.partnerRepo.On("GetByID", mock.Anything, partnerID).Return(testPartner, nil)
	suite.partnerRepo.On("UpdatePassword", mock.Anything, partnerID, mock.AnythingOfType("string")).Return(nil)

	err := suite.service.UpdatePassword(partnerID, currentPassword, newPassword)

	require.NoError(t, err)
	suite.partnerRepo.AssertExpectations(t)
}

func TestIntegration_PartnerDeletionWorkflow(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	partnerID := uuid.New()

	suite.partnerRepo.On("DeleteWithCoupons", mock.Anything, partnerID).Return(nil)

	err := suite.service.DeletePartnerWithCoupons(context.Background(), partnerID)

	require.NoError(t, err)
	suite.partnerRepo.AssertExpectations(t)
}

func TestIntegration_PartnerSearchAndStatistics(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	searchQuery := "test"
	status := "active"
	sortBy := "created_at"
	order := "desc"

	partners := []*Partner{createTestPartner(), createTestPartner()}
	suite.partnerRepo.On("Search", mock.Anything, searchQuery, status, sortBy, order).Return(partners, nil)

	searchResults, err := suite.partnerRepo.Search(context.Background(), searchQuery, status, sortBy, order)
	require.NoError(t, err)
	assert.Len(t, searchResults, 2)

	suite.partnerRepo.On("CountActive", mock.Anything).Return(int64(5), nil)
	suite.partnerRepo.On("CountTotal", mock.Anything).Return(int64(10), nil)

	activeCount, err := suite.partnerRepo.CountActive(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(5), activeCount)

	totalCount, err := suite.partnerRepo.CountTotal(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(10), totalCount)

	suite.partnerRepo.AssertExpectations(t)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testCases := []struct {
		name          string
		setupMocks    func()
		action        func() error
		expectedError string
	}{
		{
			name: "Partner not found during password reset",
			setupMocks: func() {
				suite.partnerRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("partner not found"))
			},
			action: func() error {
				return suite.service.ForgotPassword(context.Background(), "nonexistent@example.com")
			},
			expectedError: "failed to find partner by email",
		},
		{
			name: "Inactive partner tries to reset password",
			setupMocks: func() {
				inactivePartner := createTestPartner()
				inactivePartner.Status = "blocked"
				suite.partnerRepo.On("GetByEmail", mock.Anything, inactivePartner.Email).Return(inactivePartner, nil)
			},
			action: func() error {
				inactivePartner := createTestPartner()
				inactivePartner.Status = "blocked"
				return suite.service.ForgotPassword(context.Background(), inactivePartner.Email)
			},
			expectedError: "partner status is not active",
		},
		{
			name: "Export service failure",
			setupMocks: func() {
				partnerID := uuid.New().String()
				exportOptions := coupon.ExportOptionsRequest{
					Format:        coupon.ExportFormatCodes,
					PartnerID:     &partnerID,
					Status:        "new",
					FileFormat:    "txt",
					IncludeHeader: false,
				}
				suite.couponService.On("ExportCouponsAdvanced", exportOptions).Return([]byte{}, "", "", errors.New("export service error"))
			},
			action: func() error {
				partnerID := uuid.New()
				_, _, _, err := suite.service.ExportCoupons(partnerID, "new", "txt")
				return err
			},
			expectedError: "failed to export coupons",
		},
		{
			name: "Wrong current password during update",
			setupMocks: func() {
				testPartner := createTestPartner()
				testPartner.Password = "$2a$10$N9qo8uLOickgx2ZMRZoMye.fRQDOYF.7QOLc0dGi5t8n4hq5jKRfS"
				suite.partnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)
			},
			action: func() error {
				testPartner := createTestPartner()
				return suite.service.UpdatePassword(testPartner.ID, "wrongpassword", "newpassword")
			},
			expectedError: "current password is incorrect",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTestSuite(t)
			tc.setupMocks = func() {
				switch tc.name {
				case "Partner not found during password reset":
					testSuite.partnerRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("partner not found"))
				case "Inactive partner tries to reset password":
					inactivePartner := createTestPartner()
					inactivePartner.Status = "blocked"
					testSuite.partnerRepo.On("GetByEmail", mock.Anything, inactivePartner.Email).Return(inactivePartner, nil)
				case "Export service failure":
					testSuite.couponService.On("ExportCouponsAdvanced", mock.AnythingOfType("coupon.ExportOptionsRequest")).Return([]byte{}, "", "", errors.New("export service error"))
				case "Wrong current password during update":
					testPartner := createTestPartner()
					testPartner.Password = "$2a$10$N9qo8uLOickgx2ZMRZoMye.fRQDOYF.7QOLc0dGi5t8n4hq5jKRfS"
					testSuite.partnerRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(testPartner, nil)
				}
			}

			tc.action = func() error {
				switch tc.name {
				case "Partner not found during password reset":
					return testSuite.service.ForgotPassword(context.Background(), "nonexistent@example.com")
				case "Inactive partner tries to reset password":
					inactivePartner := createTestPartner()
					inactivePartner.Status = "blocked"
					return testSuite.service.ForgotPassword(context.Background(), inactivePartner.Email)
				case "Export service failure":
					partnerID := uuid.New()
					_, _, _, err := testSuite.service.ExportCoupons(partnerID, "new", "txt")
					return err
				case "Wrong current password during update":
					testPartner := createTestPartner()
					return testSuite.service.UpdatePassword(testPartner.ID, "wrongpassword", "newpassword")
				}
				return nil
			}

			tc.setupMocks()
			err := tc.action()

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	numPartners := 5
	partners := make([]*Partner, numPartners)
	for i := 0; i < numPartners; i++ {
		partners[i] = createTestPartner()
		partners[i].Email = generateTestEmail(i)
		suite.partnerRepo.On("GetByEmail", mock.Anything, partners[i].Email).Return(partners[i], nil)
	}

	results := make(chan error, numPartners)

	for i := 0; i < numPartners; i++ {
		go func(email string) {
			_, err := suite.partnerRepo.GetByEmail(context.Background(), email)
			results <- err
		}(partners[i].Email)
	}

	for i := 0; i < numPartners; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	suite.partnerRepo.AssertExpectations(t)
}

func TestIntegration_PartnerLifecycle(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)

	testPartner := createTestPartner()

	suite.partnerRepo.On("Create", mock.Anything, mock.AnythingOfType("*partner.Partner")).Return(nil)

	suite.partnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	suite.partnerRepo.On("UpdateStatus", mock.Anything, testPartner.ID, "blocked").Return(nil)

	suite.partnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)

	suite.partnerRepo.On("DeleteWithCoupons", mock.Anything, testPartner.ID).Return(nil)

	err := suite.partnerRepo.Create(context.Background(), testPartner)
	require.NoError(t, err)

	partner, err := suite.partnerRepo.GetByID(context.Background(), testPartner.ID)
	require.NoError(t, err)
	assert.Equal(t, testPartner.ID, partner.ID)

	err = suite.partnerRepo.UpdateStatus(context.Background(), testPartner.ID, "blocked")
	require.NoError(t, err)

	err = suite.partnerRepo.UpdateLastLogin(context.Background(), testPartner.ID)
	require.NoError(t, err)

	err = suite.partnerRepo.DeleteWithCoupons(context.Background(), testPartner.ID)
	require.NoError(t, err)

	suite.partnerRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

func generateTestEmail(index int) string {
	return "test" + string(rune('0'+index)) + "@example.com"
}

func BenchmarkIntegration_PartnerOperations(b *testing.B) {
	suite := &IntegrationTestSuite{
		partnerRepo: &MockPartnerRepository{},
	}

	testPartner := createTestPartner()
	suite.partnerRepo.On("GetByID", mock.Anything, testPartner.ID).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := suite.partnerRepo.GetByID(context.Background(), testPartner.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIntegration_ConcurrentPartnerLookup(b *testing.B) {
	suite := &IntegrationTestSuite{
		partnerRepo: &MockPartnerRepository{},
		config:      createTestConfig(),
	}

	testPartner := createTestPartner()
	suite.partnerRepo.On("GetByEmail", mock.Anything, testPartner.Email).Return(testPartner, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := suite.partnerRepo.GetByEmail(context.Background(), testPartner.Email)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
