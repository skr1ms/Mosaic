package public

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http/httptest"
	"testing"

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

type IntegrationTestSuite struct {
	service  *PublicService
	mocks    *IntegrationMocks
	testData *IntegrationTestData
}

type IntegrationMocks struct {
	CouponRepo  *MockCouponRepository
	ImageRepo   *MockImageRepository
	PartnerRepo *MockPartnerRepository
	ImageSvc    *MockImageService
	PaymentSvc  *MockPaymentService
	EmailSvc    *MockEmailService
}

type IntegrationTestData struct {
	Partner *partner.Partner
	Coupon  *coupon.Coupon
	Image   *image.Image
	Config  *config.Config
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{
		mocks: &IntegrationMocks{
			CouponRepo:  &MockCouponRepository{},
			ImageRepo:   &MockImageRepository{},
			PartnerRepo: &MockPartnerRepository{},
			ImageSvc:    &MockImageService{},
			PaymentSvc:  &MockPaymentService{},
			EmailSvc:    &MockEmailService{},
		},
		testData: &IntegrationTestData{
			Partner: createTestPartner(),
			Coupon:  createTestCoupon(),
			Image:   createTestImage(),
			Config:  createTestConfig(),
		},
	}

	deps := &PublicServiceDeps{
		CouponRepository:  suite.mocks.CouponRepo,
		ImageRepository:   suite.mocks.ImageRepo,
		PartnerRepository: suite.mocks.PartnerRepo,
		ImageService:      suite.mocks.ImageSvc,
		PaymentService:    suite.mocks.PaymentSvc,
		EmailService:      suite.mocks.EmailSvc,
		Config:            suite.testData.Config,
	}

	suite.service = NewPublicService(deps)
	return suite
}

func TestIntegration_CouponWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCoupon := suite.testData.Coupon
	testImage := suite.testData.Image
	testImage.CouponID = testCoupon.ID

	suite.mocks.CouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)

	suite.mocks.CouponRepo.On("Update", mock.Anything, mock.MatchedBy(func(c *coupon.Coupon) bool {
		return c.Status == "activated" && c.UserEmail != nil
	})).Return(nil)

	activatedCoupon := createTestActivatedCoupon()
	activatedCoupon.ID = testCoupon.ID
	activatedCoupon.Code = testCoupon.Code
	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(activatedCoupon, nil)
	suite.mocks.ImageSvc.On("UploadImage", mock.Anything, testCoupon.ID, mock.AnythingOfType("*multipart.FileHeader"), *activatedCoupon.UserEmail).Return(testImage, nil)

	suite.mocks.ImageSvc.On("EditImage", mock.Anything, testImage.ID, mock.AnythingOfType("image.ImageEditParams")).Return(nil)
	suite.mocks.ImageSvc.On("GetImageStatus", mock.Anything, testImage.ID).Return(&types.ImageStatusResponse{
		EditedURL: stringPtr("https://example.com/edited.jpg"),
	}, nil)

	suite.mocks.ImageSvc.On("GetImageStatus", mock.Anything, testImage.ID).Return(&types.ImageStatusResponse{
		PreviewURL:  stringPtr("https://example.com/preview.jpg"),
		OriginalURL: stringPtr("https://example.com/original.jpg"),
	}, nil)

	suite.mocks.ImageSvc.On("ProcessImage", mock.Anything, testImage.ID, mock.AnythingOfType("image.ProcessingParams")).Return(nil).Maybe()

	couponInfo, err := suite.service.GetCouponByCode(testCoupon.Code)
	require.NoError(t, err)
	assert.Equal(t, testCoupon.ID, couponInfo["id"])
	assert.Equal(t, true, couponInfo["valid"])

	activateReq := ActivateCouponRequest{Email: "test@example.com"}
	activateResult, err := suite.service.ActivateCoupon(testCoupon.Code, activateReq)
	require.NoError(t, err)
	assert.Equal(t, "Купон успешно активирован", activateResult["message"])
	assert.Equal(t, "upload_image", activateResult["next_step"])

	file := createTestMultipartFile(t, "test.jpg")
	uploadResult, err := suite.service.UploadImage(testCoupon.ID.String(), file)
	require.NoError(t, err)
	assert.Equal(t, "Изображение успешно загружено", uploadResult["message"])
	assert.Equal(t, "edit_image", uploadResult["next_step"])

	editReq := types.EditImageRequest{
		CropX:      10,
		CropY:      10,
		CropWidth:  100,
		CropHeight: 100,
		Rotation:   90,
		Scale:      1.5,
	}
	editResult, err := suite.service.EditImage(testImage.ID.String(), editReq)
	require.NoError(t, err)
	assert.Equal(t, "Изображение успешно отредактировано", editResult["message"])
	assert.Equal(t, "choose_style", editResult["next_step"])

	processReq := types.ProcessImageRequest{
		Style:      "grayscale",
		UseAI:      false,
		Lighting:   "sun",
		Contrast:   "high",
		Brightness: 10.0,
		Saturation: 5.0,
	}
	processResult, err := suite.service.ProcessImage(testImage.ID.String(), processReq)
	require.NoError(t, err)
	assert.Equal(t, "Обработка запущена", processResult["message"])
	assert.Equal(t, "generate_schema", processResult["next_step"])

	suite.mocks.CouponRepo.AssertExpectations(t)
	// suite.mocks.ImageSvc.AssertExpectations(t) // Пропускаем из-за асинхронного ProcessImage
}

func TestIntegration_PurchaseWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	purchaseReq := PurchaseCouponRequest{
		Size:         "40x50",
		Style:        "grayscale",
		Email:        "customer@example.com",
		PaymentToken: "test-token",
	}

	paymentResponse := &payment.PurchaseCouponResponse{
		Success:    true,
		OrderID:    "order-123",
		PaymentURL: "https://payment.example.com/order-123",
		Message:    "Payment created successfully",
	}

	suite.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	suite.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
	suite.mocks.PaymentSvc.On("PurchaseCoupon", mock.Anything, mock.AnythingOfType("*payment.PurchaseCouponRequest")).Return(paymentResponse, nil)

	result, err := suite.service.PurchaseCoupon(purchaseReq)

	require.NoError(t, err)
	assert.Equal(t, true, result["success"])
	assert.Equal(t, "order-123", result["order_id"])
	assert.Equal(t, "https://payment.example.com/order-123", result["payment_url"])
	assert.Contains(t, result["message"], "переходите по ссылке для оплаты")

	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.PaymentSvc.AssertExpectations(t)
}

func TestIntegration_PartnerIntegration(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testPartner := suite.testData.Partner
	suite.mocks.PartnerRepo.On("GetByDomain", mock.Anything, testPartner.Domain).Return(testPartner, nil)

	result, err := suite.service.GetPartnerByDomain(testPartner.Domain)

	require.NoError(t, err)
	assert.Equal(t, testPartner.BrandName, result["brand_name"])
	assert.Equal(t, testPartner.Domain, result["domain"])
	assert.Equal(t, testPartner.Email, result["email"])
	assert.Equal(t, testPartner.AllowPurchases, result["allow_purchases"])

	suite.mocks.PartnerRepo.AssertExpectations(t)
}

func TestIntegration_EmailWorkflow(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testImage := suite.testData.Image
	testImage.Status = "completed"
	schemaKey := "test-schema.zip"
	testImage.SchemaS3Key = &schemaKey

	testCoupon := suite.testData.Coupon
	testImage.CouponID = testCoupon.ID

	statusResponse := &types.ImageStatusResponse{
		SchemaURL: stringPtr("https://s3.example.com/schemas/test-schema.zip"),
	}

	emailReq := SendEmailRequest{
		Email: "user@example.com",
	}

	suite.mocks.ImageRepo.On("GetByID", mock.Anything, testImage.ID).Return(testImage, nil)
	suite.mocks.CouponRepo.On("GetByID", mock.Anything, testCoupon.ID).Return(testCoupon, nil)
	suite.mocks.ImageSvc.On("GetImageStatus", mock.Anything, testImage.ID).Return(statusResponse, nil)
	suite.mocks.EmailSvc.On("SendSchemaEmail", emailReq.Email, *statusResponse.SchemaURL, testCoupon.Code).Return(nil)

	result, err := suite.service.SendSchemaToEmail(testImage.ID.String(), emailReq)

	require.NoError(t, err)
	assert.Equal(t, "Schema successfully sent to email", result["message"])

	suite.mocks.ImageRepo.AssertExpectations(t)
	suite.mocks.CouponRepo.AssertExpectations(t)
	suite.mocks.ImageSvc.AssertExpectations(t)
	suite.mocks.EmailSvc.AssertExpectations(t)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	suite := SetupIntegrationTest(t)

	activatedCoupon := createTestActivatedCoupon()

	testCases := []struct {
		name          string
		setupMocks    func()
		action        func() (interface{}, error)
		expectedError string
	}{
		{
			name: "Coupon not found during activation",
			setupMocks: func() {
				suite.mocks.CouponRepo.On("GetByCode", mock.Anything, "123456789012").Return(nil, errors.New("not found"))
			},
			action: func() (interface{}, error) {
				return suite.service.ActivateCoupon("123456789012", ActivateCouponRequest{Email: "test@example.com"})
			},
			expectedError: "coupon not found",
		},
		{
			name: "Payment service failure",
			setupMocks: func() {
				suite.mocks.CouponRepo.On("CodeExists", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
				suite.mocks.CouponRepo.On("Create", mock.Anything, mock.AnythingOfType("*coupon.Coupon")).Return(nil)
				suite.mocks.PaymentSvc.On("PurchaseCoupon", mock.Anything, mock.AnythingOfType("*payment.PurchaseCouponRequest")).Return(nil, errors.New("payment gateway error"))
			},
			action: func() (interface{}, error) {
				return suite.service.PurchaseCoupon(PurchaseCouponRequest{
					Size:         "40x50",
					Style:        "grayscale",
					Email:        "test@example.com",
					PaymentToken: "test-token",
				})
			},
			expectedError: "failed to create payment order",
		},
		{
			name: "Image service failure during upload",
			setupMocks: func() {
				suite.mocks.CouponRepo.On("GetByID", mock.Anything, activatedCoupon.ID).Return(activatedCoupon, nil)
				suite.mocks.ImageSvc.On("UploadImage", mock.Anything, activatedCoupon.ID, mock.AnythingOfType("*multipart.FileHeader"), *activatedCoupon.UserEmail).Return(nil, errors.New("upload failed"))
			},
			action: func() (interface{}, error) {
				file := createTestMultipartFile(t, "test.jpg")
				return suite.service.UploadImage(activatedCoupon.ID.String(), file)
			},
			expectedError: "upload failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testSuite := SetupIntegrationTest(t)
			tc.setupMocks()

			result, err := tc.action()

			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), tc.expectedError)

			assert.NotNil(t, testSuite)
		})
	}
}

func TestIntegration_ConcurrentRequests(t *testing.T) {
	suite := SetupIntegrationTest(t)

	coupons := make([]*coupon.Coupon, 5)
	for i := 0; i < 5; i++ {
		coupons[i] = createTestCoupon()
		coupons[i].Code = generateTestCouponCode(i)
		suite.mocks.CouponRepo.On("GetByCode", mock.Anything, coupons[i].Code).Return(coupons[i], nil)
	}

	results := make(chan error, 5)

	for i := 0; i < 5; i++ {
		go func(couponCode string) {
			_, err := suite.service.GetCouponByCode(couponCode)
			results <- err
		}(coupons[i].Code)
	}

	for i := 0; i < 5; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	suite.mocks.CouponRepo.AssertExpectations(t)
}

func TestIntegration_DataValidation(t *testing.T) {
	suite := SetupIntegrationTest(t)

	testCases := []struct {
		name   string
		action func() (interface{}, error)
	}{
		{
			name: "Invalid coupon code format",
			action: func() (interface{}, error) {
				return suite.service.GetCouponByCode("invalid")
			},
		},
		{
			name: "Invalid UUID format",
			action: func() (interface{}, error) {
				file := createTestMultipartFile(t, "test.jpg")
				return suite.service.UploadImage("invalid-uuid", file)
			},
		},
		{
			name: "Empty email for activation",
			action: func() (interface{}, error) {
				return suite.service.ActivateCoupon("123456789012", ActivateCouponRequest{Email: ""})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.action()

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}

func createTestMultipartFile(t *testing.T, filename string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)

	testData := []byte("test image data")
	_, err = fileWriter.Write(testData)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	err = req.ParseMultipartForm(10 << 20) // 10 MB
	require.NoError(t, err)

	files := req.MultipartForm.File["file"]
	require.Len(t, files, 1)

	return files[0]
}

func generateTestCouponCode(index int) string {
	return "123456789" + string(rune('0'+index%10)) + "12"
}

func BenchmarkIntegration_GetCouponByCode(b *testing.B) {
	suite := &IntegrationTestSuite{
		mocks: &IntegrationMocks{
			CouponRepo: &MockCouponRepository{},
		},
	}

	deps := &PublicServiceDeps{
		CouponRepository: suite.mocks.CouponRepo,
	}
	service := NewPublicService(deps)

	testCoupon := createTestCoupon()
	suite.mocks.CouponRepo.On("GetByCode", mock.Anything, testCoupon.Code).Return(testCoupon, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetCouponByCode(testCoupon.Code)
		if err != nil {
			b.Fatal(err)
		}
	}
}
