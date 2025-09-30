package email

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

// EmailTemplate represents email template
type EmailTemplate struct {
	Name     string
	Subject  string
	HTMLBody string
	TextBody string
	template *template.Template
}

// TemplateData represents data for template
type TemplateData struct {
	RecipientName  string
	RecipientEmail string
	SupportEmail   string
	CompanyName    string
	WebsiteURL     string

	CouponCode     string
	SchemaURL      string
	PreviewURL     string
	OriginalURL    string
	Size           string
	Style          string
	StyleName      string
	ProcessingTime string

	ErrorMessage string
	ErrorCode    string
	RetryURL     string

	Status        string
	StatusMessage string
	Progress      int
	Timestamp     string
	ImageID       string

	PartnerName    string
	PartnerLogo    string
	PartnerEmail   string
	PartnerPhone   string
	PartnerWebsite string
}

// TemplateManager manages email templates
type TemplateManager struct {
	templates    map[string]*EmailTemplate
	templatesDir string
	logger       *middleware.Logger
}

// NewTemplateManager creates new template manager
func NewTemplateManager(templatesDir string, logger *middleware.Logger) *TemplateManager {
	if templatesDir == "" {
		templatesDir = "templates"
	}

	tm := &TemplateManager{
		templates:    make(map[string]*EmailTemplate),
		templatesDir: templatesDir,
		logger:       logger,
	}

	if err := tm.loadTemplates(); err != nil {
		tm.logger.GetZerologLogger().Error().Err(err).Msg("Failed to load email templates")
	}

	return tm
}

// loadTemplates loads all templates from filesystem
func (tm *TemplateManager) loadTemplates() error {
	templateFiles, err := os.ReadDir(tm.templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			tm.logger.GetZerologLogger().Warn().Str("dir", tm.templatesDir).Msg("Templates directory does not exist, creating it")
			if err := os.MkdirAll(tm.templatesDir, 0755); err != nil {
				tm.logger.GetZerologLogger().Error().Err(err).Str("dir", tm.templatesDir).Msg("Failed to create templates directory")
				return fmt.Errorf("failed to create templates directory: %w", err)
			}
			return nil
		}
		tm.logger.GetZerologLogger().Error().Err(err).Str("dir", tm.templatesDir).Msg("Failed to read templates directory")
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
			tm.logger.GetZerologLogger().Error().Err(err).Str("template", templateName).Msg("Failed to read template file")
			continue
		}

		tmpl, err := template.New(templateName).Parse(string(content))
		if err != nil {
			tm.logger.GetZerologLogger().Error().Err(err).Str("template", templateName).Msg("Failed to parse template")
			continue
		}

		subject := tm.getSubjectForTemplate(templateName)

		emailTemplate := &EmailTemplate{
			Name:     templateName,
			Subject:  subject,
			HTMLBody: string(content),
			template: tmpl,
		}

		tm.templates[templateName] = emailTemplate
		tm.logger.GetZerologLogger().Info().Str("template", templateName).Msg("Email template loaded successfully")
	}

	return nil
}

func (tm *TemplateManager) getSubjectForTemplate(templateName string) string {
	subjects := map[string]string{
		"schema_ready":       "🎨 Your diamond mosaic schema is ready!",
		"processing_error":   "❌ Error processing your image",
		"status_update":      "📊 Update on your order status",
		"welcome":            "👋 Welcome to the world of diamond mosaic!",
		"password_reset":     "🔐 Password reset",
		"order_confirmation": "✅ Order confirmation",
		"payment_success":    "💳 Payment successful",
		"payment_failed":     "❌ Payment failed",
	}

	if subject, exists := subjects[templateName]; exists {
		return subject
	}

	return "Notification from diamond mosaic service"
}

func (tm *TemplateManager) GetTemplate(name string) (*EmailTemplate, error) {
	template, exists := tm.templates[name]
	if !exists {
		tm.logger.GetZerologLogger().Error().Str("template_name", name).Msg("Email template not found")
		return nil, fmt.Errorf("template '%s' not found", name)
	}

	return template, nil
}

func (tm *TemplateManager) RenderTemplate(templateName string, data TemplateData) (string, string, string, error) {
	template, err := tm.GetTemplate(templateName)
	if err != nil {
		return "", "", "", err
	}

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
		tm.logger.GetZerologLogger().Error().Err(err).Str("template_name", templateName).Msg("Failed to render email template")
		return "", "", "", fmt.Errorf("failed to render template: %w", err)
	}

	textBody := tm.htmlToText(htmlBuffer.String())

	return template.Subject, htmlBuffer.String(), textBody, nil
}

// htmlToText converts HTML to text (simplified version)
func (tm *TemplateManager) htmlToText(html string) string {
	text := html

	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	text = strings.ReplaceAll(text, "</p>", "\n\n")
	text = strings.ReplaceAll(text, "</div>", "\n")
	text = strings.ReplaceAll(text, "</h1>", "\n")
	text = strings.ReplaceAll(text, "</h2>", "\n")
	text = strings.ReplaceAll(text, "</h3>", "\n")

	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}

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

// ListTemplates returns list of all available templates
func (tm *TemplateManager) ListTemplates() []string {
	var names []string
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// ReloadTemplates reloads all templates
func (tm *TemplateManager) ReloadTemplates() error {
	tm.templates = make(map[string]*EmailTemplate)
	return tm.loadTemplates()
}
