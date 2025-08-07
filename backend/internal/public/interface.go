package public

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/coupon"
	"github.com/skr1ms/mosaic/internal/image"
	"github.com/skr1ms/mosaic/internal/partner"
	"github.com/skr1ms/mosaic/internal/payment"
	"github.com/skr1ms/mosaic/internal/types"
)

type CouponRepositoryInterface interface {
	Create(ctx context.Context, coupon *coupon.Coupon) error
	GetByID(ctx context.Context, id uuid.UUID) (*coupon.Coupon, error)
	GetByCode(ctx context.Context, code string) (*coupon.Coupon, error)
	Update(ctx context.Context, coupon *coupon.Coupon) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*coupon.Coupon, error)
	GetByPartnerID(ctx context.Context, partnerID uuid.UUID) ([]*coupon.Coupon, error)
	CodeExists(ctx context.Context, code string) (bool, error)
}

type ImageRepositoryInterface interface {
	Create(ctx context.Context, image *image.Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*image.Image, error)
	Update(ctx context.Context, image *image.Image) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCouponID(ctx context.Context, couponID uuid.UUID) (*image.Image, error)
	GetQueuedTasks(ctx context.Context) ([]*image.Image, error)
}

type PartnerRepositoryInterface interface {
	Create(ctx context.Context, partner *partner.Partner) error
	GetByID(ctx context.Context, id uuid.UUID) (*partner.Partner, error)
	GetByDomain(ctx context.Context, domain string) (*partner.Partner, error)
	Update(ctx context.Context, partner *partner.Partner) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context, sortBy string, order string) ([]*partner.Partner, error)
}

type ImageServiceInterface interface {
	UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*image.Image, error)
	EditImage(ctx context.Context, imageID uuid.UUID, params image.ImageEditParams) error
	ProcessImage(ctx context.Context, imageID uuid.UUID, params image.ProcessingParams) error
	GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error
	GetImageStatus(ctx context.Context, imageID uuid.UUID) (*types.ImageStatusResponse, error)
}

type PaymentServiceInterface interface {
	PurchaseCoupon(ctx context.Context, req *payment.PurchaseCouponRequest) (*payment.PurchaseCouponResponse, error)
}

type EmailServiceInterface interface {
	SendSchemaEmail(email, schemaURL, couponCode string) error
}

type ConfigInterface interface {
	GetServerConfig() config.ServerConfig
}

type PublicServiceInterface interface {
	GetPartnerByDomain(domain string) (map[string]interface{}, error)

	GetCouponByCode(code string) (map[string]interface{}, error)
	ActivateCoupon(code string, req ActivateCouponRequest) (map[string]interface{}, error)
	PurchaseCoupon(req PurchaseCouponRequest) (map[string]interface{}, error)

	UploadImage(couponID string, file *multipart.FileHeader) (map[string]interface{}, error)
	EditImage(imageID string, req types.EditImageRequest) (map[string]interface{}, error)
	ProcessImage(imageID string, req types.ProcessImageRequest) (map[string]interface{}, error)
	GetImagePreview(imageID string) (map[string]interface{}, error)
	GetProcessingStatus(imageID string) (map[string]interface{}, error)
	GetImageForDownload(imageID string) (*image.Image, error)
	SendSchemaToEmail(imageID string, req SendEmailRequest) (map[string]interface{}, error)

	GetAvailableSizes() []map[string]interface{}
	GetAvailableStyles() []map[string]interface{}
}

type PublicServiceDepsInterface struct {
	CouponRepository  CouponRepositoryInterface
	ImageRepository   ImageRepositoryInterface
	PartnerRepository PartnerRepositoryInterface
	ImageService      ImageServiceInterface
	PaymentService    PaymentServiceInterface
	EmailService      EmailServiceInterface
	Config            *config.Config
}

type CouponValidatorInterface interface {
	ValidateCode(code string) error
	ValidateActivationRequest(req ActivateCouponRequest) error
	ValidatePurchaseRequest(req PurchaseCouponRequest) error
}

type ImageValidatorInterface interface {
	ValidateFile(file *multipart.FileHeader) error
	ValidateEditRequest(req types.EditImageRequest) error
	ValidateProcessRequest(req types.ProcessImageRequest) error
}

type UUIDGeneratorInterface interface {
	New() uuid.UUID
	Parse(s string) (uuid.UUID, error)
}

type CouponCodeGeneratorInterface interface {
	GenerateUniqueCouponCode(partnerCode string, repo CouponRepositoryInterface) (string, error)
}

type TimeProviderInterface interface {
	Now() int64
}

type LoggerInterface interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

type CacheInterface interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl int64) error
	Delete(key string) error
	Exists(key string) bool
}

type S3StorageInterface interface {
	UploadFile(key string, data []byte, contentType string) error
	GetPresignedURL(key string, ttl int64) (string, error)
	DeleteFile(key string) error
	FileExists(key string) (bool, error)
}

type QueueInterface interface {
	Enqueue(queueName string, payload interface{}) error
	Dequeue(queueName string) (interface{}, error)
	GetQueueSize(queueName string) (int, error)
}

type ImageProcessorInterface interface {
	ProcessWithStyle(imageData []byte, style string, params image.ProcessingParams) ([]byte, error)
	GeneratePreview(imageData []byte) ([]byte, error)
	ValidateImage(imageData []byte) error
}

type SchemaGeneratorInterface interface {
	GenerateFromImage(imageData []byte, params image.ProcessingParams) ([]byte, error)
	ValidateSchema(schemaData []byte) error
}

type PaymentGatewayInterface interface {
	CreatePayment(amount float64, currency string, description string) (*payment.AlfaBankRegisterResponse, error)
	GetPaymentStatus(paymentID string) (*payment.AlfaBankStatusResponse, error)
	RefundPayment(paymentID string, amount float64) error
}

type NotificationServiceInterface interface {
	SendSMS(phone, message string) error
	SendPushNotification(deviceToken, title, body string) error
	SendTelegramMessage(chatID int64, message string) error
}

type AnalyticsInterface interface {
	TrackEvent(event string, properties map[string]interface{}) error
	TrackUserAction(userID string, action string, properties map[string]interface{}) error
	IncrementCounter(metric string, value int) error
}

type HealthCheckerInterface interface {
	CheckDatabase() error
	CheckRedis() error
	CheckS3() error
	CheckExternalServices() error
	GetOverallHealth() map[string]interface{}
}

type MetricsCollectorInterface interface {
	IncrementCounter(name string, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
}
