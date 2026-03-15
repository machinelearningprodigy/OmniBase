package identity

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type EmailTemplate struct {
	Type      string    `json:"type" db:"type"`
	Subject   string    `json:"subject" db:"subject"`
	BodyHTML  string    `json:"body_html" db:"body_html"`
	BodyText  string    `json:"body_text" db:"body_text"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (s *Service) GetEmailTemplates(ctx context.Context) ([]EmailTemplate, error) {
	rows, err := s.db.Query(ctx, "SELECT type, subject, body_html, body_text, updated_at FROM auth.email_templates")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []EmailTemplate
	for rows.Next() {
		var t EmailTemplate
		if err := rows.Scan(&t.Type, &t.Subject, &t.BodyHTML, &t.BodyText, &t.UpdatedAt); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (s *Service) SaveEmailTemplate(ctx context.Context, tmpl EmailTemplate) error {
	s.log.Info("saving email template", zap.String("type", tmpl.Type), zap.String("subject", tmpl.Subject))
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.email_templates (type, subject, body_html, body_text, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (type) DO UPDATE SET 
		subject = EXCLUDED.subject, body_html = EXCLUDED.body_html, body_text = EXCLUDED.body_text, updated_at = NOW()
	`, tmpl.Type, tmpl.Subject, tmpl.BodyHTML, tmpl.BodyText)
	if err != nil {
		s.log.Error("failed to save email template", zap.Error(err), zap.String("type", tmpl.Type))
	}
	return err
}

type SMTPSettings struct {
	ProjectID        string `json:"project_id" db:"project_id"`
	EnableCustomSMTP bool   `json:"enable_custom_smtp" db:"enable_custom_smtp"`
	SenderEmail      string `json:"sender_email" db:"sender_email"`
	SenderName       string `json:"sender_name" db:"sender_name"`
	Host             string `json:"host" db:"host"`
	Port             int    `json:"port" db:"port"`
	Username         string `json:"username" db:"username"`
	Password         string `json:"password" db:"password"`
	Secure           bool      `json:"secure" db:"secure"`
	MinInterval      int       `json:"min_interval" db:"min_interval"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

func (s *Service) GetSMTPSettings(ctx context.Context, projectID string) (*SMTPSettings, error) {
	if projectID == "" {
		projectID = "default"
	}
	var set SMTPSettings
	err := s.db.QueryRow(ctx, `
		SELECT project_id, enable_custom_smtp, sender_email, sender_name, host, port, username, password, secure, min_interval, updated_at 
		FROM auth.smtp_settings WHERE project_id = $1
	`, projectID).Scan(
		&set.ProjectID, &set.EnableCustomSMTP, &set.SenderEmail, &set.SenderName,
		&set.Host, &set.Port, &set.Username, &set.Password, &set.Secure, &set.MinInterval, &set.UpdatedAt,
	)
	if err != nil {
		// return default empty settings if not found
		return &SMTPSettings{ProjectID: projectID, MinInterval: 1}, nil
	}
	return &set, nil
}

func (s *Service) SaveSMTPSettings(ctx context.Context, set SMTPSettings) error {
	if set.ProjectID == "" {
		set.ProjectID = "default"
	}
	s.log.Info("saving smtp settings", zap.String("project_id", set.ProjectID), zap.String("host", set.Host))
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.smtp_settings (project_id, enable_custom_smtp, sender_email, sender_name, host, port, username, password, secure, min_interval, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (project_id) DO UPDATE SET 
		enable_custom_smtp = EXCLUDED.enable_custom_smtp, sender_email = EXCLUDED.sender_email, sender_name = EXCLUDED.sender_name, 
		host = EXCLUDED.host, port = EXCLUDED.port, username = EXCLUDED.username, password = EXCLUDED.password, 
		secure = EXCLUDED.secure, min_interval = EXCLUDED.min_interval, updated_at = NOW()
	`, set.ProjectID, set.EnableCustomSMTP, set.SenderEmail, set.SenderName, set.Host, set.Port, set.Username, set.Password, set.Secure, set.MinInterval)
	if err != nil {
		s.log.Error("failed to save smtp settings", zap.Error(err), zap.String("project_id", set.ProjectID))
	}
	return err
}
