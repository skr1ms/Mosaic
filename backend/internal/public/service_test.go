package public

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
)

type MockCouponRepository struct {
	mock.Mock
}

func (m *MockCouponRepository) Create(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
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

func (m *MockCouponRepository) Update(ctx context.Context, coupon *coupon.Coupon) error {
	args := m.Called(ctx, coupon)
	return args.Error(0)
}

func (m *MockCouponRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCouponRepository) GetAll(ctx context.Context) ([]*coupon.Coupon, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error) {
	args := m.Called(ctx, partnerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*coupon.Coupon), args.Error(1)
}

func (m *MockCouponRepository) CodeExists(ctx context.Context, code string) (bool, error) {
	args := m.Called(ctx, code)
	return args.Bool(0), args.Error(1)
}

type MockImageRepository struct {
	mock.Mock
}

func (m *MockImageRepository) Create(ctx context.Context, image *image.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockImageRepository) GetByID(ctx context.Context, id uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) Update(ctx context.Context, image *image.Image) error {
	args := m.Called(ctx, image)
	return args.Error(0)
}

func (m *MockImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImageRepository) GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error) {
	args := m.Called(ctx, couponID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageRepository) GetQueuedTasks(ctx context.Context) ([]*image.Image, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*image.Image), args.Error(1)
}

type MockPartnerRepository struct {
	mock.Mock
}

func (m *MockPartnerRepository) Create(ctx context.Context, partner *partner.Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error) {
	args := m.Called(ctx, id)
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

func (m *MockPartnerRepository) Update(ctx context.Context, partner *partner.Partner) error {
	args := m.Called(ctx, partner)
	return args.Error(0)
}

func (m *MockPartnerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartnerRepository) GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error) {
	args := m.Called(ctx, sortBy, order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*partner.Partner), args.Error(1)
}

type MockImageService struct {
	mock.Mock
}

func (m *MockImageService) UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*image.Image, error) {
	args := m.Called(ctx, couponID, file, userEmail)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*image.Image), args.Error(1)
}

func (m *MockImageService) EditImage(ctx context.Context, imageID uuid.UUID, params image.ImageEditParams) error {
	args := m.Called(ctx, imageID, params)
	return args.Error(0)
}

func (m *MockImageService) ProcessImage(ctx context.Context, imageID uuid.UUID, params image.ProcessingParams) error {
	args := m.Called(ctx, imageID, params)
	return args.Error(0)
}

func (m *MockImageService) GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error {
	args := m.Called(ctx, imageID, confirmed)
	return args.Error(0)
}

func (m *MockImageService) GetImageStatus(ctx context.Context, imageID uuid.UUID) (*types.ImageStatusResponse, error) {
	args := m.Called(ctx, imageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ImageStatusResponse), args.Error(1)
}

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) PurchaseCoupon(ctx context.Context, req *payment.PurchaseCouponRequest) (*payment.PurchaseCouponResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payment.PurchaseCouponResponse), args.Error(1)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendSchemaEmail(email, schemaURL, couponCode string) error {
	args := m.Called(email, schemaURL, couponCode)
	return args.Error(0)
}

// Helper functions for creating test data

func createTestCoupon() *coupon.Coupon {
	return &coupon.Coupon{
		ID:     uuid.New(),
		Code:   "123456789012",
		Size:   "40x50",
		Style:  "grayscale",
		Status: "new",
	}
}

func createTestActivatedCoupon() *coupon.Coupon {
	email := "test@example.com"
	activatedAt := time.Now()
	return &coupon.Coupon{
		ID:          uuid.New(),
		Code:        "123456789012",
		Size:        "40x50",
		Style:       "grayscale",
		Status:      "activated",
		UserEmail:   &email,
		ActivatedAt: &activatedAt,
	}
}

func createTestPartner() *partner.Partner {
	return &partner.Partner{
		ID:             uuid.New(),
		BrandName:      "Test Partner",
		Domain:         "test.example.com",
		LogoURL:        "https://example.com/logo.png",
		Email:          "contact@test.example.com",
		Phone:          "+1234567890",
		Address:        "123 Test Street",
		AllowPurchases: true,
	}
}

func createTestImage() *image.Image {
	return &image.Image{
		ID:       uuid.New(),
		CouponID: uuid.New(),
		Status:   "uploaded",
	}
}

func createTestConfig() *config.Config {
	return &config.Config{
		ServerConfig: config.ServerConfig{
			PaymentSuccessURL: "https://example.com/success",
		},
	}
}

func TestPublicService_GetPartnerByDomain_Success(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}
	deps := &PublicServiceDeps{
		PartnerRepository: mockPartnerRepo,
	}
	service := NewPublicService(deps)

	testPartner := createTestPartner()
	mockPartnerRepo.On("GetByDomain", mock.Anything, "test.example.com").Return(testPartner, nil)

	result, err := service.GetPartnerByDomain("test.example.com")

	require.NoError(t, err)
	assert.Equal(t, testPartner.BrandName, result["brand_name"])
	assert.Equal(t, testPartner.Domain, result["domain"])
	assert.Equal(t, testPartner.Email, result["email"])
	assert.Equal(t, testPartner.AllowPurchases, result["allow_purchases"])

	mockPartnerRepo.AssertExpectations(t)
}

func TestPublicService_GetPartnerByDomain_NotFound(t *testing.T) {
	mockPartnerRepo := &MockPartnerRepository{}
	deps := &PublicServiceDeps{
		PartnerRepository: mockPartnerRepo,
	}
	service := NewPublicService(deps)

	mockPartnerRepo.On("GetByDomain", mock.Anything, "nonexistent.example.com").Return(nil, errors.New("not found"))

	result, err := service.GetPartnerByDomain("nonexistent.example.com")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get partner by domain")

	mockPartnerRepo.AssertExpectations(t)
}

func TestPublicService_GetCouponByCode_Success(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	mockCouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(testCoupon, nil)

	result, err := service.GetCouponByCode("123456789012")

	require.NoError(t, err)
	assert.Equal(t, testCoupon.ID, result["id"])
	assert.Equal(t, testCoupon.Code, result["code"])
	assert.Equal(t, testCoupon.Size, result["size"])
	assert.Equal(t, testCoupon.Style, result["style"])
	assert.Equal(t, testCoupon.Status, result["status"])
	assert.Equal(t, true, result["valid"])

	mockCouponRepo.AssertExpectations(t)
}

func TestPublicService_GetCouponByCode_InvalidFormat(t *testing.T) {
	deps := &PublicServiceDeps{}
	service := NewPublicService(deps)

	testCases := []struct {
		name string
		code string
	}{
		{"too short", "12345"},
		{"too long", "1234567890123"},
		{"contains letters", "12345678901a"},
		{"empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := service.GetCouponByCode(tc.code)

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "invalid coupon code")
		})
	}
}

func TestPublicService_GetCouponByCode_NotFound(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	mockCouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(nil, errors.New("not found"))

	result, err := service.GetCouponByCode("123456789012")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coupon not found")

	mockCouponRepo.AssertExpectations(t)
}

func TestPublicService_ActivateCoupon_Success(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	req := ActivateCouponRequest{
		Email: "test@example.com",
	}

	mockCouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(testCoupon, nil)
	mockCouponRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *coupon.Coupon) bool {
		return c.Status == "activated" && c.UserEmail != nil && *c.UserEmail == "test@example.com"
	})).Return(nil)

	result, err := service.ActivateCoupon("123456789012", req)

	require.NoError(t, err)
	assert.Equal(t, "Купон успешно активирован", result["message"])
	assert.Equal(t, testCoupon.ID, result["coupon_id"])
	assert.Equal(t, "upload_image", result["next_step"])

	mockCouponRepo.AssertExpectations(t)
}

func TestPublicService_ActivateCoupon_CouponNotFound(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	req := ActivateCouponRequest{
		Email: "test@example.com",
	}

	mockCouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(nil, errors.New("not found"))

	result, err := service.ActivateCoupon("123456789012", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coupon not found")

	mockCouponRepo.AssertExpectations(t)
}

func TestPublicService_ActivateCoupon_AlreadyUsed(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	testCoupon.Status = "activated"
	req := ActivateCouponRequest{
		Email: "test@example.com",
	}

	mockCouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(testCoupon, nil)

	result, err := service.ActivateCoupon("123456789012", req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "coupon already used")

	mockCouponRepo.AssertExpectations(t)
}

func TestPublicService_UploadImage_Success(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	mockImageService := &MockImageService{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
		ImageService:     mockImageService,
	}
	service := NewPublicService(deps)

	testCoupon := createTestActivatedCoupon()
	testImage := createTestImage()

	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     1024,
	}

	mockCouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil)
	mockImageService.On("UploadImage", mock.Anything, testCoupon.ID, file, *testCoupon.UserEmail).Return(testImage, nil)

	// Act
	result, err := service.UploadImage(testCoupon.ID.String(), file)

	// Assert
	require.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "Изображение успешно загружено", result["message"])
	assert.Equal(t, testImage.ID, result["image_id"])
	assert.Equal(t, "edit_image", result["next_step"])
	assert.Equal(t, testCoupon.Size, result["coupon_size"])
	assert.Equal(t, testCoupon.Style, result["coupon_style"])

	mockCouponRepo.AssertExpectations(t)
	mockImageService.AssertExpectations(t)
}

func TestPublicService_UploadImage_InvalidCouponID(t *testing.T) {
	deps := &PublicServiceDeps{}
	service := NewPublicService(deps)

	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     1024,
	}

	result, err := service.UploadImage("invalid-uuid", file)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid coupon id")
}

func TestPublicService_PurchaseCoupon_Success(t *testing.T) {
	mockCouponRepo := &MockCouponRepository{}
	mockPaymentService := &MockPaymentService{}
	deps := &PublicServiceDeps{
		CouponRepository: mockCouponRepo,
		PaymentService:   mockPaymentService,
		Config:           createTestConfig(),
	}
	service := NewPublicService(deps)

	req := PurchaseCouponRequest{
		Size:         "40x50",
		Style:        "grayscale",
		Email:        "test@example.com",
		PaymentToken: "test-token",
	}

	paymentResponse := &payment.PurchaseCouponResponse{
		Success:    true,
		OrderID:    "order123",
		PaymentURL: "https://payment.example.com/order123",
		Message:    "Success",
	}

	mockCouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	mockCouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockPaymentService.On("PurchaseCoupon", mock.Anything, mock.AnythingOfType("*payment.PurchaseCouponRequest")).Return(paymentResponse, nil)

	result, err := service.PurchaseCoupon(req)

	require.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "order123", result["order_id"])
	assert.Equal(t, "https://payment.example.com/order123", result["payment_url"])

	mockCouponRepo.AssertExpectations(t)
	mockPaymentService.AssertExpectations(t)
}

func TestPublicService_GetAvailableSizes(t *testing.T) {
	deps := &PublicServiceDeps{}
	service := NewPublicService(deps)

	sizes := service.GetAvailableSizes()

	assert.NotEmpty(t, sizes)
	assert.Len(t, sizes, 6)

	firstSize := sizes[0]
	assert.Contains(t, firstSize, "size")
	assert.Contains(t, firstSize, "title")
	assert.Contains(t, firstSize, "price")

	sizeValues := []string{}
	for _, size := range sizes {
		sizeValues = append(sizeValues, size["size"].(string))
	}

	expectedSizes := []string{"21x30", "30x40", "40x40", "40x50", "40x60", "50x70"}
	for _, expectedSize := range expectedSizes {
		assert.Contains(t, sizeValues, expectedSize)
	}
}

func TestPublicService_GetAvailableStyles(t *testing.T) {
	deps := &PublicServiceDeps{}
	service := NewPublicService(deps)

	styles := service.GetAvailableStyles()

	assert.NotEmpty(t, styles)
	assert.Len(t, styles, 4)

	firstStyle := styles[0]
	assert.Contains(t, firstStyle, "style")
	assert.Contains(t, firstStyle, "title")
	assert.Contains(t, firstStyle, "description")

	styleValues := []string{}
	for _, style := range styles {
		styleValues = append(styleValues, style["style"].(string))
	}

	expectedStyles := []string{"grayscale", "skin_tones", "pop_art", "max_colors"}
	for _, expectedStyle := range expectedStyles {
		assert.Contains(t, styleValues, expectedStyle)
	}
}

func TestPublicService_EditImage_Success(t *testing.T) {
	mockImageService := &MockImageService{}
	deps := &PublicServiceDeps{
		ImageService: mockImageService,
	}
	service := NewPublicService(deps)

	imageID := uuid.New()
	req := types.EditImageRequest{
		CropX:      10,
		CropY:      10,
		CropWidth:  100,
		CropHeight: 100,
		Rotation:   90,
		Scale:      1.5,
	}

	statusResponse := &types.ImageStatusResponse{
		EditedURL: stringPtr("https://example.com/edited.jpg"),
	}

	mockImageService.On("EditImage", mock.Anything, imageID, mock.AnythingOfType("image.ImageEditParams")).Return(nil)
	mockImageService.On("GetImageStatus", mock.Anything, imageID).Return(statusResponse, nil)

	result, err := service.EditImage(imageID.String(), req)

	require.NoError(t, err)
	assert.Equal(t, "Изображение успешно отредактировано", result["message"])
	assert.Equal(t, "choose_style", result["next_step"])
	assert.Equal(t, "https://example.com/edited.jpg", result["preview_url"])

	mockImageService.AssertExpectations(t)
}

func TestPublicService_ProcessImage_Success(t *testing.T) {
	mockImageService := &MockImageService{}
	deps := &PublicServiceDeps{
		ImageService: mockImageService,
	}
	service := NewPublicService(deps)

	imageID := uuid.New()
	req := types.ProcessImageRequest{
		Style:      "grayscale",
		UseAI:      false,
		Lighting:   "sun",
		Contrast:   "high",
		Brightness: 10.0,
		Saturation: 5.0,
	}

	statusResponse := &types.ImageStatusResponse{
		PreviewURL:  stringPtr("https://example.com/preview.jpg"),
		OriginalURL: stringPtr("https://example.com/original.jpg"),
	}

	// ProcessImage вызывается асинхронно, поэтому не можем его моккировать
	// mockImageService.On("ProcessImage", mock.Anything, imageID, mock.AnythingOfType("image.ProcessingParams")).Return(nil)
	mockImageService.On("GetImageStatus", mock.Anything, imageID).Return(statusResponse, nil)

	result, err := service.ProcessImage(imageID.String(), req)

	require.NoError(t, err)
	assert.Equal(t, "Обработка запущена", result["message"])
	assert.Equal(t, "generate_schema", result["next_step"])
	assert.Equal(t, "https://example.com/preview.jpg", result["preview_url"])
	assert.Equal(t, "https://example.com/original.jpg", result["original_url"])

	mockImageService.AssertExpectations(t)
}

func TestPublicService_GetImagePreview_Success(t *testing.T) {
	mockImageRepo := &MockImageRepository{}
	mockImageService := &MockImageService{}
	deps := &PublicServiceDeps{
		ImageRepository: mockImageRepo,
		ImageService:    mockImageService,
	}
	service := NewPublicService(deps)

	testImage := createTestImage()
	statusResponse := &types.ImageStatusResponse{
		PreviewURL:  stringPtr("https://example.com/preview.jpg"),
		OriginalURL: stringPtr("https://example.com/original.jpg"),
	}

	mockImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	mockImageService.On("GetImageStatus", mock.Anything, testImage.ID).Return(statusResponse, nil)

	result, err := service.GetImagePreview(testImage.ID.String())

	require.NoError(t, err)
	assert.Equal(t, testImage.ID, result["id"])
	assert.Equal(t, testImage.Status, result["status"])
	assert.Equal(t, "https://example.com/preview.jpg", (*result["preview_url"].(*string)))
	assert.Equal(t, "https://example.com/original.jpg", (*result["original_url"].(*string)))

	mockImageRepo.AssertExpectations(t)
	mockImageService.AssertExpectations(t)
}

func TestPublicService_GetProcessingStatus_Success(t *testing.T) {
	mockImageRepo := &MockImageRepository{}
	deps := &PublicServiceDeps{
		ImageRepository: mockImageRepo,
	}
	service := NewPublicService(deps)

	testImage := createTestImage()
	testImage.Status = "processing"
	startedAt := time.Now()
	testImage.StartedAt = &startedAt

	mockImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)

	result, err := service.GetProcessingStatus(testImage.ID.String())

	require.NoError(t, err)
	assert.Equal(t, testImage.ID, result["id"])
	assert.Equal(t, "processing", result["status"])
	assert.Equal(t, 50, result["progress"])

	mockImageRepo.AssertExpectations(t)
}

func TestPublicService_SendSchemaToEmail_Success(t *testing.T) {
	mockImageRepo := &MockImageRepository{}
	mockCouponRepo := &MockCouponRepository{}
	mockImageService := &MockImageService{}
	mockEmailService := &MockEmailService{}
	deps := &PublicServiceDeps{
		ImageRepository:  mockImageRepo,
		CouponRepository: mockCouponRepo,
		ImageService:     mockImageService,
		EmailService:     mockEmailService,
	}
	service := NewPublicService(deps)

	testImage := createTestImage()
	testImage.Status = "completed"
	schemaKey := "schema.zip"
	testImage.SchemaS3Key = &schemaKey

	testCoupon := createTestCoupon()

	statusResponse := &types.ImageStatusResponse{
		SchemaURL: stringPtr("https://example.com/schema.zip"),
	}

	req := SendEmailRequest{
		Email: "test@example.com",
	}

	mockImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	mockCouponRepo.On("GetByID", mock.Anything, testImage.CouponID).Return(testCoupon, nil)
	mockImageService.On("GetImageStatus", mock.Anything, testImage.ID).Return(statusResponse, nil)
	mockEmailService.On("SendSchemaEmail", "test@example.com", "https://example.com/schema.zip", testCoupon.Code).Return(nil)

	result, err := service.SendSchemaToEmail(testImage.ID.String(), req)

	require.NoError(t, err)
	assert.Equal(t, "Schema successfully sent to email", result["message"])

	mockImageRepo.AssertExpectations(t)
	mockCouponRepo.AssertExpectations(t)
	mockImageService.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func stringPtr(s string) *string {
	return &s
}

func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"123456789012", true},
		{"000000000000", true},
		{"12345a789012", false},
		{"", false},
		{"12345", false},
		{"abc", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isNumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
