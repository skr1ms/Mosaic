// pkg/mail/mail.go
package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
)

type Mailer struct {
	cfg       *config.Config
	templates *TemplateManager
}

func NewMailer(cfg *config.Config) *Mailer {
	// Инициализируем систему шаблонов с новым путем
	templateManager := NewTemplateManager("backend/pkg/email/templates")

	return &Mailer{
		cfg:       cfg,
		templates: templateManager,
	}
}

func (m *Mailer) SendResetPasswordEmail(to, resetLink string) error {
	log.Info().Msg("Sending reset password email to " + to)
	auth := smtp.PlainAuth("",
		m.cfg.SMTPConfig.Username,
		m.cfg.SMTPConfig.Password,
		m.cfg.SMTPConfig.Host,
	)
	log.Info().Msg("Authentication successful")

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: Сброс пароля\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"<h2>Сброс пароля</h2>"+
			"<p>Для сброса пароля перейдите по ссылке:</p>"+
			"<a href=\"%s\">Сбросить пароль</a>"+
			"<p><small>Ссылка действительна 1 час</small></p>",
		m.cfg.SMTPConfig.From, to, resetLink,
	))

	tlsConfig := &tls.Config{
		ServerName: m.cfg.SMTPConfig.Host,
	}

	conn, err := tls.Dial("tcp", net.JoinHostPort(m.cfg.SMTPConfig.Host, "465"), tlsConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to SMTP server")
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	log.Info().Msg("SMTP client creation successful")

	client, err := smtp.NewClient(conn, m.cfg.SMTPConfig.Host)
	if err != nil {
		log.Error().Err(err).Msg("SMTP client creation failed")
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	log.Info().Msg("SMTP client authentication successful")

	if err := client.Auth(auth); err != nil {
		log.Error().Err(err).Msg("authentication failed")
		return fmt.Errorf("authentication failed: %w", err)
	}

	log.Info().Msg("SMTP client sender setup successful")

	if err := client.Mail(m.cfg.SMTPConfig.From); err != nil {
		log.Error().Err(err).Msg("sender setup failed")
		return fmt.Errorf("sender setup failed: %w", err)
	}

	log.Info().Msg("SMTP client recipient setup successful")

	if err := client.Rcpt(to); err != nil {
		log.Error().Err(err).Msg("recipient setup failed")
		return fmt.Errorf("recipient setup failed: %w", err)
	}

	log.Info().Msg("SMTP client data writer successful")

	w, err := client.Data()
	if err != nil {
		log.Error().Err(err).Msg("data writer failed")
		return fmt.Errorf("data writer failed: %w", err)
	}

	log.Info().Msg("SMTP client message writing successful")

	if _, err := w.Write(msg); err != nil {
		log.Error().Err(err).Msg("message writing failed")
		return fmt.Errorf("message writing failed: %w", err)
	}

	log.Info().Msg("SMTP client writer close successful")

	if err := w.Close(); err != nil {
		log.Error().Err(err).Msg("writer close failed")
		return fmt.Errorf("writer close failed: %w", err)
	}

	log.Info().Msg("SMTP client quit successful")

	return client.Quit()
}

// SendSchemaEmail отправляет готовую схему мозаики на email пользователя
func (m *Mailer) SendSchemaEmail(to, schemaURL, couponCode string) error {
	log.Info().Msg("Sending schema email to " + to)

	// Подготавливаем данные для шаблона
	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		SchemaURL:      schemaURL,
	}

	// Рендерим шаблон
	subject, htmlBody, textBody, err := m.templates.RenderTemplate("schema_ready", templateData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render schema_ready template, using fallback")
		// Используем старый хардкод HTML как fallback
		return m.sendSchemaEmailFallback(to, schemaURL, couponCode)
	}

	// Отправляем email с шаблоном
	return m.sendEmail(to, subject, htmlBody, textBody)
}

// sendSchemaEmailFallback - fallback метод с хардкод HTML
func (m *Mailer) sendSchemaEmailFallback(to, schemaURL, couponCode string) error {
	subject := "🎨 Ваша схема алмазной мозаики готова!"
	htmlBody := fmt.Sprintf(
		"<h2>🎨 Ваша схема алмазной мозаики готова!</h2>"+
			"<p>Здравствуйте!</p>"+
			"<p>Ваша персональная схема алмазной мозаики по купону <strong>%s</strong> успешно создана и готова к скачиванию.</p>"+
			"<p><a href=\"%s\" style=\"background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; display: inline-block; border-radius: 4px;\">📥 Скачать схему</a></p>"+
			"<p><strong>Что входит в ZIP-архив:</strong></p>"+
			"<ul>"+
			"<li>📄 <strong>schema.pdf</strong> - Подробная схема алмазной мозаики с цветовой картой</li>"+
			"<li>🖼️ <strong>original.jpg</strong> - Ваше оригинальное изображение</li>"+
			"<li>👁️ <strong>preview.jpg</strong> - Превью готовой мозаики</li>"+
			"</ul>"+
			"<p><em>Приятного творчества! 🎨✨</em></p>"+
			"<hr>"+
			"<p><small>Ссылка для скачивания действительна в течение 30 дней. Если у вас возникли вопросы, свяжитесь с нашей службой поддержки.</small></p>",
		couponCode, schemaURL,
	)

	textBody := fmt.Sprintf(
		"Ваша схема алмазной мозаики готова!\n\n"+
			"Купон: %s\n"+
			"Ссылка для скачивания: %s\n\n"+
			"Что входит в ZIP-архив:\n"+
			"- schema.pdf - Подробная схема алмазной мозаики с цветовой картой\n"+
			"- original.jpg - Ваше оригинальное изображение\n"+
			"- preview.jpg - Превью готовой мозаики\n\n"+
			"Приятного творчества!\n\n"+
			"Ссылка для скачивания действительна в течение 30 дней.",
		couponCode, schemaURL,
	)

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// sendEmail - универсальный метод для отправки email
func (m *Mailer) sendEmail(to, subject, htmlBody, textBody string) error {
	auth := smtp.PlainAuth("",
		m.cfg.SMTPConfig.Username,
		m.cfg.SMTPConfig.Password,
		m.cfg.SMTPConfig.Host,
	)

	// Формируем MIME сообщение с поддержкой HTML и текста
	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n\r\n"+
			"--boundary123\r\n"+
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n"+
			"%s\r\n\r\n"+
			"--boundary123\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n\r\n"+
			"%s\r\n\r\n"+
			"--boundary123--\r\n",
		m.cfg.SMTPConfig.From, to, subject, textBody, htmlBody,
	)

	tlsConfig := &tls.Config{
		ServerName: m.cfg.SMTPConfig.Host,
	}

	conn, err := tls.Dial("tcp", net.JoinHostPort(m.cfg.SMTPConfig.Host, "465"), tlsConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to SMTP server")
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.SMTPConfig.Host)
	if err != nil {
		log.Error().Err(err).Msg("SMTP client creation failed")
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		log.Error().Err(err).Msg("authentication failed")
		return fmt.Errorf("authentication failed: %w", err)
	}

	if err := client.Mail(m.cfg.SMTPConfig.From); err != nil {
		log.Error().Err(err).Msg("sender setup failed")
		return fmt.Errorf("sender setup failed: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		log.Error().Err(err).Msg("recipient setup failed")
		return fmt.Errorf("recipient setup failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		log.Error().Err(err).Msg("data writer failed")
		return fmt.Errorf("data writer failed: %w", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		log.Error().Err(err).Msg("message writing failed")
		return fmt.Errorf("message writing failed: %w", err)
	}

	if err := w.Close(); err != nil {
		log.Error().Err(err).Msg("writer close failed")
		return fmt.Errorf("writer close failed: %w", err)
	}

	log.Info().Msg("Email sent successfully to " + to)
	return client.Quit()
}

// SendProcessingErrorEmail отправляет уведомление об ошибке обработки
func (m *Mailer) SendProcessingErrorEmail(to, couponCode, errorMessage string) error {
	log.Info().Msg("Sending processing error email to " + to)

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		ErrorMessage:   errorMessage,
		Timestamp:      time.Now().Format("02.01.2006 15:04:05"),
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("processing_error", templateData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render processing_error template")
		// Fallback
		subject = "❌ Ошибка при обработке изображения"
		htmlBody = fmt.Sprintf(
			"<h2>❌ Ошибка при обработке изображения</h2>"+
				"<p>К сожалению, при обработке вашего изображения по купону <strong>%s</strong> произошла ошибка.</p>"+
				"<p><strong>Ошибка:</strong> %s</p>"+
				"<p>Пожалуйста, попробуйте загрузить изображение заново или обратитесь в службу поддержки.</p>",
			couponCode, errorMessage,
		)
		textBody = fmt.Sprintf(
			"Ошибка при обработке изображения\n\n"+
				"К сожалению, при обработке вашего изображения по купону %s произошла ошибка.\n\n"+
				"Ошибка: %s\n\n"+
				"Пожалуйста, попробуйте загрузить изображение заново или обратитесь в службу поддержки.",
			couponCode, errorMessage,
		)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// SendStatusUpdateEmail отправляет уведомление об изменении статуса
func (m *Mailer) SendStatusUpdateEmail(to, couponCode, status, message string) error {
	log.Info().Msg("Sending status update email to " + to)

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		Status:         status,
		StatusMessage:  message,
		Timestamp:      time.Now().Format("02.01.2006 15:04:05"),
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("status_update", templateData)
	if err != nil {
		log.Error().Err(err).Msg("Failed to render status_update template")
		// Fallback
		subject = "ℹ️ Обновление статуса заказа"
		htmlBody = fmt.Sprintf(
			"<h2>ℹ️ Обновление статуса заказа</h2>"+
				"<p>Статус вашего заказа по купону <strong>%s</strong> изменился.</p>"+
				"<p><strong>Новый статус:</strong> %s</p>"+
				"<p>%s</p>",
			couponCode, status, message,
		)
		textBody = fmt.Sprintf(
			"Обновление статуса заказа\n\n"+
				"Статус вашего заказа по купону %s изменился.\n\n"+
				"Новый статус: %s\n\n%s",
			couponCode, status, message,
		)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}
