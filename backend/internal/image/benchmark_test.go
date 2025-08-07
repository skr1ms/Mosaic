package image

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/skr1ms/mosaic/pkg/zip"
	"github.com/stretchr/testify/mock"
)

func BenchmarkRepository_GetByStatus_Queued(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImages := make([]*Image, 100)
	for i := 0; i < 100; i++ {
		testImages[i] = createTestImage()
	}

	mockRepo.On("GetByStatus", ctx, "queued").Return(testImages, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByStatus(ctx, "queued")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_GetAll(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImages := make([]*Image, 1000)
	for i := 0; i < 1000; i++ {
		testImages[i] = createTestImage()
	}

	mockRepo.On("GetAll", ctx).Return(testImages, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetAll(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkS3Client_UploadFile(b *testing.B) {
	mockS3 := &MockS3Client{}
	ctx := context.Background()
	couponID := uuid.New()

	mockS3.On("UploadFile", ctx, mock.Anything, int64(1024), "image/jpeg", "originals", couponID).Return("originals/test.jpg", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockS3.UploadFile(ctx, bytes.NewReader([]byte("test")), 1024, "image/jpeg", "originals", couponID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkS3Client_DownloadFile(b *testing.B) {
	mockS3 := &MockS3Client{}
	ctx := context.Background()

	imageData := bytes.NewReader([]byte("fake-image-data"))
	mockS3.On("DownloadFile", ctx, "test.jpg").Return(imageData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockS3.DownloadFile(ctx, "test.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStableDiffusion_EncodeImageToBase64(b *testing.B) {
	mockSD := &MockStableDiffusionClient{}

	imageData := make([]byte, 1024*1024) // 1MB image
	mockSD.On("EncodeImageToBase64", imageData).Return("base64-encoded-data")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = mockSD.EncodeImageToBase64(imageData)
	}
}

func BenchmarkStableDiffusion_DecodeBase64Image(b *testing.B) {
	mockSD := &MockStableDiffusionClient{}

	base64Data := "base64-encoded-data"
	decodedData := make([]byte, 1024*1024) // 1MB image
	mockSD.On("DecodeBase64Image", base64Data).Return(decodedData, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockSD.DecodeBase64Image(base64Data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkZipService_CreateSchemaArchive(b *testing.B) {
	mockZip := &MockZipService{}
	imageID := uuid.New()

	files := []zip.FileData{
		{
			Name:    "original.jpg",
			Content: bytes.NewReader(make([]byte, 1024*1024)), // 1MB
			Size:    1024 * 1024,
		},
		{
			Name:    "processed.jpg",
			Content: bytes.NewReader(make([]byte, 1024*1024)), // 1MB
			Size:    1024 * 1024,
		},
	}

	zipBuffer := bytes.NewBuffer(make([]byte, 2*1024*1024)) // 2MB zip
	mockZip.On("CreateSchemaArchive", imageID, files).Return(zipBuffer, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockZip.CreateSchemaArchive(imageID, files)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_Create(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImage := createTestImage()
	mockRepo.On("Create", ctx, mock.AnythingOfType("*image.Image")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockRepo.Create(ctx, testImage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_GetByID(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImage := createTestImage()
	mockRepo.On("GetByID", ctx, testImage.ID).Return(testImage, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetByID(ctx, testImage.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_Update(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImage := createTestImage()
	mockRepo.On("Update", ctx, mock.AnythingOfType("*image.Image")).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockRepo.Update(ctx, testImage)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_GetStatistics(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	stats := map[string]int64{
		"queued":     100,
		"processing": 50,
		"completed":  1000,
		"failed":     10,
	}
	mockRepo.On("GetStatistics", ctx).Return(stats, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetStatistics(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_StartProcessing(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()
	imageID := uuid.New()

	mockRepo.On("StartProcessing", ctx, imageID).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockRepo.StartProcessing(ctx, imageID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_CompleteProcessing(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()
	imageID := uuid.New()

	mockRepo.On("CompleteProcessing", ctx, imageID).Return(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := mockRepo.CompleteProcessing(ctx, imageID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepository_GetNextInQueue(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImage := createTestImage()
	mockRepo.On("GetNextInQueue", ctx).Return(testImage, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mockRepo.GetNextInQueue(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConcurrentRepository_GetByStatus(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	testImages := make([]*Image, 50)
	for i := 0; i < 50; i++ {
		testImages[i] = createTestImage()
	}

	mockRepo.On("GetByStatus", ctx, "queued").Return(testImages, nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mockRepo.GetByStatus(ctx, "queued")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkConcurrentS3_GetFileURL(b *testing.B) {
	mockS3 := &MockS3Client{}
	ctx := context.Background()

	mockS3.On("GetFileURL", ctx, "test.jpg", 24*time.Hour).Return("https://example.com/test.jpg", nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := mockS3.GetFileURL(ctx, "test.jpg", 24*time.Hour)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkMemoryAllocation_LargeImageList(b *testing.B) {
	mockRepo := &MockImageRepository{}
	ctx := context.Background()

	// Create a large dataset
	testImages := make([]*Image, 10000)
	for i := 0; i < 10000; i++ {
		testImages[i] = createTestImage()
	}

	mockRepo.On("GetByStatus", ctx, "queued").Return(testImages, nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := mockRepo.GetByStatus(ctx, "queued")
		if err != nil {
			b.Fatal(err)
		}
		_ = len(result)
	}
}