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
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/pkg/mosaic"
	"github.com/skr1ms/mosaic/pkg/stableDiffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockImageRepository struct {
	mock.Mock
}

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

func (m *MockImageRepository) GetAllWithPartner(ctx context.Context) ([]*ImageWithPartner, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ImageWithPartner), args.Error(1)
}

func (m *MockImageRepository) GetByStatusWithPartner(ctx context.Context, status string) ([]*ImageWithPartner, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ImageWithPartner), args.Error(1)
}

func (m *MockImageRepository) GetWithFilters(ctx context.Context, status, dateFrom, dateTo string) ([]*ImageWithPartner, error) {
	args := m.Called(ctx, status, dateFrom, dateTo)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ImageWithPartner), args.Error(1)
}

// Mock Coupon Repository
type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByCode(ctx context.Context, code string) (*Coupon, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Coupon), args.Error(1)
}

func (m *MockCouponRepository) Update(ctx context.Context, coupon *Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

// Mock S3 Client
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) UploadFile(ctx context.Context, data io.Reader, size int64, contentType, prefix string, couponID uuid.UUID) (string, error) {
	args := m.Called(ctx, data, size, contentType, prefix, couponID)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockS3Client) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockS3Client) GetFileURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	args := m.Called(ctx, key, expiry)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) GetSignedURL(key string, expires time.Duration) (string, error) {
	args := m.Called(key, expires)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) UploadFileWithKey(ctx context.Context, reader io.Reader, size int64, contentType string, objectKey string) (string, error) {
	args := m.Called(ctx, reader, size, contentType, objectKey)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) UploadPreviewFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, previewID uuid.UUID) (string, error) {
	args := m.Called(ctx, reader, size, contentType, folder, previewID)
	return args.String(0), args.Error(1)
}

// Mock Stable Diffusion Client
type MockStableDiffusionClient struct {
	mock.Mock
}

func (m *MockStableDiffusionClient) ProcessImage(ctx context.Context, req stableDiffusion.ProcessImageRequest) (string, error) {
	args := m.Called(ctx, req)
	return args.String(0), args.Error(1)
}

func (m *MockStableDiffusionClient) EncodeImageToBase64(data []byte) string {
	args := m.Called(data)
	return args.String(0)
}

func (m *MockStableDiffusionClient) DecodeBase64Image(base64Data string) ([]byte, error) {
	args := m.Called(base64Data)
	return args.Get(0).([]byte), args.Error(1)
}

// Mock Email Service
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendSchemaEmail(email, schemaURL, couponCode string) error {
	args := m.Called(email, schemaURL, couponCode)
	return args.Error(0)
}

// Mock Zip Service
type MockZipService struct {
	mock.Mock
}

func (m *MockZipService) CreateSchemaArchive(imageID uuid.UUID, files []zip.FileData) (*bytes.Buffer, error) {
	args := m.Called(imageID, files)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

// Mock Mosaic Generator
type MockMosaicGenerator struct {
	mock.Mock
}

func (m *MockMosaicGenerator) Generate(ctx context.Context, req *mosaic.GenerationRequest) (*mosaic.GenerationResult, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*mosaic.GenerationResult), args.Error(1)
}

func createTestImage() *Image {
	return &Image{
		ID:                 uuid.New(),
		CouponID:           uuid.New(),
		OriginalImageS3Key: "test/original.jpg",
		UserEmail:          "test@example.com",
		Status:             "uploaded",
		Priority:           1,
		ProcessingParams: &ProcessingParams{
			Style: "grayscale",
			UseAI: true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestCoupon() *coupon.Coupon {
	now := time.Now()
	return &coupon.Coupon{
		ID:        uuid.New(),
		Code:      "TEST123",
		Status:    "activated",
		Size:      "30x40",
		Style:     "grayscale",
		CreatedAt: now,
	}
}

func bytesToReadCloser(data []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(data))
}

type MockFileHeader struct {
	multipart.FileHeader
	content []byte
}

func (m *MockFileHeader) Open() (multipart.File, error) {
	return &MockFile{
		content: m.content,
		pos:     0,
	}, nil
}

type MockFile struct {
	content []byte
	pos     int64
}

func (m *MockFile) Read(p []byte) (n int, err error) {
	if m.pos >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.pos:])
	m.pos += int64(n)
	return n, nil
}

func (m *MockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	return n, nil
}

func (m *MockFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.pos = offset
	case io.SeekCurrent:
		m.pos += offset
	case io.SeekEnd:
		m.pos = int64(len(m.content)) + offset
	}
	return m.pos, nil
}

func (m *MockFile) Close() error {
	return nil
}

func TestImageService_GetQueue(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		mockSetup     func(*MockImageRepository)
		expectedError bool
		expectedCount int
	}{
		{
			name:   "successful_get_queue_with_status",
			status: "queued",
			mockSetup: func(repo *MockImageRepository) {
				images := []*ImageWithPartner{
					{Image: &Image{ID: uuid.New(), Status: "queued"}, PartnerID: uuid.New(), PartnerCode: "TEST1"},
					{Image: &Image{ID: uuid.New(), Status: "queued"}, PartnerID: uuid.New(), PartnerCode: "TEST2"},
				}
				repo.On("GetByStatusWithPartner", mock.Anything, "queued").Return(images, nil)
			},
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:   "successful_get_all_queue",
			status: "",
			mockSetup: func(repo *MockImageRepository) {
				images := []*ImageWithPartner{
					{Image: &Image{ID: uuid.New(), Status: "queued"}, PartnerID: uuid.New(), PartnerCode: "TEST1"},
					{Image: &Image{ID: uuid.New(), Status: "processing"}, PartnerID: uuid.New(), PartnerCode: "TEST2"},
					{Image: &Image{ID: uuid.New(), Status: "completed"}, PartnerID: uuid.New(), PartnerCode: "TEST3"},
				}
				repo.On("GetAllWithPartner", mock.Anything).Return(images, nil)
			},
			expectedError: false,
			expectedCount: 3,
		},
		{
			name:   "repository_error",
			status: "queued",
			mockSetup: func(repo *MockImageRepository) {
				repo.On("GetByStatusWithPartner", mock.Anything, "queued").Return(nil, errors.New("database error"))
			},
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:   "empty_queue",
			status: "queued",
			mockSetup: func(repo *MockImageRepository) {
				repo.On("GetByStatusWithPartner", mock.Anything, "queued").Return([]*ImageWithPartner{}, nil)
			},
			expectedError: false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockImageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			deps := &ImageServiceDeps{
				ImageRepository: mockRepo,
			}
			service := NewImageService(deps)

			result, err := service.GetQueue(tt.status)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestImageService_GetImageStatus(t *testing.T) {
	imageID := uuid.New()

	tests := []struct {
		name           string
		imageID        uuid.UUID
		mockSetup      func(*MockImageRepository, *MockS3Client)
		expectedError  bool
		expectedStatus string
	}{
		{
			name:    "successful_get_status",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository, s3Client *MockS3Client) {
				image := createTestImage()
				image.ID = imageID
				image.Status = "uploaded"
				repo.On("GetByID", mock.Anything, imageID).Return(image, nil)
				s3Client.On("GetFileURL", mock.Anything, image.OriginalImageS3Key, mock.AnythingOfType("time.Duration")).Return("http://test.com/original.jpg", nil)
			},
			expectedError:  false,
			expectedStatus: "uploaded",
		},
		{
			name:    "image_not_found",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository, s3Client *MockS3Client) {
				repo.On("GetByID", mock.Anything, imageID).Return(nil, errors.New("not found"))
			},
			expectedError:  true,
			expectedStatus: "",
		},
		{
			name:    "image_with_error_message",
			imageID: imageID,
			mockSetup: func(repo *MockImageRepository, s3Client *MockS3Client) {
				image := createTestImage()
				image.ID = imageID
				image.Status = "failed"
				errorMsg := "processing failed"
				image.ErrorMessage = &errorMsg
				repo.On("GetByID", mock.Anything, imageID).Return(image, nil)
				s3Client.On("GetFileURL", mock.Anything, image.OriginalImageS3Key, mock.AnythingOfType("time.Duration")).Return("http://test.com/original.jpg", nil)
			},
			expectedError:  false,
			expectedStatus: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockImageRepository)
			mockS3Client := new(MockS3Client)

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockS3Client)
			}

			deps := &ImageServiceDeps{
				ImageRepository: mockRepo,
				S3Client:        mockS3Client,
			}
			service := NewImageService(deps)

			result, err := service.GetImageStatus(context.Background(), tt.imageID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, imageID, result.ImageID)
				assert.Equal(t, tt.expectedStatus, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockS3Client.AssertExpectations(t)
		})
	}
}

func TestImageService_GetCouponRepository(t *testing.T) {
	deps := &ImageServiceDeps{
		CouponRepository: &MockCouponRepository{},
	}
	service := NewImageService(deps)

	result := service.GetCouponRepository()
	assert.NotNil(t, result)
	assert.Equal(t, deps.CouponRepository, result)
}

func TestImageService_GetS3Client(t *testing.T) {
	deps := &ImageServiceDeps{
		S3Client: &MockS3Client{},
	}
	service := NewImageService(deps)

	result := service.GetS3Client()
	assert.NotNil(t, result)
	assert.Equal(t, deps.S3Client, result)
}

func TestImageService_GetImageRepository(t *testing.T) {
	deps := &ImageServiceDeps{
		ImageRepository: &MockImageRepository{},
	}
	service := NewImageService(deps)

	result := service.GetImageRepository()
	assert.NotNil(t, result)
	assert.Equal(t, deps.ImageRepository, result)
}

func TestImageService_getStatusMessage(t *testing.T) {
	service := &ImageService{}

	tests := []struct {
		status   string
		expected string
	}{
		{"uploaded", "Изображение загружено, готово к редактированию"},
		{"edited", "Изображение отредактировано, готово к обработке"},
		{"processing", "Изображение обрабатывается..."},
		{"processed", "Изображение обработано, готово к созданию схемы"},
		{"completed", "Схема алмазной мозаики создана"},
		{"failed", "Произошла ошибка при обработке"},
		{"unknown", "Неизвестный статус"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := service.getStatusMessage(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImageService_calculateProgress(t *testing.T) {
	service := &ImageService{}

	tests := []struct {
		status   string
		expected int
	}{
		{"uploaded", 20},
		{"edited", 40},
		{"processing", 60},
		{"processed", 80},
		{"completed", 100},
		{"failed", 0},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			result := service.calculateProgress(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImageService_isValidImageType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{"valid_jpeg", "image/jpeg", true},
		{"valid_png", "image/png", true},
		{"invalid_gif", "image/gif", false},
		{"invalid_text", "text/plain", false},
		{"empty", "", false},
		{"case_insensitive_jpeg", "IMAGE/JPEG", false},
		{"case_insensitive_png", "IMAGE/PNG", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileHeader := &multipart.FileHeader{
				Header: map[string][]string{
					"Content-Type": {tt.contentType},
				},
			}
			result := isValidImageType(fileHeader)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImageService_parseCouponSize(t *testing.T) {
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
		{"", 1200, 1600},
		{"invalid", 1200, 1600},
	}

	for _, tt := range tests {
		t.Run(tt.size, func(t *testing.T) {
			width, height := parseCouponSize(tt.size)
			assert.Equal(t, tt.expectedWidth, width)
			assert.Equal(t, tt.expectedHeight, height)
		})
	}
}

func TestNewImageService(t *testing.T) {
	deps := &ImageServiceDeps{
		ImageRepository: &MockImageRepository{},
	}

	service := NewImageService(deps)

	assert.NotNil(t, service)
	assert.NotNil(t, service.deps)
	assert.Equal(t, deps, service.deps)
}

func TestImageService_EdgeCases(t *testing.T) {
	t.Run("nil_dependencies", func(t *testing.T) {
		service := NewImageService(nil)
		assert.NotNil(t, service)
		assert.Nil(t, service.deps)
	})

	t.Run("empty_status_messages", func(t *testing.T) {
		service := &ImageService{}

		result := service.getStatusMessage("")
		assert.Equal(t, "Неизвестный статус", result)

		progress := service.calculateProgress("")
		assert.Equal(t, 0, progress)
	})

	t.Run("large_coupon_sizes", func(t *testing.T) {
		width, height := parseCouponSize("100x100")
		assert.Equal(t, 1200, width)
		assert.Equal(t, 1600, height)
	})
}

func TestImageService_ProcessingParams(t *testing.T) {
	t.Run("valid_processing_params", func(t *testing.T) {
		params := ProcessingParams{
			Style:      "grayscale",
			UseAI:      true,
			Lighting:   "sun",
			Contrast:   "high",
			Brightness: 50.0,
			Saturation: -25.0,
			Settings: map[string]any{
				"custom": "value",
			},
		}

		assert.Equal(t, "grayscale", params.Style)
		assert.True(t, params.UseAI)
		assert.Equal(t, "sun", params.Lighting)
		assert.Equal(t, "high", params.Contrast)
		assert.Equal(t, 50.0, params.Brightness)
		assert.Equal(t, -25.0, params.Saturation)
		assert.Equal(t, "value", params.Settings["custom"])
	})

	t.Run("default_processing_params", func(t *testing.T) {
		params := ProcessingParams{
			Style: "grayscale",
			UseAI: false,
		}

		assert.Equal(t, "grayscale", params.Style)
		assert.False(t, params.UseAI)
		assert.Empty(t, params.Lighting)
		assert.Empty(t, params.Contrast)
		assert.Equal(t, 0.0, params.Brightness)
		assert.Equal(t, 0.0, params.Saturation)
		assert.Nil(t, params.Settings)
	})
}

func TestImageService_ImageEditParams(t *testing.T) {
	t.Run("valid_edit_params", func(t *testing.T) {
		params := ImageEditParams{
			CropX:      10,
			CropY:      20,
			CropWidth:  100,
			CropHeight: 150,
			Rotation:   90,
			Scale:      2.0,
		}

		assert.Equal(t, 10, params.CropX)
		assert.Equal(t, 20, params.CropY)
		assert.Equal(t, 100, params.CropWidth)
		assert.Equal(t, 150, params.CropHeight)
		assert.Equal(t, 90, params.Rotation)
		assert.Equal(t, 2.0, params.Scale)
	})

	t.Run("default_edit_params", func(t *testing.T) {
		params := ImageEditParams{}

		assert.Equal(t, 0, params.CropX)
		assert.Equal(t, 0, params.CropY)
		assert.Equal(t, 0, params.CropWidth)
		assert.Equal(t, 0, params.CropHeight)
		assert.Equal(t, 0, params.Rotation)
		assert.Equal(t, 0.0, params.Scale)
	})
}

func TestImageService_ImageStruct(t *testing.T) {
	t.Run("create_test_image", func(t *testing.T) {
		image := createTestImage()

		assert.NotNil(t, image)
		assert.NotEqual(t, uuid.Nil, image.ID)
		assert.NotEqual(t, uuid.Nil, image.CouponID)
		assert.Equal(t, "test/original.jpg", image.OriginalImageS3Key)
		assert.Equal(t, "test@example.com", image.UserEmail)
		assert.Equal(t, "uploaded", image.Status)
		assert.Equal(t, 1, image.Priority)
		assert.NotNil(t, image.ProcessingParams)
		assert.Equal(t, "grayscale", image.ProcessingParams.Style)
		assert.True(t, image.ProcessingParams.UseAI)
		assert.False(t, image.CreatedAt.IsZero())
		assert.False(t, image.UpdatedAt.IsZero())
	})

	t.Run("image_with_optional_fields", func(t *testing.T) {
		editedImageS3Key := "test/edited.jpg"
		image := &Image{
			EditedImageS3Key: &editedImageS3Key,
			Status:           "edited",
		}

		assert.NotNil(t, image.EditedImageS3Key)
		assert.Equal(t, "test/edited.jpg", *image.EditedImageS3Key)
		assert.Equal(t, "edited", image.Status)
	})
}

func TestImageService_CouponStruct(t *testing.T) {
	t.Run("create_test_coupon", func(t *testing.T) {
		coupon := createTestCoupon()

		assert.NotNil(t, coupon)
		assert.NotEqual(t, uuid.Nil, coupon.ID)
		assert.Equal(t, "TEST123", coupon.Code)
		assert.Equal(t, "activated", coupon.Status)
		assert.Equal(t, "30x40", coupon.Size)
		assert.Equal(t, "grayscale", coupon.Style)
		assert.False(t, coupon.CreatedAt.IsZero())
	})

	t.Run("coupon_with_user_email", func(t *testing.T) {
		coupon := createTestCoupon()
		email := "user@example.com"
		coupon.UserEmail = &email

		assert.NotNil(t, coupon.UserEmail)
		assert.Equal(t, "user@example.com", *coupon.UserEmail)
	})
}

func TestImageService_HelperFunctions(t *testing.T) {
	t.Run("bytesToReadCloser", func(t *testing.T) {
		data := []byte("test data")
		reader := bytesToReadCloser(data)
		assert.NotNil(t, reader)

		readData, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, data, readData)

		err = reader.Close()
		assert.NoError(t, err)
	})
}

func TestImageService_MockInterfaces(t *testing.T) {
	t.Run("mock_image_repository", func(t *testing.T) {
		mockRepo := new(MockImageRepository)
		assert.NotNil(t, mockRepo)

		ctx := context.Background()
		imageID := uuid.New()

		mockRepo.On("GetByID", ctx, imageID).Return(nil, errors.New("not found"))

		result, err := mockRepo.GetByID(ctx, imageID)
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, "not found", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("mock_coupon_repository", func(t *testing.T) {
		mockRepo := new(MockCouponRepository)
		assert.NotNil(t, mockRepo)

		ctx := context.Background()
		couponID := uuid.New()

		mockRepo.On("GetByID", ctx, couponID).Return(nil, errors.New("not found"))

		result, err := mockRepo.GetByID(ctx, couponID)
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Equal(t, "not found", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("mock_s3_client", func(t *testing.T) {
		mockS3 := new(MockS3Client)
		assert.NotNil(t, mockS3)

		ctx := context.Background()
		key := "test/key"

		mockS3.On("GetFileURL", ctx, key, mock.AnythingOfType("time.Duration")).Return("http://test.com/file", nil)

		result, err := mockS3.GetFileURL(ctx, key, 24*time.Hour)
		assert.NoError(t, err)
		assert.Equal(t, "http://test.com/file", result)

		mockS3.AssertExpectations(t)
	})
}
