// pkg/mail/mail.go
package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"

	"github.com/rs/zerolog/log"
	"github.com/skr1ms/mosaic/config"
)

type Mailer struct {
	cfg *config.Config
}

func NewMailer(cfg *config.Config) *Mailer {
	return &Mailer{cfg: cfg}
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
