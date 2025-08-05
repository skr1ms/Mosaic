package email

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTemplateManager_LoadTemplates(t *testing.T) {
	tempDir := t.TempDir()

	testTemplate := `<h1>Hello {{.RecipientName}}!</h1>
<p>Your coupon: {{.CouponCode}}</p>`

	err := os.WriteFile(filepath.Join(tempDir, "test_template.html"), []byte(testTemplate), 0644)
	if err != nil {
		t.Fatalf("Failed to create test template: %v", err)
	}

	tm := NewTemplateManager(tempDir)
	err = tm.loadTemplates()
	if err != nil {
		t.Fatalf("Failed to load templates: %v", err)
	}

	if len(tm.templates) == 0 {
		t.Error("No templates loaded")
	}

	data := TemplateData{
		RecipientName: "John Doe",
		CouponCode:    "TEST123",
	}

	subject, htmlBody, textBody, err := tm.RenderTemplate("test_template", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	// Subject будет дефолтным для неизвестных шаблонов
	expectedSubject := "Уведомление от сервиса алмазной мозаики"
	if subject != expectedSubject {
		t.Errorf("Expected subject '%s', got '%s'", expectedSubject, subject)
	}

	if htmlBody == "" {
		t.Error("HTML body is empty")
	}

	if textBody == "" {
		t.Error("Text body is empty")
	}

	t.Logf("Subject: %s", subject)
	t.Logf("HTML Body: %s", htmlBody)
	t.Logf("Text Body: %s", textBody)
}

func TestTemplateData_AllFields(t *testing.T) {
	// Проверяем, что все поля структуры TemplateData доступны
	data := TemplateData{
		RecipientName:  "John Doe",
		RecipientEmail: "john@example.com",
		SupportEmail:   "support@example.com",
		CompanyName:    "Mosaic Company",
		WebsiteURL:     "https://mosaic.example.com",
		CouponCode:     "TEST123",
		SchemaURL:      "https://s3.example.com/schema.pdf",
		PreviewURL:     "https://s3.example.com/preview.jpg",
		OriginalURL:    "https://s3.example.com/original.jpg",
		Size:           "30x40",
		Style:          "Классический",
		ProcessingTime: "15 минут",
		ErrorMessage:   "Test error",
		ErrorCode:      "E001",
		RetryURL:       "https://example.com/retry",
		Status:         "processing",
		StatusMessage:  "Обработка изображения",
		Progress:       75,
		Timestamp:      "01.01.2024 12:00:00",
	}

	if data.RecipientName == "" {
		t.Error("RecipientName is empty")
	}

	if data.Progress != 75 {
		t.Errorf("Expected Progress 75, got %d", data.Progress)
	}

	fields := map[string]interface{}{
		"RecipientEmail": data.RecipientEmail,
		"SupportEmail":   data.SupportEmail,
		"CompanyName":    data.CompanyName,
		"WebsiteURL":     data.WebsiteURL,
		"CouponCode":     data.CouponCode,
		"SchemaURL":      data.SchemaURL,
		"PreviewURL":     data.PreviewURL,
		"OriginalURL":    data.OriginalURL,
		"Size":           data.Size,
		"Style":          data.Style,
		"ProcessingTime": data.ProcessingTime,
		"ErrorMessage":   data.ErrorMessage,
		"ErrorCode":      data.ErrorCode,
		"RetryURL":       data.RetryURL,
		"Status":         data.Status,
		"StatusMessage":  data.StatusMessage,
		"Timestamp":      data.Timestamp,
	}

	for fieldName, fieldValue := range fields {
		if fieldValue == "" {
			t.Errorf("Field %s is empty", fieldName)
		}
	}

	t.Logf("TemplateData structure is valid with %d%% progress and %d fields", data.Progress, len(fields)+2)
}
