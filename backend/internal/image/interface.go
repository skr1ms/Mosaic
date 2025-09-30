package image

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/internal/types"
	"github.com/skr1ms/mosaic/pkg/mosaic"
	"github.com/skr1ms/mosaic/pkg/stableDiffusion"
	"github.com/skr1ms/mosaic/pkg/zip"
)

type Coupon struct {
	ID          uuid.UUID  `json:"id"`
	Code        string     `json:"code"`
	Size        string     `json:"size"`
	Style       string     `json:"style"`
	Status      string     `json:"status"`
	UserEmail   *string    `json:"user_email"`
	CompletedAt *time.Time `json:"completed_at"`
	StonesCount *int       `json:"stones_count"`
}

type ImageRepositoryInterface interface {
	Create(ctx context.Context, task *Image) error
	GetByID(ctx context.Context, id uuid.UUID) (*Image, error)
	GetByCouponID(ctx context.Context, couponID uuid.UUID) (*Image, error)
	GetNextInQueue(ctx context.Context) (*Image, error)
	GetQueuedTasks(ctx context.Context) ([]*Image, error)
	GetProcessingTasks(ctx context.Context) ([]*Image, error)
	StartProcessing(ctx context.Context, id uuid.UUID) error
	CompleteProcessing(ctx context.Context, id uuid.UUID) error
	FailProcessing(ctx context.Context, id uuid.UUID, errorMessage string) error
	RetryTask(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, task *Image) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetAll(ctx context.Context) ([]*Image, error)
	GetByStatus(ctx context.Context, status string) ([]*Image, error)
	GetAllWithPartner(ctx context.Context) ([]*ImageWithPartner, error)
	GetByStatusWithPartner(ctx context.Context, status string) ([]*ImageWithPartner, error)
	GetWithFilters(ctx context.Context, status, dateFrom, dateTo string) ([]*ImageWithPartner, error)
	GetFailedTasksForRetry(ctx context.Context) ([]*Image, error)
	GetStatistics(ctx context.Context) (map[string]int64, error)
}

type CouponRepositoryInterface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*Coupon, error)
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	Update(ctx context.Context, coupon *Coupon) error
}

type S3ClientInterface interface {
	UploadFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, couponID uuid.UUID) (string, error)
	UploadFileWithKey(ctx context.Context, reader io.Reader, size int64, contentType string, objectKey string) (string, error)
	UploadPreviewFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, previewID uuid.UUID) (string, error)
	DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectKey string) error
	GetFileURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
	GetSignedURL(objectKey string, expires time.Duration) (string, error)
}

type StableDiffusionClientInterface interface {
	ProcessImage(ctx context.Context, req stableDiffusion.ProcessImageRequest) (string, error)
	EncodeImageToBase64(data []byte) string
	DecodeBase64Image(base64Data string) ([]byte, error)
	CheckHealth(ctx context.Context) error
}

type EmailServiceInterface interface {
	SendSchemaEmail(email, schemaURL, couponCode string) error
}

type ZipServiceInterface interface {
	CreateSchemaArchive(imageID uuid.UUID, files []zip.FileData) (*bytes.Buffer, error)
}

type MosaicGeneratorInterface interface {
	Generate(ctx context.Context, req *mosaic.GenerationRequest) (*mosaic.GenerationResult, error)
}

type ImageServiceInterface interface {
	GetQueue(status string) ([]*ImageWithPartner, error)
	GetQueueWithFilters(status, dateFrom, dateTo string) ([]*ImageWithPartner, error)
	UploadImage(ctx context.Context, couponID uuid.UUID, file *multipart.FileHeader, userEmail string) (*Image, error)
	EditImage(ctx context.Context, imageID uuid.UUID, editParams ImageEditParams) error
	ProcessImage(ctx context.Context, imageID uuid.UUID, processParams *ProcessingParams) error
	GenerateSchema(ctx context.Context, imageID uuid.UUID, confirmed bool) error
	GetImageStatus(ctx context.Context, imageID uuid.UUID) (*types.ImageStatusResponse, error)

	GetCouponRepository() CouponRepositoryInterface
	GetS3Client() S3ClientInterface
	GetImageRepository() ImageRepositoryInterface
}

type ImageHandlerInterface interface {
	UploadImage(c any) error
	EditImage(c any) error
	ProcessImage(c any) error
	GenerateSchema(c any) error
	GetImageStatus(c any) error

	GetQueue(c any) error
	GetTaskByID(c any) error
	AddToQueue(c any) error
	StartProcessing(c any) error
	CompleteProcessing(c any) error
	FailProcessing(c any) error
	RetryTask(c any) error
	DeleteTask(c any) error
	GetStatistics(c any) error
	GetNextTask(c any) error
}

type ImageValidatorInterface interface {
	ValidateImageFile(file *multipart.FileHeader) error
	ValidateEditParams(params ImageEditParams) error
	ValidateProcessingParams(params ProcessingParams) error
	ValidateImageFormat(data []byte) error
	ValidateImageSize(size int64) error
}

type QueueManagerInterface interface {
	AddToQueue(task *Image) error
	GetNextTask() (*Image, error)
	MarkTaskAsProcessing(id uuid.UUID) error
	CompleteTask(id uuid.UUID) error
	FailTask(id uuid.UUID, errorMessage string) error
	RetryTask(id uuid.UUID) error
	GetQueueStatistics() (map[string]int64, error)
}

type CacheInterface interface {
	Get(key string) (any, error)
	Set(key string, value any, ttl time.Duration) error
	Delete(key string) error
	Exists(key string) bool
}

type MetricsInterface interface {
	IncrementCounter(name string, labels map[string]string)
	RecordDuration(name string, duration time.Duration, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
}

type LoggerInterface interface {
	Info(msg string, fields ...any)
	Error(msg string, err error, fields ...any)
	Warn(msg string, fields ...any)
	Debug(msg string, fields ...any)
}

type ConfigInterface interface {
	GetS3MinioConfig() config.S3MinioConfig
	GetStableDiffusionConfig() config.StableDiffusionConfig
	GetMosaicGeneratorConfig() config.MosaicGeneratorConfig
}

type NotificationInterface interface {
	SendProcessingStarted(userEmail string, imageID uuid.UUID) error
	SendProcessingCompleted(userEmail string, imageID uuid.UUID, schemaURL string) error
	SendProcessingFailed(userEmail string, imageID uuid.UUID, errorMessage string) error
}

type HealthCheckerInterface interface {
	CheckImageRepository() error
	CheckS3Storage() error
	CheckStableDiffusion() error
	CheckEmailService() error
	GetOverallHealth() map[string]any
}

type BackupInterface interface {
	BackupImageData(imageID uuid.UUID) error
	RestoreImageData(imageID uuid.UUID) error
	CleanupOldBackups(olderThan time.Duration) error
}

type SecurityInterface interface {
	ValidateUserAccess(userEmail string, imageID uuid.UUID) error
	ScanImageForThreats(imageData []byte) error
	EncryptSensitiveData(data []byte) ([]byte, error)
	DecryptSensitiveData(data []byte) ([]byte, error)
}

type AnalyticsInterface interface {
	TrackImageUpload(couponID uuid.UUID, userEmail string) error
	TrackImageProcessing(imageID uuid.UUID, style string, useAI bool) error
	TrackSchemaGeneration(imageID uuid.UUID, successful bool) error
	GetProcessingStatistics(from, to time.Time) (map[string]any, error)
}

type RateLimiterInterface interface {
	CheckUserLimit(userEmail string, action string) error
	IncrementUserAction(userEmail string, action string) error
	GetUserLimits(userEmail string) (map[string]int, error)
	ResetUserLimits(userEmail string) error
}

type WebhookInterface interface {
	SendImageProcessedWebhook(imageID uuid.UUID, status string) error
	SendSchemaGeneratedWebhook(imageID uuid.UUID, schemaURL string) error
	RegisterWebhookHandler(event string, url string) error
}
