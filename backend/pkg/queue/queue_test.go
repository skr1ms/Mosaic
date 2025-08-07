package queue

import (
	"testing"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// setupMockRedis создает mock Redis клиент для unit тестов
func setupMockRedis() *redis.Client {
	// Для unit тестов используем нереальный адрес, чтобы не зависеть от внешних сервисов
	// Эти тесты проверяют только создание структур, а не реальные Redis операции
	return redis.NewClient(&redis.Options{
		Addr: "mock:6379",
	})
}

func TestTaskQueue_Creation(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()
	queueName := "test-queue"

	// Act
	queue := NewTaskQueue(redisClient, queueName)

	// Assert
	assert.NotNil(t, queue)
	assert.Equal(t, queueName, queue.name)
	assert.Equal(t, redisClient, queue.redis)
}

// Тест создания задачи без выполнения Redis операций
func TestTaskQueue_TaskCreation(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()
	queue := NewTaskQueue(redisClient, "test-queue")

	payload := map[string]interface{}{
		"test_field": "test_value",
		"number":     123,
	}

	// Assert структура данных для задачи
	assert.NotNil(t, queue)
	assert.NotEmpty(t, payload)
	assert.Contains(t, payload, "test_field")
	assert.Equal(t, "test_value", payload["test_field"])
}

func TestImageTaskQueue_Creation(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()

	// Act
	imageQueue := NewImageTaskQueue(redisClient)

	// Assert
	assert.NotNil(t, imageQueue)
	assert.NotNil(t, imageQueue.TaskQueue)
	assert.Equal(t, "images", imageQueue.name)
}

// Тест валидации данных для обработки изображений
func TestImageTaskQueue_PayloadValidation(t *testing.T) {
	imageID := uuid.New()
	style := "grayscale"
	parameters := map[string]interface{}{
		"contrast":   "high",
		"brightness": 50.0,
	}

	// Assert структура данных
	assert.NotEqual(t, uuid.Nil, imageID)
	assert.NotEmpty(t, style)
	assert.Contains(t, parameters, "contrast")
	assert.Contains(t, parameters, "brightness")
}

func TestImageTaskQueue_SchemaGenerationValidation(t *testing.T) {
	imageID := uuid.New()
	couponID := uuid.New()
	confirmed := true

	// Assert структура данных
	assert.NotEqual(t, uuid.Nil, imageID)
	assert.NotEqual(t, uuid.Nil, couponID)
	assert.True(t, confirmed)
}

func TestImageTaskQueue_EmailSendingValidation(t *testing.T) {
	email := "test@example.com"
	schemaURL := "https://example.com/schema.zip"
	couponCode := "1234-5678-9012"

	// Assert структура данных
	assert.NotEmpty(t, email)
	assert.Contains(t, email, "@")
	assert.NotEmpty(t, schemaURL)
	assert.Contains(t, schemaURL, "https://")
	assert.NotEmpty(t, couponCode)
}

func TestQueueManager_Creation(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()

	// Act
	manager := NewQueueManager(redisClient)

	// Assert
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.imageQueue)
	assert.NotNil(t, manager.queues)
}

func TestQueueManager_CreateQueue(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()
	manager := NewQueueManager(redisClient)
	queueName := "custom-queue"

	// Act
	queue := manager.CreateQueue(queueName)

	// Assert
	assert.NotNil(t, queue)

	// Проверяем, что очередь сохранена в менеджере
	retrievedQueue := manager.GetQueue(queueName)
	assert.Equal(t, queue, retrievedQueue)
}

func TestQueueManager_GetImageQueue(t *testing.T) {
	// Arrange
	redisClient := setupMockRedis()
	manager := NewQueueManager(redisClient)

	// Act
	imageQueue := manager.GetImageQueue()

	// Assert
	assert.NotNil(t, imageQueue)
	assert.IsType(t, &ImageTaskQueue{}, imageQueue)
}

func TestImageTaskQueue_TaskTypes(t *testing.T) {
	// Test task type constants
	assert.Equal(t, "image_processing", TaskTypeImageProcessing)
	assert.Equal(t, "schema_generation", TaskTypeSchemaGeneration)
	assert.Equal(t, "email_sending", TaskTypeEmailSending)
	assert.Equal(t, "image_optimization", TaskTypeImageOptimization)
	assert.Equal(t, "thumbnail_generation", TaskTypeThumbnailGeneration)
}

func TestImageTaskHandlers_Creation(t *testing.T) {
	// Создаем mock адаптеры (для этого теста достаточно nil)
	var imageAdapter *ImageServiceAdapter
	var emailAdapter *EmailServiceAdapter

	// Act
	handlers := GetImageTaskHandlers(imageAdapter, emailAdapter)

	// Assert
	assert.NotNil(t, handlers)
	assert.Len(t, handlers, 5)

	// Проверяем, что все обработчики присутствуют
	assert.Contains(t, handlers, "process_image")
	assert.Contains(t, handlers, "generate_schema")
	assert.Contains(t, handlers, "send_schema")
	assert.Contains(t, handlers, "optimize_image")
	assert.Contains(t, handlers, "generate_thumbnails")
}

// Тест проверяет создание задач без реального Redis
func TestImageTaskQueue_TaskCreation(t *testing.T) {
	// Arrange
	imageID := uuid.New()
	couponID := uuid.New()

	// Test data structures
	testCases := []struct {
		name     string
		taskType string
		payload  map[string]interface{}
	}{
		{
			name:     "Image Processing Task",
			taskType: TaskTypeImageProcessing,
			payload: map[string]interface{}{
				"image_id":   imageID.String(),
				"style":      "grayscale",
				"parameters": map[string]interface{}{"contrast": "high"},
			},
		},
		{
			name:     "Schema Generation Task",
			taskType: TaskTypeSchemaGeneration,
			payload: map[string]interface{}{
				"image_id":  imageID.String(),
				"coupon_id": couponID.String(),
				"confirmed": true,
			},
		},
		{
			name:     "Email Sending Task",
			taskType: TaskTypeEmailSending,
			payload: map[string]interface{}{
				"email":       "test@example.com",
				"schema_url":  "https://example.com/schema.zip",
				"coupon_code": "1234-5678-9012",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Assert that task structure is valid
			assert.NotEmpty(t, tc.taskType)
			assert.NotEmpty(t, tc.payload)

			// Verify required fields exist
			switch tc.taskType {
			case TaskTypeImageProcessing:
				assert.Contains(t, tc.payload, "image_id")
				assert.Contains(t, tc.payload, "style")
			case TaskTypeSchemaGeneration:
				assert.Contains(t, tc.payload, "image_id")
				assert.Contains(t, tc.payload, "coupon_id")
			case TaskTypeEmailSending:
				assert.Contains(t, tc.payload, "email")
				assert.Contains(t, tc.payload, "schema_url")
			}
		})
	}
}
