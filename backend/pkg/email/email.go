package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"github.com/skr1ms/mosaic/config"
	"github.com/skr1ms/mosaic/pkg/middleware"
)

type Mailer struct {
	cfg       *config.Config
	templates *TemplateManager
	logger    *middleware.Logger
}

func NewMailer(cfg *config.Config, logger *middleware.Logger) *Mailer {
	templateManager := NewTemplateManager("/app/pkg/email/templates", logger)

	return &Mailer{
		cfg:       cfg,
		templates: templateManager,
		logger:    logger,
	}
}

func (m *Mailer) SendResetPasswordEmail(to, resetLink string) error {
	m.logger.GetZerologLogger().Info().Str("to", to).Msg("Sending reset password email")

	subject := "Password Reset"
	htmlBody := fmt.Sprintf(
		"<h2>Password Reset</h2>"+
			"<p>To reset your password, click the link below:</p>"+
			"<a href=\"%s\">Reset Password</a>"+
			"<p><small>Link is valid for 1 hour</small></p>",
		resetLink,
	)

	textBody := fmt.Sprintf(
		"Password Reset\n\n"+
			"To reset your password, click the link below:\n"+
			"%s\n\n"+
			"Link is valid for 1 hour",
		resetLink,
	)

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// SendSchemaEmail sends completed mosaic schema to user's email
func (m *Mailer) SendSchemaEmail(to, schemaURL, couponCode string) error {
	m.logger.GetZerologLogger().Info().Str("to", to).Msg("Sending schema email")

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		SchemaURL:      schemaURL,
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("schema_ready", templateData)
	if err != nil {
		m.logger.GetZerologLogger().Error().Err(err).Msg("Failed to render schema_ready template, using fallback")
		return m.sendSchemaEmailFallback(to, schemaURL, couponCode)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// sendSchemaEmailFallback fallback method with hardcoded HTML template
func (m *Mailer) sendSchemaEmailFallback(to, schemaURL, couponCode string) error {
	subject := "🎨 Your diamond mosaic schema is ready!"
	htmlBody := fmt.Sprintf(
		"<h2>🎨 Your diamond mosaic schema is ready!</h2>"+
			"<p>Hello!</p>"+
			"<p>Your personal diamond mosaic schema for coupon <strong>%s</strong> has been successfully created and is ready for download.</p>"+
			"<p><a href=\"%s\" style=\"background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; display: inline-block; border-radius: 4px;\">📥 Download Schema</a></p>"+
			"<p><strong>What's included in the ZIP archive:</strong></p>"+
			"<ul>"+
			"<li>🖼️ <strong>original.jpg</strong> - Your original image</li>"+
			"<li>👁️ <strong>preview.png</strong> - Preview of the finished mosaic</li>"+
			"<li>🧩 <strong>mosaic_scheme.png</strong> - Mosaic schema (colored cells + symbols)</li>"+
			"</ul>"+
			"<p><em>Happy creating! 🎨✨</em></p>"+
			"<hr>"+
			"<p><small>Download link is valid for 30 days. If you have any questions, please contact our support team.</small></p>",
		couponCode, schemaURL,
	)

	textBody := fmt.Sprintf(
		"Your diamond mosaic schema is ready!\n\n"+
			"Coupon: %s\n"+
			"Download link: %s\n\n"+
			"What's included in the ZIP archive:\n"+
			"- original.jpg - Your original image\n"+
			"- preview.png - Preview of the finished mosaic\n"+
			"- mosaic_scheme.png - Mosaic schema (colored cells + symbols)\n\n"+
			"Happy creating!\n\n"+
			"Download link is valid for 30 days.",
		couponCode, schemaURL,
	)

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// sendEmail universal method for sending emails with SMTP
func (m *Mailer) sendEmail(to, subject, htmlBody, textBody string) error {
	err := m.sendEmailViaSMTP(to, subject, htmlBody, textBody, m.cfg.SMTPConfig.Host, fmt.Sprintf("%d", m.cfg.SMTPConfig.Port), m.cfg.SMTPConfig.Username, m.cfg.SMTPConfig.Password, m.cfg.SMTPConfig.From, m.cfg.SMTPConfig.SSL)
	if err != nil {
		m.logger.GetZerologLogger().Warn().Err(err).Str("host", m.cfg.SMTPConfig.Host).Msg("Primary SMTP failed, trying test fallback")

		gmailErr := m.sendEmailViaSMTP(to, subject, htmlBody, textBody, "smtp.mailtrap.io", "2525", "test", "test", "test@example.com", false)
		if gmailErr != nil {
			m.logger.GetZerologLogger().Error().Err(gmailErr).Msg("Test SMTP fallback also failed")
			m.logger.GetZerologLogger().Error().Err(err).Err(gmailErr).Msg("Both SMTP servers failed")
			return fmt.Errorf("both SMTP servers failed: primary: %w, fallback: %w", err, gmailErr)
		}
		m.logger.GetZerologLogger().Info().Msg("Email sent successfully via test SMTP fallback")
		return nil
	}

	m.logger.GetZerologLogger().Info().Str("to", to).Str("subject", subject).Msg("Email sent successfully via primary SMTP")
	return nil
}

func (m *Mailer) sendEmailViaSMTP(to, subject, htmlBody, textBody, host, port, username, password, from string, useSSL bool) error {
	auth := smtp.PlainAuth("", username, password, host)

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
		from, to, subject, textBody, htmlBody,
	)

	portNum := port
	if portNum == "" {
		if useSSL {
			portNum = "465"
		} else {
			portNum = "587"
		}
	}

	if useSSL {
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
		}

		dialer := &net.Dialer{
			Timeout: 30 * time.Second,
		}

		conn, err := dialer.Dial("tcp", net.JoinHostPort(host, portNum))
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("TCP connection failed")
			return fmt.Errorf("TCP connection failed: %w", err)
		}
		defer conn.Close()

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}

		tlsConn := tls.Client(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("TLS handshake failed")
			return fmt.Errorf("TLS handshake failed: %w", err)
		}
		defer tlsConn.Close()

		client, err := smtp.NewClient(tlsConn, host)
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP client creation failed")
			return fmt.Errorf("SMTP client creation failed: %w", err)
		}
		defer client.Close()

		if err := client.Hello("localhost"); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP HELO failed")
			return fmt.Errorf("SMTP HELO failed: %w", err)
		}

		if err := client.Auth(auth); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP authentication failed")
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err := client.Mail(from); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Str("from", from).Msg("SMTP sender setup failed")
			return fmt.Errorf("SMTP sender setup failed: %w", err)
		}

		if err := client.Rcpt(to); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Str("to", to).Msg("SMTP recipient setup failed")
			return fmt.Errorf("SMTP recipient setup failed: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP data writer failed")
			return fmt.Errorf("SMTP data writer failed: %w", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(msg))
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP message writing failed")
			return fmt.Errorf("SMTP message writing failed: %w", err)
		}

		return nil
	} else {
		conn, err := smtp.Dial(net.JoinHostPort(host, portNum))
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP connection failed")
			return fmt.Errorf("SMTP connection failed: %w", err)
		}
		defer conn.Close()

		if err := conn.StartTLS(&tls.Config{ServerName: host}); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("STARTTLS failed")
			return fmt.Errorf("STARTTLS failed: %w", err)
		}

		if err := conn.Auth(auth); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP authentication failed")
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err := conn.Mail(from); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Str("from", from).Msg("SMTP sender setup failed")
			return fmt.Errorf("SMTP sender setup failed: %w", err)
		}

		if err := conn.Rcpt(to); err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Str("to", to).Msg("SMTP recipient setup failed")
			return fmt.Errorf("SMTP recipient setup failed: %w", err)
		}

		w, err := conn.Data()
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP data writer failed")
			return fmt.Errorf("SMTP data writer failed: %w", err)
		}
		defer w.Close()

		_, err = w.Write([]byte(msg))
		if err != nil {
			m.logger.GetZerologLogger().Error().Err(err).Str("host", host).Str("port", portNum).Msg("SMTP message writing failed")
			return fmt.Errorf("SMTP message writing failed: %w", err)
		}

		return nil
	}
}

func (m *Mailer) SendProcessingErrorEmail(to, couponCode, errorMessage string) error {
	m.logger.GetZerologLogger().Info().Str("to", to).Msg("Sending processing error email")

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		ErrorMessage:   errorMessage,
		Timestamp:      time.Now().Format("02.01.2006 15:04:05"),
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("processing_error", templateData)
	if err != nil {
		m.logger.GetZerologLogger().Error().Err(err).Msg("Failed to render processing_error template")
		subject = "❌ Error processing your image"
		htmlBody = fmt.Sprintf(
			"<h2>❌ Error processing your image</h2>"+
				"<p>Unfortunately, an error occurred while processing your image for coupon <strong>%s</strong>.</p>"+
				"<p><strong>Error:</strong> %s</p>"+
				"<p>Please try uploading the image again or contact support.</p>",
			couponCode, errorMessage,
		)
		textBody = fmt.Sprintf(
			"Error processing your image\n\n"+
				"Unfortunately, an error occurred while processing your image for coupon %s.\n\n"+
				"Error: %s\n\n"+
				"Please try uploading the image again or contact support.",
			couponCode, errorMessage,
		)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}

func (m *Mailer) SendStatusUpdateEmail(to, couponCode, status, message string) error {
	m.logger.GetZerologLogger().Info().Str("to", to).Msg("Sending status update email")

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		Status:         status,
		StatusMessage:  message,
		Timestamp:      time.Now().Format("02.01.2006 15:04:05"),
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("status_update", templateData)
	if err != nil {
		m.logger.GetZerologLogger().Error().Err(err).Msg("Failed to render status_update template")
		subject = "ℹ️ Order status update"
		htmlBody = fmt.Sprintf(
			"<h2>ℹ️ Order status update</h2>"+
				"<p>The status of your order for coupon <strong>%s</strong> has changed.</p>"+
				"<p><strong>New status:</strong> %s</p>"+
				"<p>%s</p>",
			couponCode, status, message,
		)
		textBody = fmt.Sprintf(
			"Order status update\n\n"+
				"The status of your order for coupon %s has changed.\n\n"+
				"New status: %s\n\n%s",
			couponCode, status, message,
		)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// SendCouponPurchaseEmail sends coupon code to user after successful purchase
func (m *Mailer) SendCouponPurchaseEmail(to, couponCode, size, style string) error {
	m.logger.GetZerologLogger().Info().Str("to", to).Str("coupon", couponCode).Msg("Sending coupon purchase email")

	templateData := TemplateData{
		RecipientEmail: to,
		CouponCode:     couponCode,
		Size:           size,
		StyleName:      m.getStyleName(style),
	}

	subject, htmlBody, textBody, err := m.templates.RenderTemplate("coupon_purchased", templateData)
	if err != nil {
		m.logger.GetZerologLogger().Error().Err(err).Msg("Failed to render coupon_purchased template, using fallback")
		return m.sendCouponPurchaseEmailFallback(to, couponCode, size, style)
	}

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// sendCouponPurchaseEmailFallback fallback method with hardcoded HTML template
func (m *Mailer) sendCouponPurchaseEmailFallback(to, couponCode, size, style string) error {
	subject := "🎉 Ваш купон готов к использованию!"
	htmlBody := fmt.Sprintf(
		"<h2>🎉 Ваш купон готов к использованию!</h2>"+
			"<p>Спасибо за покупку!</p>"+
			"<p>Ваш купон: <strong style=\"font-size: 18px; color: #4CAF50;\">%s</strong></p>"+
			"<p><strong>Размер:</strong> %s см</p>"+
			"<p><strong>Стиль:</strong> %s</p>"+
			"<p>Теперь вы можете:</p>"+
			"<ol>"+
			"<li>✨ Ввести код купона: <strong>%s</strong></li>"+
			"<li>🖼️ Загрузить свое изображение на сайте</li>"+
			"<li>🎨 Получить готовую схему алмазной мозаики</li>"+
			"</ol>"+
			"<p><a href=\"https://photo.doyoupaint.com/\" style=\"background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; display: inline-block; border-radius: 4px; margin: 10px 0;\">🚀 Создать схему</a></p>"+
			"<p><em>Удачного творчества! 🎨✨</em></p>"+
			"<hr>"+
			"<p><small>Сохраните этот код купона. Если у вас есть вопросы, обратитесь в службу поддержки.</small></p>",
		couponCode, size, m.getStyleName(style), couponCode,
	)

	textBody := fmt.Sprintf(
		"Ваш купон готов к использованию!\n\n"+
			"Спасибо за покупку!\n\n"+
			"Ваш купон: %s\n"+
			"Размер: %s см\n"+
			"Стиль: %s\n\n"+
			"Теперь вы можете:\n"+
			"1. Ввести код купона: %s\n"+
			"2. Загрузить свое изображение на сайте\n"+
			"3. Получить готовую схему алмазной мозаики\n\n"+
			"Перейти на сайт: https://photo.doyoupaint.com/\n\n"+
			"Удачного творчества!\n\n"+
			"Сохраните этот код купона.",
		couponCode, size, m.getStyleName(style), couponCode,
	)

	return m.sendEmail(to, subject, htmlBody, textBody)
}

// getStyleName converts style code to readable name
func (m *Mailer) getStyleName(style string) string {
	switch style {
	case "grayscale":
		return "Оттенки серого"
	case "skin_tone":
		return "Оттенки телесного"
	case "pop_art":
		return "Поп-арт"
	case "max_colors":
		return "Максимум цветов"
	default:
		return style
	}
}
