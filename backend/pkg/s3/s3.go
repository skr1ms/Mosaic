package s3

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
)

type S3Client struct {
	client     *minio.Client
	bucketName string
}

// NewS3Client создает новый клиент для работы с MinIO S3
func NewS3Client(cfg config.S3MinioConfig) (*S3Client, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	s3Client := &S3Client{
		client:     minioClient,
		bucketName: cfg.BucketName,
	}

	// Создаем bucket если он не существует
	if err := s3Client.ensureBucketExists(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return s3Client, nil
}

// ensureBucketExists проверяет существование bucket и создает его при необходимости
func (s *S3Client) ensureBucketExists(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", s.bucketName).Msg("Created S3 bucket")
	}

	return nil
}

// UploadFile загружает файл в S3 и возвращает путь к файлу
func (s *S3Client) UploadFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, couponID uuid.UUID) (string, error) {
	// Генерируем уникальное имя файла
	fileName := fmt.Sprintf("%s_%d", couponID.String(), time.Now().Unix())

	// Определяем расширение по типу контента
	ext := getExtensionFromContentType(contentType)
	if ext != "" {
		fileName += ext
	}

	// Формируем полный путь
	objectKey := filepath.Join(folder, fileName)
	// Заменяем обратные слеши на прямые для S3
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// Загружаем файл
	_, err := s.client.PutObject(ctx, s.bucketName, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	log.Info().
		Str("bucket", s.bucketName).
		Str("key", objectKey).
		Str("coupon_id", couponID.String()).
		Msg("File uploaded to S3")

	return objectKey, nil
}

// DownloadFile скачивает файл из S3
func (s *S3Client) DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return object, nil
}

// DeleteFile удаляет файл из S3
func (s *S3Client) DeleteFile(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	log.Info().
		Str("bucket", s.bucketName).
		Str("key", objectKey).
		Msg("File deleted from S3")

	return nil
}

// GetFileURL возвращает подписанный URL для доступа к файлу
func (s *S3Client) GetFileURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// ListFiles возвращает список файлов по префиксу
func (s *S3Client) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("error listing objects: %w", object.Err)
		}
		files = append(files, object.Key)
	}

	return files, nil
}

// CopyFile копирует файл внутри S3
func (s *S3Client) CopyFile(ctx context.Context, srcKey, destKey string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: s.bucketName,
		Object: srcKey,
	}

	destOpts := minio.CopyDestOptions{
		Bucket: s.bucketName,
		Object: destKey,
	}

	_, err := s.client.CopyObject(ctx, destOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy file in S3: %w", err)
	}

	log.Info().
		Str("bucket", s.bucketName).
		Str("src_key", srcKey).
		Str("dest_key", destKey).
		Msg("File copied in S3")

	return nil
}

// getExtensionFromContentType возвращает расширение файла по MIME типу
func getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "application/zip":
		return ".zip"
	default:
		return ""
	}
}

// GetBucketName возвращает имя bucket
func (s *S3Client) GetBucketName() string {
	return s.bucketName
}
