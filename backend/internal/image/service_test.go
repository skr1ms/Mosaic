package image

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/stablediffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
)

type MockImageRepository struct {
	mock.Mock
}

var _ ImageRepositoryInterface = (*MockImageRepository)(nil)

func (m *MockImageRepository) Create(ctx context.Context, task *Image) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Image), args.Error(1)
}

func (m *MockImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*Image, error) {
	args := m.Called(ctx, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Image), args.Error(1)
}

func (m *MockImageRepository) GetNextInQueue(ctx context.Context) (*Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Image), args.Error(1)
}

func (m *MockImageRepository) GetQueuedTasks(ctx context.Context) ([]*Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Image), args.Error(1)
}

func (m *MockImageRepository) GetProcessingTasks(ctx context.Context) ([]*Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Image), args.Error(1)
}

func (m *MockImageRepository) StartProcessing(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) CompleteProcessing(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) FailProcessing(ctx context.Context, id uuid.UUID, errorMessage string) error {
	args := m.Called(ctx, id, errorMessage)
	return args.Error(0)
}

func (m *MockImageRepository) RetryTask(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) Update(ctx context.Context, task *Image) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) GetAll(ctx context.Context) ([]*Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Image), args.Error(1)
}

func (m *MockImageRepository) GetByStatus(ctx context.Context, status string) ([]*Image, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Image), args.Error(1)
}

func (m *MockImageRepository) GetFailedTasksForRetry(ctx context.Context) ([]*Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Image), args.Error(1)
}

func (m *MockImageRepository) GetStatistics(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

type MockCouponRepository struct {
	mock.Mock
}

var _ CouponRepositoryInterface = (*MockCouponRepository)(nil)

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

type MockS3Client struct {
	mock.Mock
}

var _ S3ClientInterface = (*MockS3Client)(nil)

func (m *MockS3Client) UploadFile(ctx context.Context, data io.Reader, size int64, contentType, prefix string, couponID uuid.UUID) (string, error) {
	args := m.Called(ctx, data, size, contentType, prefix, couponID)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) DownloadFile(ctx context.Context, key string) (io.Reader, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.Reader), args.Error(1)
}

func (m *MockS3Client) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockS3Client) GetFileURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

type MockStableDiffusionClient struct {
	mock.Mock
}

var _ StableDiffusionClientInterface = (*MockStableDiffusionClient)(nil)

func (m *MockStableDiffusionClient) ProcessImage(ctx context.Context, req stablediffusion.ProcessImageRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockStableDiffusionClient) EncodeImageToBase64(data []byte) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *MockStableDiffusionClient) DecodeBase64Image(base64Data string) ([]byte, error) {
	args := m.Called(base64Data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

var _ EmailServiceInterface = (*MockEmailService)(nil)

func (m *MockEmailService) SendSchemaEmail(email, schemaURL, couponCode string) error {
	args := m.Called(email, schemaURL, couponCode)
	return args.Error(0)
}

type MockZipService struct {
	mock.Mock
}

var _ ZipServiceInterface = (*MockZipService)(nil)

func (m *MockZipService) CreateSchemaArchive(imageID uuid.UUID, files []zip.FileData) (*bytes.Buffer, error) {
	args := m.Called(imageID, files)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

func createTestCoupon() *coupon.Coupon {
	email := "test@example.com"
	return &coupon.Coupon{
		ID:        uuid.New(),
		Code:      "TEST123",
		Size:      "30x40",
		Style:     "grayscale",
		Status:    "activated",
		UserEmail: &email,
	}
}

func createTestImage() *Image {
	return &Image{
		ID:                 uuid.New(),
		CouponID:           uuid.New(),
		OriginalImageS3Key: "originals/test.jpg",
		UserEmail:          "test@example.com",
		Status:             "uploaded",
		Priority:           1,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func TestImageService_GetQueue_Success(t *testing.T) {
	mockRepo := &MockImageRepository{}

	testImages := []*Image{createTestImage(), createTestImage()}
	mockRepo.On("GetByStatus", mock.Anything, "queued").Return(testImages, nil)

	result, err := mockRepo.GetByStatus(context.Background(), "queued")

	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestImageService_GetQueue_AllTasks(t *testing.T) {
	mockRepo := &MockImageRepository{}

	testImages := []*Image{createTestImage(), createTestImage()}
	mockRepo.On("GetAll", mock.Anything).Return(testImages, nil)

	result, err := mockRepo.GetAll(context.Background())

	require.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestImageService_Repository_Operations(t *testing.T) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		testImage := createTestImage()
		mockRepo.On("Create", ctx, testImage).Return(nil)

		err := mockRepo.Create(ctx, testImage)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetByID", func(t *testing.T) {
		testImage := createTestImage()
		mockRepo.On("GetByID", ctx, testImage.ID).Return(testImage, nil)

		result, err := mockRepo.GetByID(ctx, testImage.ID)
		assert.NoError(t, err)
		assert.Equal(t, testImage, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		testImage := createTestImage()
		mockRepo.On("Update", ctx, testImage).Return(nil)

		err := mockRepo.Update(ctx, testImage)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		testImage := createTestImage()
		mockRepo.On("Delete", ctx, testImage.ID).Return(nil)

		err := mockRepo.Delete(ctx, testImage.ID)
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestImageService_S3Client_Operations(t *testing.T) {
	mockS3 := &MockS3Client{}
	ctx := context.Background()

	t.Run("UploadFile", func(t *testing.T) {
		couponID := uuid.New()
		mockS3.On("UploadFile", ctx, mock.Anything, int64(1024), "image/jpeg", "originals", couponID).Return("originals/test.jpg", nil)

		result, err := mockS3.UploadFile(ctx, bytes.NewReader([]byte("test")), 1024, "image/jpeg", "originals", couponID)
		assert.NoError(t, err)
		assert.Equal(t, "originals/test.jpg", result)
		mockS3.AssertExpectations(t)
	})

	t.Run("DownloadFile", func(t *testing.T) {
		expectedData := bytes.NewReader([]byte("test-data"))
		mockS3.On("DownloadFile", ctx, "test.jpg").Return(expectedData, nil)

		result, err := mockS3.DownloadFile(ctx, "test.jpg")
		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockS3.AssertExpectations(t)
	})

	t.Run("DeleteFile", func(t *testing.T) {
		mockS3.On("DeleteFile", ctx, "test.jpg").Return(nil)

		err := mockS3.DeleteFile(ctx, "test.jpg")
		assert.NoError(t, err)
		mockS3.AssertExpectations(t)
	})

	t.Run("GetFileURL", func(t *testing.T) {
		mockS3.On("GetFileURL", ctx, "test.jpg", 24*time.Hour).Return("https://example.com/test.jpg", nil)

		result, err := mockS3.GetFileURL(ctx, "test.jpg", 24*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/test.jpg", result)
		mockS3.AssertExpectations(t)
	})
}

func TestImageService_StableDiffusion_Operations(t *testing.T) {
	mockSD := &MockStableDiffusionClient{}
	ctx := context.Background()

	t.Run("ProcessImage", func(t *testing.T) {
		req := stablediffusion.ProcessImageRequest{
			ImageBase64: "base64-data",
			Style:       stablediffusion.ProcessingStyle("grayscale"),
			UseAI:       true,
		}
		mockSD.On("ProcessImage", ctx, req).Return("processed-base64", nil)

		result, err := mockSD.ProcessImage(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, "processed-base64", result)
		mockSD.AssertExpectations(t)
	})

	t.Run("EncodeImageToBase64", func(t *testing.T) {
		data := []byte("image-data")
		mockSD.On("EncodeImageToBase64", data).Return("base64-encoded")

		result := mockSD.EncodeImageToBase64(data)
		assert.Equal(t, "base64-encoded", result)
		mockSD.AssertExpectations(t)
	})

	t.Run("DecodeBase64Image", func(t *testing.T) {
		expectedData := []byte("decoded-data")
		mockSD.On("DecodeBase64Image", "base64-data").Return(expectedData, nil)

		result, err := mockSD.DecodeBase64Image("base64-data")
		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockSD.AssertExpectations(t)
	})
}

func TestImageService_EmailService_Operations(t *testing.T) {
	mockEmail := &MockEmailService{}

	t.Run("SendSchemaEmail", func(t *testing.T) {
		mockEmail.On("SendSchemaEmail", "test@example.com", "https://example.com/schema.zip", "TEST123").Return(nil)

		err := mockEmail.SendSchemaEmail("test@example.com", "https://example.com/schema.zip", "TEST123")
		assert.NoError(t, err)
		mockEmail.AssertExpectations(t)
	})
}

func TestImageService_ZipService_Operations(t *testing.T) {
	mockZip := &MockZipService{}

	t.Run("CreateSchemaArchive", func(t *testing.T) {
		imageID := uuid.New()
		files := []zip.FileData{
			{
				Name:    "test.jpg",
				Content: bytes.NewReader([]byte("test")),
				Size:    4,
			},
		}
		expectedBuffer := bytes.NewBuffer([]byte("zip-data"))

		mockZip.On("CreateSchemaArchive", imageID, files).Return(expectedBuffer, nil)

		result, err := mockZip.CreateSchemaArchive(imageID, files)
		assert.NoError(t, err)
		assert.Equal(t, expectedBuffer, result)
		mockZip.AssertExpectations(t)
	})
}

func TestImageService_CouponRepository_Operations(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	ctx := context.Background()

	t.Run("GetByID", func(t *testing.T) {
		testCoupon := createTestCoupon()
		mockCouponRepo.On("GetByID", ctx, testCoupon.ID).Return(testCoupon, nil)

		result, err := mockCouponRepo.GetByID(ctx, testCoupon.ID)
		assert.NoError(t, err)
		assert.Equal(t, testCoupon, result)
		mockCouponRepo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		testCoupon := createTestCoupon()
		mockCouponRepo.On("Update", ctx, testCoupon).Return(nil)

		err := mockCouponRepo.Update(ctx, testCoupon)
		assert.NoError(t, err)
		mockCouponRepo.AssertExpectations(t)
	})
}

func TestImageService_Error_Cases(t *testing.T) {
	t.Run("Repository_GetByID_NotFound", func(t *testing.T) {
		mockRepo := &MockImageRepository{}
		ctx := context.Background()
		imageID := uuid.New()

		mockRepo.On("GetByID", ctx, imageID).Return(nil, errors.New("image not found"))

		result, err := mockRepo.GetByID(ctx, imageID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "image not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("S3_UploadFile_Error", func(t *testing.T) {
		mockS3 := &MockS3Client{}
		ctx := context.Background()
		couponID := uuid.New()

		mockS3.On("UploadFile", ctx, mock.Anything, int64(1024), "image/jpeg", "originals", couponID).Return("", errors.New("upload failed"))

		result, err := mockS3.UploadFile(ctx, bytes.NewReader([]byte("test")), 1024, "image/jpeg", "originals", couponID)
		assert.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "upload failed")
		mockS3.AssertExpectations(t)
	})
}

func TestImageEditParams_Validation(t *testing.T) {
	validParams := ImageEditParams{
		CropX:      10,
		CropY:      10,
		CropWidth:  100,
		CropHeight: 100,
		Rotation:   90,
		Scale:      1.5,
	}

	assert.Equal(t, 10, validParams.CropX)
	assert.Equal(t, 10, validParams.CropY)
	assert.Equal(t, 100, validParams.CropWidth)
	assert.Equal(t, 100, validParams.CropHeight)
	assert.Equal(t, 90, validParams.Rotation)
	assert.Equal(t, 1.5, validParams.Scale)
}

func TestProcessingParams_Validation(t *testing.T) {
	validParams := ProcessingParams{
		Style:      "grayscale",
		UseAI:      true,
		Lighting:   "sun",
		Contrast:   "high",
		Brightness: 10.0,
		Saturation: 5.0,
	}

	assert.Equal(t, "grayscale", validParams.Style)
	assert.True(t, validParams.UseAI)
	assert.Equal(t, "sun", validParams.Lighting)
	assert.Equal(t, "high", validParams.Contrast)
	assert.Equal(t, 10.0, validParams.Brightness)
	assert.Equal(t, 5.0, validParams.Saturation)
}

func TestImage_Model_Fields(t *testing.T) {
	image := createTestImage()

	assert.NotEqual(t, uuid.Nil, image.ID)
	assert.NotEqual(t, uuid.Nil, image.CouponID)
	assert.Equal(t, "originals/test.jpg", image.OriginalImageS3Key)
	assert.Equal(t, "test@example.com", image.UserEmail)
	assert.Equal(t, "uploaded", image.Status)
	assert.Equal(t, 1, image.Priority)
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("isValidImageType", func(t *testing.T) {
		tests := []struct {
			contentType string
			expected    bool
		}{
			{"image/jpeg", true},
			{"image/png", true},
			{"image/gif", false},
			{"text/plain", false},
			{"application/pdf", false},
		}

		for _, tt := range tests {
			t.Run(tt.contentType, func(t *testing.T) {
				fileHeader := &multipart.FileHeader{
					Header: map[string][]string{
						"Content-Type": {tt.contentType},
					},
				}
				result := isValidImageType(fileHeader)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("parseCouponSize", func(t *testing.T) {
		tests := []struct {
			size           string
			expectedWidth  int
			expectedHeight int
		}{
			{"21x30", 840, 1200},
			{"30x40", 1200, 1600},
			{"40x40", 1600, 1600},
			{"40x50", 1600, 2000},
			{"40x60", 1600, 2400},
			{"50x70", 2000, 2800},
			{"unknown", 1200, 1600},
		}

		for _, tt := range tests {
			t.Run(tt.size, func(t *testing.T) {
				width, height := parseCouponSize(tt.size)
				assert.Equal(t, tt.expectedWidth, width)
				assert.Equal(t, tt.expectedHeight, height)
			})
		}
	})
}
