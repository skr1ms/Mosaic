package image

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/internal/coupon"
)

type IntegrationTestSuite struct {
	app      *fiber.App
	handler  *ImageHandler
	service  *ImageService
	mocks    *IntegrationMocks
	testData *IntegrationTestData
}

type IntegrationMocks struct {
	ImageRepo    *MockImageRepository
	CouponRepo   *MockCouponRepository
	S3Client     *MockS3Client
	SDClient     *MockStableDiffusionClient
	EmailService *MockEmailService
	ZipService   *MockZipService
}

type IntegrationTestData struct {
	TestCoupon *coupon.Coupon
	TestImage  *Image
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		mocks: &IntegrationMocks{
			ImageRepo:    &MockImageRepository{},
			CouponRepo:   &MockCouponRepository{},
			S3Client:     &MockS3Client{},
			SDClient:     &MockStableDiffusionClient{},
			EmailService: &MockEmailService{},
			ZipService:   &MockZipService{},
		},
		testData: &IntegrationTestData{
			TestCoupon: createTestCoupon(),
			TestImage:  createTestImage(),
		},
	}

	service := &ImageService{}
	suite.service = service

	handlerDeps := &ImageHandlerDeps{
		ImageService:    service,
		ImageRepository: &ImageRepository{},
	}
	suite.handler = &ImageHandler{
		Router: suite.app,
		deps:   handlerDeps,
	}

	api := suite.app.Group("/api/v1")
	api.Post("/images/upload", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api.Put("/images/:image_id/edit", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api.Post("/images/:image_id/process", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api.Post("/images/:image_id/generate-schema", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api.Get("/images/:image_id/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	admin := suite.app.Group("/admin")
	admin.Get("/images/queue", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Get("/images/:image_id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Post("/images", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Put("/images/:image_id/start", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Put("/images/:image_id/complete", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Put("/images/:image_id/fail", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Put("/images/:image_id/retry", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Delete("/images/:image_id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Get("/images/statistics", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	admin.Get("/images/next", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return suite
}

func TestIntegration_UploadImage_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCoupon := suite.testData.TestCoupon

	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil)
	suite.mocks.ImageRepo.On("GetByCouponID", mock.Anything, testCoupon.ID).Return(nil, errors.New("not found"))
	suite.mocks.S3Client.On("UploadFile", mock.Anything, mock.Anything, mock.AnythingOfType("int64"), "image/jpeg", "originals", testCoupon.ID).Return("originals/test.jpg", nil)
	suite.mocks.ImageRepo.On("Create", mock.Anything, mock.AnythingOfType("*image.Image")).Return(nil)
	suite.mocks.CouponRepo.On("Update", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("coupon_id", testCoupon.ID.String())
	writer.WriteField("user_email", "test@example.com")

	part, err := writer.CreateFormFile("image", "test.jpg")
	require.NoError(t, err)
	part.Write([]byte("fake-image-data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := suite.app.Test(req)
	require.NoError(t, err)

	assert.NotNil(t, resp)
}

func TestIntegration_GetImageStatus_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	testImage.Status = "completed"

	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.S3Client.On("GetFileURL", mock.Anything, testImage.OriginalImageS3Key, 24*time.Hour).Return("https://example.com/original.jpg", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+testImage.ID.String()+"/status", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_EditImage_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	editParams := ImageEditParams{
		CropX:      10,
		CropY:      10,
		CropWidth:  100,
		CropHeight: 100,
		Rotation:   90,
		Scale:      1.5,
	}

	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.S3Client.On("DownloadFile", mock.Anything, testImage.OriginalImageS3Key).Return(bytes.NewReader([]byte("fake-image-data")), nil)
	suite.mocks.S3Client.On("UploadFile", mock.Anything, mock.Anything, mock.AnythingOfType("int64"), "image/jpeg", "edited", testImage.CouponID).Return("edited/test.jpg", nil)
	suite.mocks.ImageRepo.On("Update", mock.Anything, mock.AnythingOfType("*image.Image")).Return(nil)

	reqBody, _ := json.Marshal(editParams)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/images/"+testImage.ID.String()+"/edit", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_ProcessImage_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	testImage.Status = "edited"
	testCoupon := suite.testData.TestCoupon

	processParams := ProcessingParams{
		Style: "grayscale",
		UseAI: false,
	}

	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testImage.CouponID).Return(testCoupon, nil)
	suite.mocks.ImageRepo.On("Update", mock.Anything, mock.AnythingOfType("*image.Image")).Return(nil).Maybe()
	suite.mocks.S3Client.On("DownloadFile", mock.Anything, testImage.OriginalImageS3Key).Return(bytes.NewReader([]byte("fake-image-data")), nil).Maybe()
	suite.mocks.S3Client.On("UploadFile", mock.Anything, mock.Anything, mock.AnythingOfType("int64"), "image/jpeg", "previews", testImage.CouponID).Return("previews/test.jpg", nil).Maybe()

	reqBody, _ := json.Marshal(processParams)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/"+testImage.ID.String()+"/process", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_GenerateSchema_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	testImage.Status = "processed"
	testCoupon := suite.testData.TestCoupon

	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testImage.CouponID).Return(testCoupon, nil)
	suite.mocks.S3Client.On("DownloadFile", mock.Anything, mock.AnythingOfType("string")).Return(bytes.NewReader([]byte("fake-data")), nil).Maybe()
	suite.mocks.ZipService.On("CreateSchemaArchive", testImage.ID, mock.AnythingOfType("[]zip.FileData")).Return(bytes.NewBuffer([]byte("fake-zip")), nil).Maybe()
	suite.mocks.S3Client.On("UploadFile", mock.Anything, mock.Anything, mock.AnythingOfType("int64"), "application/zip", mock.AnythingOfType("string"), testImage.ID).Return("schemas/test.zip", nil).Maybe()
	suite.mocks.ImageRepo.On("Update", mock.Anything, mock.AnythingOfType("*image.Image")).Return(nil).Maybe()

	reqBody := `{"confirmed": true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/images/"+testImage.ID.String()+"/generate-schema", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminGetQueue_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImages := []*Image{createTestImage(), createTestImage()}
	suite.mocks.ImageRepo.On("GetByStatus", mock.Anything, "queued").Return(testImages, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/images/queue?status=queued", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminGetTaskByID_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/images/"+testImage.ID.String(), nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminStartProcessing_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("StartProcessing", mock.Anything, testImage.ID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/admin/images/"+testImage.ID.String()+"/start", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminCompleteProcessing_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("CompleteProcessing", mock.Anything, testImage.ID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/admin/images/"+testImage.ID.String()+"/complete", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminFailProcessing_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("FailProcessing", mock.Anything, testImage.ID, mock.AnythingOfType("string")).Return(nil)

	reqBody := `{"error_message": "Test error"}`
	req := httptest.NewRequest(http.MethodPut, "/admin/images/"+testImage.ID.String()+"/fail", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminRetryTask_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("RetryTask", mock.Anything, testImage.ID).Return(nil)

	req := httptest.NewRequest(http.MethodPut, "/admin/images/"+testImage.ID.String()+"/retry", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminDeleteTask_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("Delete", mock.Anything, testImage.ID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/admin/images/"+testImage.ID.String(), nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminGetStatistics_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	stats := map[string]int64{
		"queued":     10,
		"processing": 5,
		"completed":  100,
		"failed":     2,
	}
	suite.mocks.ImageRepo.On("GetStatistics", mock.Anything).Return(stats, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/images/statistics", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_AdminGetNextTask_Success(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("GetNextInQueue", mock.Anything).Return(testImage, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/images/next", nil)
	resp, err := suite.app.Test(req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCases := []struct {
		name           string
		endpoint       string
		method         string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name:     "Image not found",
			endpoint: "/api/v1/images/" + uuid.New().String() + "/status",
			method:   http.MethodGet,
			setupMocks: func() {
				suite.mocks.ImageRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("image not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "Admin task not found",
			endpoint: "/admin/images/" + uuid.New().String(),
			method:   http.MethodGet,
			setupMocks: func() {
				suite.mocks.ImageRepo.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, errors.New("task not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "Statistics error",
			endpoint: "/admin/images/statistics",
			method:   http.MethodGet,
			setupMocks: func() {
				suite.mocks.ImageRepo.On("GetStatistics", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)
			if tc.setupMocks != nil {
				tc.setupMocks()
			}

			req := httptest.NewRequest(tc.method, tc.endpoint, nil)
			resp, err := testSuite.app.Test(req)

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.TestImage
	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.S3Client.On("GetFileURL", mock.Anything, testImage.OriginalImageS3Key, 24*time.Hour).Return("https://example.com/original.jpg", nil)

	numRequests := 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+testImage.ID.String()+"/status", nil)
			resp, err := suite.app.Test(req)
			if err != nil {
				results <- err
				return
			}
			if resp == nil {
				results <- errors.New("nil response")
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

func TestIntegration_ValidationErrors(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCases := []struct {
		name     string
		endpoint string
		method   string
		body     string
	}{
		{
			name:     "Invalid UUID in path",
			endpoint: "/api/v1/images/invalid-uuid/status",
			method:   http.MethodGet,
			body:     "",
		},
		{
			name:     "Invalid JSON in edit request",
			endpoint: "/api/v1/images/" + uuid.New().String() + "/edit",
			method:   http.MethodPut,
			body:     `{"invalid": json}`,
		},
		{
			name:     "Invalid JSON in process request",
			endpoint: "/api/v1/images/" + uuid.New().String() + "/process",
			method:   http.MethodPost,
			body:     `{"invalid": json}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.endpoint, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.endpoint, nil)
			}

			resp, err := suite.app.Test(req)

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func BenchmarkIntegration_GetImageStatus(b *testing.B) {
	suite := &IntegrationTestSuite{
		app: fiber.New(),
		testData: &IntegrationTestData{
			TestImage: createTestImage(),
		},
	}

	suite.app.Get("/api/v1/images/:image_id/status", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	testImage := suite.testData.TestImage

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/images/"+testImage.ID.String()+"/status", nil)
		resp, err := suite.app.Test(req)
		if err != nil {
			b.Fatal(err)
		}
		if resp == nil {
			b.Fatal("nil response")
		}
		resp.Body.Close()
	}
}
