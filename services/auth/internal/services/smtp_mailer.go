package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/mail"
	"net/smtp"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"go.uber.org/zap"
)

type smtpConfig struct {
	EnableCustomSMTP bool
	SenderEmail      string
	SenderName       string
	Host             string
	Port             int
	Username         string
	Password         string
	Secure           bool
}

type SmtpMailer struct {
	db  *pgxpool.Pool
	log *zap.Logger
	cfg *config.Config
}

func NewSmtpMailer(db *pgxpool.Pool, log *zap.Logger, cfg *config.Config) *SmtpMailer {
	return &SmtpMailer{db: db, log: log, cfg: cfg}
}

func (m *SmtpMailer) getSmtpConfig(ctx context.Context) (*smtpConfig, error) {
	var conf smtpConfig
	err := m.db.QueryRow(ctx, `
		SELECT enable_custom_smtp, sender_email, sender_name, host, port, username, password, secure 
		FROM auth.smtp_settings WHERE project_id = 'default'
	`).Scan(
		&conf.EnableCustomSMTP, &conf.SenderEmail, &conf.SenderName,
		&conf.Host, &conf.Port, &conf.Username, &conf.Password, &conf.Secure,
	)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}

func (m *SmtpMailer) getTemplate(ctx context.Context, templateType string) (subject, bodyHTML, bodyText string, err error) {
	err = m.db.QueryRow(ctx, `
		SELECT subject, body_html, body_text 
		FROM auth.email_templates WHERE type = $1
	`, templateType).Scan(&subject, &bodyHTML, &bodyText)
	return
}

func (m *SmtpMailer) send(to, templateType string, data map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conf, err := m.getSmtpConfig(ctx)
	if err != nil || !conf.EnableCustomSMTP || conf.Host == "" {
		m.log.Info("Custom SMTP not enabled or configured, skipping email send", zap.String("to", to), zap.String("type", templateType))
		return nil
	}

	subjectTmpl, bodyHTMLTmpl, bodyTextTmpl, err := m.getTemplate(ctx, templateType)
	if err != nil {
		m.log.Error("Failed to fetch email template", zap.Error(err), zap.String("type", templateType))
		return err
	}

	// Interpolate templates
	subj, err := m.render(subjectTmpl, data)
	if err != nil {
		return err
	}
	html, err := m.render(bodyHTMLTmpl, data)
	if err != nil {
		return err
	}
	text, err := m.render(bodyTextTmpl, data)
	if err != nil {
		return err
	}

	from := mail.Address{Name: conf.SenderName, Address: conf.SenderEmail}
	recipient := mail.Address{Address: to}

	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = recipient.String()
	headers["Subject"] = subj
	headers["MIME-Version"] = "1.0"

	boundary := "mixed-boundary"
	headers["Content-Type"] = fmt.Sprintf("multipart/alternative; boundary=\"%s\"", boundary)

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")

	// Text Part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(text + "\r\n")

	// HTML Part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	msg.WriteString(html + "\r\n")
	msg.WriteString(fmt.Sprintf("--%s--", boundary))

	// Connect to SMTP
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	var auth smtp.Auth
	if conf.Username != "" {
		auth = smtp.PlainAuth("", conf.Username, conf.Password, conf.Host)
	}

	if conf.Secure {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         conf.Host,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		c, err := smtp.NewClient(conn, conf.Host)
		if err != nil {
			return fmt.Errorf("SMTP client error: %w", err)
		}
		defer c.Quit()

		if auth != nil {
			if err = c.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth failed: %w", err)
			}
		}

		if err = c.Mail(conf.SenderEmail); err != nil {
			return err
		}
		if err = c.Rcpt(to); err != nil {
			return err
		}

		w, err := c.Data()
		if err != nil {
			return err
		}
		_, err = w.Write(msg.Bytes())
		if err != nil {
			return err
		}
		return w.Close()
	}

	// Non-secure (starttls or plain)
	return smtp.SendMail(addr, auth, conf.SenderEmail, []string{to}, msg.Bytes())
}

func (m *SmtpMailer) render(tmplStr string, data interface{}) (string, error) {
	t, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Mailer Interface Implementation

func (m *SmtpMailer) SendConfirmation(to, name, confirmURL string) error {
	data := map[string]interface{}{
		"ConfirmationURL": confirmURL,
		"Email":          to,
		"Name":           name,
		"SiteURL":        m.cfg.SiteURL,
	}
	return m.send(to, "signup", data)
}

func (m *SmtpMailer) SendPasswordReset(to, name, resetURL string) error {
	data := map[string]interface{}{
		"ConfirmationURL": resetURL, // Using same key as signup for simplicity in templates
		"Email":          to,
		"Name":           name,
		"SiteURL":        m.cfg.SiteURL,
	}
	return m.send(to, "reset_password", data)
}

func (m *SmtpMailer) SendMagicLink(to, magicURL string) error {
	data := map[string]interface{}{
		"ConfirmationURL": magicURL,
		"Email":          to,
		"SiteURL":        m.cfg.SiteURL,
	}
	return m.send(to, "magic_link", data)
}

// Additional Methods for security alerts
func (m *SmtpMailer) SendInvite(to, inviteURL string) error {
	data := map[string]interface{}{
		"ConfirmationURL": inviteURL,
		"Email":          to,
		"SiteURL":        m.cfg.SiteURL,
	}
	return m.send(to, "invite", data)
}

func (m *SmtpMailer) SendSecurityAlert(to, alertType string, details map[string]interface{}) error {
	data := map[string]interface{}{
		"Email":   to,
		"Details": details,
		"Type":    alertType,
		"SiteURL": m.cfg.SiteURL,
	}
	return m.send(to, alertType, data)
}
