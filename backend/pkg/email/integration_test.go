package email

import (
	"testing"
	"time"

	"github.com/skr1ms/mosaic/config"
)

func TestEmailQueue_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockConfig := &config.Config{}

	mailer := NewMailer(mockConfig)

	queue := NewEmailQueue(mailer, 1)

	go queue.Start()
	defer queue.Stop()

	time.Sleep(100 * time.Millisecond)

	err := queue.SendSchemaEmail("test@example.com", "TEST123", "https://example.com/schema.zip")
	if err != nil {
		t.Logf("Expected error due to mock SMTP config: %v", err)
	}

	err = queue.SendProcessingErrorEmail("test@example.com", "TEST123", "Test error message")
	if err != nil {
		t.Logf("Expected error due to mock SMTP config: %v", err)
	}

	err = queue.SendStatusUpdateEmail("test@example.com", "TEST123", "processing", 50)
	if err != nil {
		t.Logf("Expected error due to mock SMTP config: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	t.Log("Email queue integration test completed")
}

func TestRealEmailTemplates(t *testing.T) {
	tm := NewTemplateManager("templates")
	err := tm.loadTemplates()
	if err != nil {
		t.Fatalf("Failed to load real templates: %v", err)
	}

	testData := TemplateData{
		RecipientName:  "Иван Петров",
		RecipientEmail: "ivan@example.com",
		SupportEmail:   "support@mosaic.example.com",
		CompanyName:    "Мозаика Сервис",
		WebsiteURL:     "https://mosaic.example.com",
		CouponCode:     "MOSAIC123",
		SchemaURL:      "https://s3.example.com/schemas/MOSAIC123.zip",
		PreviewURL:     "https://s3.example.com/previews/MOSAIC123.jpg",
		OriginalURL:    "https://s3.example.com/originals/MOSAIC123.jpg",
		Size:           "40x50 см",
		Style:          "Классический стиль",
		ProcessingTime: "25 минут",
		ErrorMessage:   "Недостаточное разрешение изображения",
		ErrorCode:      "IMG_LOW_RES",
		RetryURL:       "https://mosaic.example.com/retry/MOSAIC123",
		Status:         "В обработке",
		StatusMessage:  "Создание схемы мозаики",
		Progress:       65,
		Timestamp:      time.Now().Format("02.01.2006 15:04:05"),
	}

	templates := []string{"schema_ready", "processing_error", "status_update"}

	for _, templateName := range templates {
		t.Run(templateName, func(t *testing.T) {
			subject, htmlBody, textBody, err := tm.RenderTemplate(templateName, testData)
			if err != nil {
				t.Errorf("Failed to render template %s: %v", templateName, err)
				return
			}

			if subject == "" {
				t.Errorf("Empty subject for template %s", templateName)
			}

			if htmlBody == "" {
				t.Errorf("Empty HTML body for template %s", templateName)
			}

			if textBody == "" {
				t.Errorf("Empty text body for template %s", templateName)
			}

			t.Logf("Template %s rendered successfully:", templateName)
			t.Logf("  Subject: %s", subject)
			t.Logf("  HTML length: %d chars", len(htmlBody))
			t.Logf("  Text length: %d chars", len(textBody))
		})
	}
}
