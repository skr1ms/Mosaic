package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// EmailTemplate представляет шаблон email
type EmailTemplate struct {
	Name     string
	Subject  string
	HTMLBody string
	TextBody string
	template *template.Template
}

// TemplateData представляет данные для шаблона
type TemplateData struct {
	// Общие данные
	RecipientName  string
	RecipientEmail string
	SupportEmail   string
	CompanyName    string
	WebsiteURL     string

	// Данные для схемы мозаики
	CouponCode     string
	SchemaURL      string
	PreviewURL     string
	OriginalURL    string
	Size           string
	Style          string
	ProcessingTime string

	// Данные ошибки
	ErrorMessage string
	ErrorCode    string
	RetryURL     string

	// Данные статуса
	Status        string
	StatusMessage string
	Progress      int
	Timestamp     string
	ImageID       string

	// Данные партнера (для White Label)
	PartnerName    string
	PartnerLogo    string
	PartnerEmail   string
	PartnerPhone   string
	PartnerWebsite string
}

// TemplateManager управляет email шаблонами
type TemplateManager struct {
	templates    map[string]*EmailTemplate
	templatesDir string
}

// NewTemplateManager создает новый менеджер шаблонов
func NewTemplateManager(templatesDir string) *TemplateManager {
	if templatesDir == "" {
		templatesDir = "templates"
	}

	tm := &TemplateManager{
		templates:    make(map[string]*EmailTemplate),
		templatesDir: templatesDir,
	}

	if err := tm.loadTemplates(); err != nil {
		log.Error().Err(err).Msg("Failed to load email templates")
	}

	return tm
}

// loadTemplates загружает все шаблоны из файловой системы
func (tm *TemplateManager) loadTemplates() error {
	templateFiles, err := os.ReadDir(tm.templatesDir)
	if err != nil {
		// Если директория не существует, создаем её
		if os.IsNotExist(err) {
			log.Warn().Str("dir", tm.templatesDir).Msg("Templates directory does not exist, creating it")
			if err := os.MkdirAll(tm.templatesDir, 0755); err != nil {
				return fmt.Errorf("failed to create templates directory: %w", err)
			}
			return nil // Директория создана, но шаблонов пока нет
		}
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, file := range templateFiles {
		if !strings.HasSuffix(file.Name(), ".html") {
			continue
		}

		templateName := strings.TrimSuffix(file.Name(), ".html")
		templatePath := filepath.Join(tm.templatesDir, file.Name())

		content, err := os.ReadFile(templatePath)
		if err != nil {
			log.Error().Err(err).Str("template", templateName).Msg("Failed to read template file")
			continue
		}

		tmpl, err := template.New(templateName).Parse(string(content))
		if err != nil {
			log.Error().Err(err).Str("template", templateName).Msg("Failed to parse template")
			continue
		}

		// Определяем subject для каждого шаблона
		subject := tm.getSubjectForTemplate(templateName)

		emailTemplate := &EmailTemplate{
			Name:     templateName,
			Subject:  subject,
			HTMLBody: string(content),
			template: tmpl,
		}

		tm.templates[templateName] = emailTemplate
		log.Info().Str("template", templateName).Msg("Email template loaded successfully")
	}

	return nil
} // getSubjectForTemplate возвращает тему письма для шаблона
func (tm *TemplateManager) getSubjectForTemplate(templateName string) string {
	subjects := map[string]string{
		"schema_ready":       "🎨 Ваша схема алмазной мозаики готова!",
		"processing_error":   "❌ Ошибка при обработке изображения",
		"status_update":      "📊 Обновление статуса вашего заказа",
		"welcome":            "👋 Добро пожаловать в мир алмазной мозаики!",
		"password_reset":     "🔐 Сброс пароля",
		"order_confirmation": "✅ Подтверждение заказа",
		"payment_success":    "💳 Оплата прошла успешно",
		"payment_failed":     "❌ Ошибка оплаты",
	}

	if subject, exists := subjects[templateName]; exists {
		return subject
	}

	return "Уведомление от сервиса алмазной мозаики"
}

// GetTemplate возвращает шаблон по имени
func (tm *TemplateManager) GetTemplate(name string) (*EmailTemplate, error) {
	template, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

// RenderTemplate рендерит шаблон с данными
func (tm *TemplateManager) RenderTemplate(templateName string, data TemplateData) (string, string, string, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return "", "", "", err
	}

	// Устанавливаем значения по умолчанию
	if data.CompanyName == "" {
		data.CompanyName = "DoYouPaint"
	}
	if data.SupportEmail == "" {
		data.SupportEmail = "support@doyoupaint.com"
	}
	if data.WebsiteURL == "" {
		data.WebsiteURL = "https://doyoupaint.com"
	}

	var htmlBuffer bytes.Buffer
	if err := template.template.Execute(&htmlBuffer, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render template: %w", err)
	}

	// Генерируем текстовую версию из HTML (упрощенно)
	textBody := tm.htmlToText(htmlBuffer.String())

	return template.Subject, htmlBuffer.String(), textBody, nil
} // htmlToText конвертирует HTML в текст (упрощенная версия)
func (tm *TemplateManager) htmlToText(html string) string {
	// Простая конвертация HTML в текст
	text := html

	// Удаляем HTML теги
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n\n")
	text = strings.ReplaceAll(text, "</div>", "\n")
	text = strings.ReplaceAll(text, "</h1>", "\n")
	text = strings.ReplaceAll(text, "</h2>", "\n")
	text = strings.ReplaceAll(text, "</h3>", "\n")

	// Удаляем все остальные HTML теги
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}

	// Убираем лишние пробелы и переносы
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// ListTemplates возвращает список всех доступных шаблонов
func (tm *TemplateManager) ListTemplates() []string {
	var names []string
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// ReloadTemplates перезагружает все шаблоны
func (tm *TemplateManager) ReloadTemplates() error {
	tm.templates = make(map[string]*EmailTemplate)
	return tm.loadTemplates()
}
