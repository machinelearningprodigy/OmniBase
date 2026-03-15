package identity

import "context"

type EmailTemplate struct {
	Type      string `json:"type" db:"type"`
	Subject   string `json:"subject" db:"subject"`
	BodyHTML  string `json:"body_html" db:"body_html"`
	BodyText  string `json:"body_text" db:"body_text"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
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
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.email_templates (type, subject, body_html, body_text, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (type) DO UPDATE SET 
		subject = EXCLUDED.subject, body_html = EXCLUDED.body_html, body_text = EXCLUDED.body_text, updated_at = NOW()
	`, tmpl.Type, tmpl.Subject, tmpl.BodyHTML, tmpl.BodyText)
	return err
}
