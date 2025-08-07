package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	app     *fiber.App
	service *AuthService
	mocks   *IntegrationMocks
}

type IntegrationMocks struct {
	AdminRepo   *MockAdminRepository
	PartnerRepo *MockPartnerRepository
	JwtService  *MockJwtService
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		mocks: &IntegrationMocks{
			AdminRepo:   &MockAdminRepository{},
			PartnerRepo: &MockPartnerRepository{},
			JwtService:  &MockJwtService{},
		},
	}

	deps := &AuthServiceDeps{
		AdminRepository:   suite.mocks.AdminRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		JwtService:        suite.mocks.JwtService,
	}
	suite.service = NewAuthService(deps)

	handlerDeps := &AuthHandlerDeps{
		AuthService: suite.service,
	}

	NewAuthHandler(suite.app, handlerDeps)

	return suite
}

func TestIntegration_AdminLogin_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	suite.mocks.AdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	suite.mocks.JwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	suite.mocks.AdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	requestBody := LoginRequest{
		Login:    "admin123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
	assert.Contains(t, response, "admin")

	adminData := response["admin"].(map[string]interface{})
	assert.Equal(t, testAdmin.ID.String(), adminData["id"])
	assert.Equal(t, testAdmin.Login, adminData["login"])
	assert.Equal(t, "admin", adminData["role"])

	suite.mocks.AdminRepo.AssertExpectations(t)
	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_AdminLogin_InvalidCredentials(t *testing.T) {
	suite := SetupIntegrationTest(t)

	suite.mocks.AdminRepo.On("GetByLogin", "nonexistent123").Return(nil, errors.New("admin not found"))

	requestBody := LoginRequest{
		Login:    "nonexistent123",
		Password: "wrongpassword",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	suite.mocks.AdminRepo.AssertExpectations(t)
}

func TestIntegration_AdminLogin_InvalidRequestBody(t *testing.T) {
	suite := SetupIntegrationTest(t)

	req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_PartnerLogin_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := createTestPartner()
	testTokenPair := createTestTokenPair()

	suite.mocks.PartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)
	suite.mocks.JwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(testTokenPair, nil)
	suite.mocks.PartnerRepo.On("UpdateLastLogin", mock.Anything, testPartner.ID).Return(nil)

	requestBody := LoginRequest{
		Login:    "partner123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/partner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
	assert.Contains(t, response, "partner")

	partnerData := response["partner"].(map[string]interface{})
	assert.Equal(t, testPartner.ID.String(), partnerData["id"])
	assert.Equal(t, testPartner.Login, partnerData["login"])
	assert.Equal(t, testPartner.PartnerCode, partnerData["partner_code"])
	assert.Equal(t, testPartner.BrandName, partnerData["brand_name"])
	assert.Equal(t, "partner", partnerData["role"])

	suite.mocks.PartnerRepo.AssertExpectations(t)
	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_PartnerLogin_PartnerBlocked(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := createTestPartner()
	testPartner.Status = "blocked"

	suite.mocks.PartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)

	requestBody := LoginRequest{
		Login:    "partner123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/partner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	suite.mocks.PartnerRepo.AssertExpectations(t)
}

func TestIntegration_RefreshAdminTokens_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testClaims := createTestClaims("admin")
	testTokenPair := createTestTokenPair()

	suite.mocks.JwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	suite.mocks.JwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	requestBody := RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")
	assert.Equal(t, testTokenPair.AccessToken, response["access_token"])
	assert.Equal(t, testTokenPair.RefreshToken, response["refresh_token"])

	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_RefreshAdminTokens_InvalidToken(t *testing.T) {
	suite := SetupIntegrationTest(t)

	suite.mocks.JwtService.On("ValidateRefreshToken", "invalid_token").Return(nil, errors.New("invalid token"))

	requestBody := RefreshTokenRequest{
		RefreshToken: "invalid_token",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_RefreshPartnerTokens_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testClaims := createTestClaims("partner")
	testTokenPair := createTestTokenPair()

	suite.mocks.JwtService.On("ValidateRefreshToken", "valid_refresh_token").Return(testClaims, nil)
	suite.mocks.JwtService.On("RefreshTokens", "valid_refresh_token").Return(testTokenPair, nil)

	requestBody := RefreshTokenRequest{
		RefreshToken: "valid_refresh_token",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh/partner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Contains(t, response, "access_token")
	assert.Contains(t, response, "refresh_token")

	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_RefreshPartnerTokens_WrongRole(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testClaims := createTestClaims("admin") // Wrong role for partner refresh

	suite.mocks.JwtService.On("ValidateRefreshToken", "admin_refresh_token").Return(testClaims, nil)

	requestBody := RefreshTokenRequest{
		RefreshToken: "admin_refresh_token",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/refresh/partner", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	suite.mocks.AdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil).Times(5)
	suite.mocks.JwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil).Times(5)
	suite.mocks.AdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil).Times(5)

	requestBody := LoginRequest{
		Login:    "admin123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)

	numRequests := 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

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

	suite.mocks.AdminRepo.AssertExpectations(t)
	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	testCases := []struct {
		name           string
		method         string
		endpoint       string
		body           interface{}
		setupMocks     func(*IntegrationMocks)
		expectedStatus int
	}{
		{
			name:     "Admin login - database error",
			method:   http.MethodPost,
			endpoint: "/login/admin",
			body: LoginRequest{
				Login:    "admin123",
				Password: "password123",
			},
			setupMocks: func(mocks *IntegrationMocks) {
				mocks.AdminRepo.On("GetByLogin", "admin123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Partner login - token generation error",
			method:   http.MethodPost,
			endpoint: "/login/partner",
			body: LoginRequest{
				Login:    "partner123",
				Password: "password123",
			},
			setupMocks: func(mocks *IntegrationMocks) {
				testPartner := createTestPartner()
				mocks.PartnerRepo.On("GetByLogin", mock.Anything, "partner123").Return(testPartner, nil)
				mocks.JwtService.On("CreateTokenPair", testPartner.ID, testPartner.Login, "partner").Return(nil, errors.New("token generation failed"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "Admin refresh - invalid request body",
			method:   http.MethodPost,
			endpoint: "/refresh/admin",
			body:     "invalid json",
			setupMocks: func(mocks *IntegrationMocks) {
				// без моков
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "Partner refresh - service error",
			method:   http.MethodPost,
			endpoint: "/refresh/partner",
			body: RefreshTokenRequest{
				RefreshToken: "some_token",
			},
			setupMocks: func(mocks *IntegrationMocks) {
				mocks.JwtService.On("ValidateRefreshToken", "some_token").Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			suite := SetupIntegrationTest(t)
			tc.setupMocks(suite.mocks)

			var jsonData []byte
			var err error

			if str, ok := tc.body.(string); ok {
				jsonData = []byte(str)
			} else {
				jsonData, err = json.Marshal(tc.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(tc.method, tc.endpoint, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestIntegration_RequestValidation(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCases := []struct {
		name     string
		endpoint string
		body     interface{}
		expected int
	}{
		{
			name:     "Short login request",
			endpoint: "/login/admin",
			body: LoginRequest{
				Login:    "ab",
				Password: "password123",
			},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Short password request",
			endpoint: "/login/partner",
			body: LoginRequest{
				Login:    "partner123",
				Password: "",
			},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Empty refresh token",
			endpoint: "/refresh/admin",
			body: RefreshTokenRequest{
				RefreshToken: "",
			},
			expected: http.StatusBadRequest,
		},
		{
			name:     "Missing fields",
			endpoint: "/login/admin",
			body:     map[string]string{},
			expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tc.body)
			req := httptest.NewRequest(http.MethodPost, tc.endpoint, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			resp, err := suite.app.Test(req)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, resp.StatusCode)
		})
	}
}

func TestIntegration_ResponseFormat(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	suite.mocks.AdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	suite.mocks.JwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	suite.mocks.AdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	requestBody := LoginRequest{
		Login:    "admin123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	requiredFields := []string{"message", "admin", "access_token", "refresh_token", "expires_in"}
	for _, field := range requiredFields {
		assert.Contains(t, response, field, "Response should contain field: %s", field)
	}

	admin := response["admin"].(map[string]interface{})
	adminFields := []string{"id", "login", "role"}
	for _, field := range adminFields {
		assert.Contains(t, admin, field, "Admin object should contain field: %s", field)
	}

	// Check daa types
	assert.IsType(t, "", response["access_token"])
	assert.IsType(t, "", response["refresh_token"])
	assert.IsType(t, float64(0), response["expires_in"])

	suite.mocks.AdminRepo.AssertExpectations(t)
	suite.mocks.JwtService.AssertExpectations(t)
}

func TestIntegration_Headers(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	suite.mocks.AdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	suite.mocks.JwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	suite.mocks.AdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	requestBody := LoginRequest{
		Login:    "admin123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "test-client/1.0")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	suite.mocks.AdminRepo.AssertExpectations(t)
	suite.mocks.JwtService.AssertExpectations(t)
}

func BenchmarkIntegration_AdminLogin(b *testing.B) {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		mocks: &IntegrationMocks{
			AdminRepo:   &MockAdminRepository{},
			PartnerRepo: &MockPartnerRepository{},
			JwtService:  &MockJwtService{},
		},
	}

	deps := &AuthServiceDeps{
		AdminRepository:   suite.mocks.AdminRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		JwtService:        suite.mocks.JwtService,
	}
	service := NewAuthService(deps)

	handlerDeps := &AuthHandlerDeps{
		AuthService: service,
	}

	NewAuthHandler(suite.app, handlerDeps)

	testAdmin := createTestAdmin()
	testTokenPair := createTestTokenPair()

	suite.mocks.AdminRepo.On("GetByLogin", "admin123").Return(testAdmin, nil)
	suite.mocks.JwtService.On("CreateTokenPair", testAdmin.ID, testAdmin.Login, "admin").Return(testTokenPair, nil)
	suite.mocks.AdminRepo.On("UpdateLastLogin", testAdmin.ID).Return(nil)

	requestBody := LoginRequest{
		Login:    "admin123",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(requestBody)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login/admin", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")

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
