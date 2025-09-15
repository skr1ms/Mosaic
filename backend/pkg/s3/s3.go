package s3

import (
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type S3Client struct {
	client        *minio.Client
	imageBucket   string
	logosBucket   string
	chatBucket    string
	previewBucket string
	publicURL     string
	logger        *middleware.Logger
}

// NewS3Client creates new client for MinIO S3 operations
func NewS3Client(cfg config.S3MinioConfig, logger *middleware.Logger) (*S3Client, error) {
	// Always use internal endpoint for MinIO connection
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	s3Client := &S3Client{
		client:        minioClient,
		imageBucket:   cfg.BucketName,
		logosBucket:   cfg.LogosBucketName,
		chatBucket:    cfg.ChatBucketName,
		previewBucket: cfg.PreviewBucketName,
		publicURL:     cfg.PublicURL,
		logger:        logger,
	}

	// Create main bucket (images) if it doesn't exist
	if err := s3Client.ensureBucketExists(context.Background(), s3Client.imageBucket); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	// Create logos bucket if it doesn't exist
	if s3Client.logosBucket != "" {
		if err := s3Client.ensureBucketExists(context.Background(), s3Client.logosBucket); err != nil {
			return nil, fmt.Errorf("failed to ensure logos bucket exists: %w", err)
		}
	}

	// Create chat bucket if it doesn't exist
	if s3Client.chatBucket != "" {
		if err := s3Client.ensureBucketExists(context.Background(), s3Client.chatBucket); err != nil {
			return nil, fmt.Errorf("failed to ensure chat bucket exists: %w", err)
		}
	}

	// Create preview bucket if it doesn't exist
	if s3Client.previewBucket != "" {
		if err := s3Client.ensureBucketExists(context.Background(), s3Client.previewBucket); err != nil {
			return nil, fmt.Errorf("failed to ensure preview bucket exists: %w", err)
		}
	}

	return s3Client, nil
}

// ensureBucketExists checks bucket existence and creates it if necessary
func (s *S3Client) ensureBucketExists(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		s.logger.GetZerologLogger().Info().Str("bucket", bucket).Msg("Created S3 bucket")
	}

	return nil
}

// UploadFile uploads file to S3 and returns file path
func (s *S3Client) UploadFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, couponID uuid.UUID) (string, error) {
	// Generate unique filename
	fileName := fmt.Sprintf("%s_%d", couponID.String(), time.Now().Unix())

	// Determine extension by content type
	ext := getExtensionFromContentType(contentType)
	if ext != "" {
		fileName += ext
	}

	// Form full path
	objectKey := filepath.Join(folder, fileName)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// Upload file
	_, err := s.client.PutObject(ctx, s.imageBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", s.imageBucket).
		Str("key", objectKey).
		Str("coupon_id", couponID.String()).
		Msg("File uploaded to S3")

	return objectKey, nil
}

// UploadFileWithKey uploads file to S3 by given objectKey and returns key
func (s *S3Client) UploadFileWithKey(ctx context.Context, reader io.Reader, size int64, contentType string, objectKey string) (string, error) {
	// Normalize path
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	_, err := s.client.PutObject(ctx, s.imageBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3 with key: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", s.imageBucket).
		Str("key", objectKey).
		Msg("File uploaded to S3 with explicit key")

	return objectKey, nil
}

// DownloadFile downloads file from S3
func (s *S3Client) DownloadFile(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.imageBucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}

	return object, nil
}

// DeleteFile deletes file from S3
func (s *S3Client) DeleteFile(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.imageBucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", s.imageBucket).
		Str("key", objectKey).
		Msg("File deleted from S3")

	return nil
}

// UploadPreviewFile uploads preview file to preview bucket and returns file path
func (s *S3Client) UploadPreviewFile(ctx context.Context, reader io.Reader, size int64, contentType, folder string, previewID uuid.UUID) (string, error) {
	// Generate unique filename for preview
	fileName := fmt.Sprintf("preview_%s_%d", previewID.String(), time.Now().Unix())

	// Determine extension by content type
	ext := getExtensionFromContentType(contentType)
	if ext != "" {
		fileName += ext
	}

	// Form full path
	objectKey := filepath.Join(folder, fileName)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	// Upload file to preview bucket
	_, err := s.client.PutObject(ctx, s.previewBucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload preview file to S3: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", s.previewBucket).
		Str("key", objectKey).
		Str("preview_id", previewID.String()).
		Msg("Preview file uploaded to S3")

	return objectKey, nil
}

// GetFileURL returns signed URL for file access
func (s *S3Client) GetFileURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	// If public URL is configured, return direct URL (assuming bucket is public)
	if s.publicURL != "" {
		// Form direct URL to file
		directURL := fmt.Sprintf("%s/%s/%s", s.publicURL, s.imageBucket, objectKey)
		return directURL, nil
	}

	// Otherwise, generate presigned URL for MinIO
	url, err := s.client.PresignedGetObject(ctx, s.imageBucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetSignedURL returns signed URL for file access (alias for GetFileURL for compatibility)
func (s *S3Client) GetSignedURL(objectKey string, expires time.Duration) (string, error) {
	ctx := context.Background()
	return s.GetFileURL(ctx, objectKey, expires)
}

// ListFiles returns list of files by prefix
func (s *S3Client) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string

	objectCh := s.client.ListObjects(ctx, s.imageBucket, minio.ListObjectsOptions{
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

// CopyFile copies file within S3
func (s *S3Client) CopyFile(ctx context.Context, srcKey, destKey string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: s.imageBucket,
		Object: srcKey,
	}

	destOpts := minio.CopyDestOptions{
		Bucket: s.imageBucket,
		Object: destKey,
	}

	_, err := s.client.CopyObject(ctx, destOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("failed to copy file in S3: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", s.imageBucket).
		Str("src_key", srcKey).
		Str("dest_key", destKey).
		Msg("File copied in S3")

	return nil
}

// getExtensionFromContentType returns file extension by MIME type
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
	case "application/zip":
		return ".zip"
	case "video/mp4":
		return ".mp4"
	case "video/mpeg":
		return ".mpeg"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/webm":
		return ".webm"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	default:
		return ""
	}
}

// GetBucketName returns bucket name
func (s *S3Client) GetBucketName() string {
	return s.imageBucket
}

// GetPreviewBucketName returns preview bucket name
func (s *S3Client) GetPreviewBucketName() string {
	return s.previewBucket
}

// DownloadFromPreviewBucket downloads file from preview bucket
func (s *S3Client) DownloadFromPreviewBucket(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if s.previewBucket == "" {
		return nil, fmt.Errorf("preview bucket not configured")
	}

	object, err := s.client.GetObject(ctx, s.previewBucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download from preview bucket: %w", err)
	}

	return object, nil
}

// UploadLogo uploads partner logo to separate bucket and returns key
func (s *S3Client) UploadLogo(ctx context.Context, reader io.Reader, size int64, contentType string, partnerID string) (string, error) {
	bucket := s.logosBucket
	if bucket == "" {
		return "", fmt.Errorf("logos bucket is not configured")
	}

	fileName := fmt.Sprintf("%s_%d", partnerID, time.Now().Unix())
	ext := getExtensionFromContentType(contentType)
	if ext != "" {
		fileName += ext
	}
	objectKey := filepath.Join(partnerID, fileName)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	_, err := s.client.PutObject(ctx, bucket, objectKey, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("failed to upload logo to S3: %w", err)
	}
	s.logger.GetZerologLogger().Info().Str("bucket", bucket).Str("key", objectKey).Str("partner_id", partnerID).Msg("Logo uploaded to S3")
	return objectKey, nil
}

// GetLogoURL returns URL for logo from logos bucket
func (s *S3Client) GetLogoURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	bucket := s.logosBucket
	if bucket == "" {
		return "", fmt.Errorf("logos bucket is not configured")
	}

	// If public URL is configured, return direct URL for public bucket
	if s.publicURL != "" {
		// Form direct URL for public bucket
		directURL := fmt.Sprintf("%s/%s/%s", s.publicURL, bucket, objectKey)
		return directURL, nil
	}

	// Otherwise, generate presigned URL
	url, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned logo URL: %w", err)
	}

	finalURL := url.String()
	if s.publicURL != "" {
		// replace minio host with public URL
		finalURL = strings.Replace(finalURL, "http://minio:9000", s.publicURL, 1)
		finalURL = strings.Replace(finalURL, "https://minio:9000", s.publicURL, 1)
	}

	// Add timestamp to avoid caching
	timestamp := time.Now().Unix()
	if strings.Contains(finalURL, "?") {
		finalURL += fmt.Sprintf("&v=%d", timestamp)
	} else {
		finalURL += fmt.Sprintf("?v=%d", timestamp)
	}

	return finalURL, nil
}

// GetChatFileURL returns public URL for chat attachment
func (s *S3Client) GetChatFileURL(objectKey string) string {
	if s.publicURL != "" {
		// Form direct URL for public chat bucket
		return fmt.Sprintf("%s/%s/%s", s.publicURL, s.chatBucket, objectKey)
	}
	// Fallback to internal endpoint (not recommended for public access)
	return fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.chatBucket, objectKey)
}

// UploadChatData uploads chat attachment file to chat bucket and returns key and presigned URL
func (s *S3Client) UploadChatData(ctx context.Context, reader io.Reader, size int64, contentType string, senderID string, targetID string, originalFilename string) (string, string, error) {
	bucket := s.chatBucket
	if bucket == "" {
		return "", "", fmt.Errorf("chat bucket is not configured")
	}
	// File name: chat/senderID/targetID/timestamp_original
	safeName := strings.ReplaceAll(originalFilename, " ", "_")
	fileName := fmt.Sprintf("%d_%s", time.Now().Unix(), safeName)
	objectKey := filepath.Join("chat", senderID, targetID, fileName)
	objectKey = strings.ReplaceAll(objectKey, "\\", "/")

	_, err := s.client.PutObject(ctx, bucket, objectKey, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload chat data to S3: %w", err)
	}

	// Use public URL directly since bucket is now public
	finalURL := s.GetChatFileURL(objectKey)

	s.logger.GetZerologLogger().Info().Str("bucket", bucket).Str("key", objectKey).Str("sender_id", senderID).Str("target_id", targetID).Str("url", finalURL).Msg("Chat data uploaded to S3")
	return objectKey, finalURL, nil
}

// DownloadChatData returns chat attachment object by key
func (s *S3Client) DownloadChatData(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	bucket := s.chatBucket
	if bucket == "" {
		return nil, "", fmt.Errorf("chat bucket is not configured")
	}
	obj, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get chat object: %w", err)
	}
	// Extract content-type from Stat if available
	var ct string
	if info, statErr := obj.Stat(); statErr == nil {
		ct = info.ContentType
	}
	return obj, ct, nil
}

// DeleteChatData deletes chat attachment object by key
func (s *S3Client) DeleteChatData(ctx context.Context, objectKey string) error {
	bucket := s.chatBucket
	if bucket == "" {
		return fmt.Errorf("chat bucket is not configured")
	}
	if objectKey == "" {
		return nil
	}
	// Normalize and validate key to ensure deletion of a single object
	key := strings.TrimSpace(objectKey)
	key = strings.TrimPrefix(key, "http://")
	key = strings.TrimPrefix(key, "https://")
	// If a full URL was passed — extract path after bucket
	if idx := strings.Index(key, "/"); idx >= 0 && strings.Contains(key, "://") {
		key = key[idx+1:]
	}
	key = strings.TrimPrefix(key, "/")
	// Require expected prefix and filename
	if !strings.HasPrefix(key, "chat/") {
		return fmt.Errorf("invalid chat object key")
	}
	parts := strings.Split(key, "/")
	if len(parts) < 4 {
		return fmt.Errorf("invalid chat object key depth")
	}
	filename := parts[len(parts)-1]
	if filename == "" || strings.HasSuffix(filename, "/") {
		return fmt.Errorf("invalid chat object filename")
	}
	// Log and delete only the specific object
	s.logger.GetZerologLogger().Info().Str("bucket", bucket).Str("key", key).Msg("Deleting chat object from S3")
	return s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

// UploadToPreviewBucket uploads file to preview-images bucket
func (s *S3Client) UploadToPreviewBucket(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	bucket := s.previewBucket
	if bucket == "" {
		return fmt.Errorf("preview bucket is not configured")
	}

	// Check if object already exists to avoid duplicates
	_, err := s.client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err == nil {
		// Object already exists, skip upload
		s.logger.GetZerologLogger().Info().
			Str("bucket", bucket).
			Str("key", objectKey).
			Msg("Preview file already exists, skipping upload")
		return nil
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", bucket).
		Str("key", objectKey).
		Int64("size", size).
		Str("content_type", contentType).
		Msg("Uploading preview to S3")

	_, err = s.client.PutObject(ctx, bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload preview to S3: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", bucket).
		Str("key", objectKey).
		Msg("Preview uploaded successfully to S3")

	return nil
}

// GetPreviewURL returns public URL for preview file
func (s *S3Client) GetPreviewURL(objectKey string) string {
	bucket := s.previewBucket
	if bucket == "" {
		s.logger.GetZerologLogger().Error().
			Msg("Preview bucket is not configured")
		return ""
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", bucket).
		Str("key", objectKey).
		Str("public_url", s.publicURL).
		Msg("Generating preview URL")

	// Always use nginx proxy path for previews
	// This works with the nginx location /preview-images/ proxy
	proxyURL := fmt.Sprintf("/preview-images/%s", objectKey)
	s.logger.GetZerologLogger().Info().
		Str("url", proxyURL).
		Msg("Generated proxy URL for preview")
	return proxyURL
}

// DeleteFromPreviewBucket deletes file from preview-images bucket
func (s *S3Client) DeleteFromPreviewBucket(ctx context.Context, objectKey string) error {
	bucket := s.previewBucket
	if bucket == "" {
		return fmt.Errorf("preview bucket is not configured")
	}

	err := s.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from preview bucket: %w", err)
	}

	s.logger.GetZerologLogger().Info().
		Str("bucket", bucket).
		Str("key", objectKey).
		Msg("Preview file deleted from S3")

	return nil
}

// SchedulePreviewDeletion schedules automatic deletion of preview file after 30 minutes
func (s *S3Client) SchedulePreviewDeletion(objectKey string) {
	go func() {
		// Wait for 30 minutes
		time.Sleep(30 * time.Minute)

		// Delete the preview file
		ctx := context.Background()
		if err := s.DeleteFromPreviewBucket(ctx, objectKey); err != nil {
			s.logger.GetZerologLogger().Error().
				Err(err).
				Str("object_key", objectKey).
				Msg("Failed to auto-delete preview file")
		} else {
			s.logger.GetZerologLogger().Info().
				Str("object_key", objectKey).
				Msg("Preview file auto-deleted after 30 minutes")
		}
	}()
}

// CleanupAllPreviews mass deletes all old preview files (EMERGENCY CLEANUP)
func (s *S3Client) CleanupAllPreviews(ctx context.Context) error {
	bucket := s.previewBucket
	if bucket == "" {
		return fmt.Errorf("preview bucket is not configured")
	}

	s.logger.GetZerologLogger().Warn().
		Str("bucket", bucket).
		Msg("EMERGENCY: Starting mass preview cleanup")

	// List all objects in preview bucket
	objectsCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Recursive: true,
	})

	deletedCount := 0
	errorCount := 0

	for object := range objectsCh {
		if object.Err != nil {
			s.logger.GetZerologLogger().Error().
				Err(object.Err).
				Msg("Error listing preview objects")
			continue
		}

		// Delete old previews (older than 1 hour)
		if time.Since(object.LastModified) > 1*time.Hour {
			err := s.client.RemoveObject(ctx, bucket, object.Key, minio.RemoveObjectOptions{})
			if err != nil {
				s.logger.GetZerologLogger().Error().
					Err(err).
					Str("key", object.Key).
					Msg("Failed to delete old preview")
				errorCount++
			} else {
				deletedCount++
			}
		}
	}

	s.logger.GetZerologLogger().Info().
		Int("deleted_count", deletedCount).
		Int("error_count", errorCount).
		Msg("Preview cleanup completed")

	return nil
}

// StartPreviewCleanupJob starts background job to clean up expired preview files
func (s *S3Client) StartPreviewCleanupJob(ctx context.Context) error {
	if s.previewBucket == "" {
		s.logger.GetZerologLogger().Warn().Msg("Preview bucket not configured, skipping cleanup job")
		return nil
	}

	s.logger.GetZerologLogger().Info().Msg("Starting preview cleanup job")

	go func() {
		ticker := time.NewTicker(10 * time.Minute) // Check every 10 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.GetZerologLogger().Info().Msg("Preview cleanup job stopped")
				return
			case <-ticker.C:
				if err := s.cleanupExpiredPreviews(ctx); err != nil {
					s.logger.GetZerologLogger().Error().
						Err(err).
						Msg("Failed to cleanup expired previews")
				}
			}
		}
	}()

	return nil
}

// cleanupExpiredPreviews removes preview files older than 30 minutes
func (s *S3Client) cleanupExpiredPreviews(ctx context.Context) error {
	if s.previewBucket == "" {
		return nil
	}

	// List all objects in preview bucket
	objectCh := s.client.ListObjects(ctx, s.previewBucket, minio.ListObjectsOptions{
		Prefix:    "previews/",
		Recursive: true,
	})

	expiredFiles := 0
	totalFiles := 0

	for object := range objectCh {
		if object.Err != nil {
			return fmt.Errorf("error listing preview objects: %w", object.Err)
		}

		totalFiles++

		// Check if file is older than 30 minutes
		if time.Since(object.LastModified) > 30*time.Minute {
			if err := s.client.RemoveObject(ctx, s.previewBucket, object.Key, minio.RemoveObjectOptions{}); err != nil {
				s.logger.GetZerologLogger().Error().
					Err(err).
					Str("bucket", s.previewBucket).
					Str("key", object.Key).
					Msg("Failed to remove expired preview")
			} else {
				expiredFiles++
				s.logger.GetZerologLogger().Debug().
					Str("bucket", s.previewBucket).
					Str("key", object.Key).
					Time("last_modified", object.LastModified).
					Msg("Expired preview file removed")
			}
		}
	}

	if expiredFiles > 0 {
		s.logger.GetZerologLogger().Info().
			Int("expired_files", expiredFiles).
			Int("total_files", totalFiles).
			Msg("Preview cleanup completed")
	}

	return nil
}

// Decode decodes an image from a reader and returns the image and format
func (c *S3Client) Decode(reader io.Reader) (image.Image, string, error) {
	// Try to decode as JPEG first
	readerBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Create a new reader from the bytes for each decode attempt
	jpegReader := io.NopCloser(strings.NewReader(string(readerBytes)))
	img, err := jpeg.Decode(jpegReader)
	if err == nil {
		return img, "jpeg", nil
	}

	// Try PNG
	pngReader := io.NopCloser(strings.NewReader(string(readerBytes)))
	img, err = png.Decode(pngReader)
	if err == nil {
		return img, "png", nil
	}

	// Try GIF
	gifReader := io.NopCloser(strings.NewReader(string(readerBytes)))
	img, err = gif.Decode(gifReader)
	if err == nil {
		return img, "gif", nil
	}

	// Try generic image decode as fallback
	genericReader := io.NopCloser(strings.NewReader(string(readerBytes)))
	img, format, err := image.Decode(genericReader)
	if err != nil {
		return nil, "", fmt.Errorf("unsupported image format or corrupted image: %w", err)
	}

	return img, format, nil
}
