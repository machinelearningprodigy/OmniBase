package identity

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
	"go.uber.org/zap"
)

func (s *Service) GenerateSAMLRequest(ctx context.Context, providerName string, projectID string, baseURL string) (string, error) {
	provider, err := s.GetIdentityProviderByName(ctx, providerName, projectID)
	if err != nil || provider.Type != "saml" || !provider.IsActive {
		return "", &AuthError{Code: "provider_error", Message: "SAML provider not found or inactive"}
	}

	metadataURL, err := url.Parse(provider.MetadataURL)
	if err != nil {
		return "", fmt.Errorf("invalid metadata URL: %w", err)
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *metadataURL)
	if err != nil {
		return "", fmt.Errorf("failed fetching SAML metadata: %w", err)
	}

	// Construct the ACS URL with projectID and providerName
	acsURL, _ := url.Parse(baseURL)
	acsURL.Path = fmt.Sprintf("/auth/v1/saml/acs/%s/%s", projectID, providerName)

	sp := samlsp.Options{
		EntityID:    baseURL + "/saml/metadata", // EntityID can remain static or be made dynamic if needed
		URL:         *acsURL,                    // This URL will be used to construct the ACS endpoint
		IDPMetadata: idpMetadata,
	}

	samlProvider, err := samlsp.New(sp)
	if err != nil {
		return "", fmt.Errorf("failed to configure SAML service provider: %w", err)
	}

	ssoLocation := samlProvider.ServiceProvider.GetSSOBindingLocation(saml.HTTPRedirectBinding)
	if ssoLocation == "" {
		return "", fmt.Errorf("idp does not support HTTP-Redirect binding")
	}

	req, err := samlProvider.ServiceProvider.MakeAuthenticationRequest(ssoLocation, saml.HTTPRedirectBinding, "")
	if err != nil {
		return "", fmt.Errorf("failed making auth req: %w", err)
	}

	u, err := req.Redirect("", &samlProvider.ServiceProvider)
	if err != nil {
		return "", fmt.Errorf("failed creating redirect: %w", err)
	}
	return u.String(), nil
}

func (s *Service) ProcessSAMLResponse(ctx context.Context, providerName string, projectID string, samlResponse string) (*services.SignUpResponse, error) {
	provider, err := s.GetIdentityProviderByName(ctx, providerName, projectID)
	if err != nil {
		return nil, err
	}

	metadataURL, _ := url.Parse(provider.MetadataURL)
	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed fetching SAML metadata: %w", err)
	}

	sp := saml.ServiceProvider{
		EntityID:    "omnibase-saml",
		IDPMetadata: idpMetadata,
	}

	// We need a dummy http.Request to ParseResponse
	req, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("SAMLResponse="+url.QueryEscape(samlResponse)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	assertion, err := sp.ParseResponse(req, []string{})
	if err != nil {
		return nil, &AuthError{Code: "saml_invalid", Message: "SAML Assertion Invalid: " + err.Error()}
	}

	email := extractSAMLEmail(assertion)
	if email == "" {
		return nil, &AuthError{Code: "saml_error", Message: "SAML assertion missing distinct email identity"}
	}

	var userID string
	err = s.db.QueryRow(ctx, "SELECT id::text FROM auth.users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		// Auto Provision user
		signUpReq := SignUpRequest{
			Email: email,
		}
		u, err := s.SignUp(ctx, signUpReq)
		if err != nil {
			return nil, err
		}
		userID = u.User.ID
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.db.Exec(ctx, "UPDATE auth.users SET last_sign_in_at = NOW() WHERE id = $1", user.ID)

	accessToken, refreshToken, err := s.IssueTokenPair(ctx, user, "SAML", "")
	if err != nil {
		return nil, err
	}

	s.log.Info("SAML SSO user authenticated", zap.String("email", email), zap.String("provider", providerName))

	return &SignUpResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    3600,
	}, nil
}

func httptestRequestValidSAML(base64resp string) *http.Request {
	req, _ := http.NewRequest("POST", "/saml/acs", nil)
	req.PostForm = url.Values{}
	req.PostForm.Add("SAMLResponse", base64resp)
	return req
}

func extractSAMLEmail(assertion *saml.Assertion) string {
	email := ""
	for _, attr := range assertion.AttributeStatements {
		for _, a := range attr.Attributes {
			if strings.Contains(strings.ToLower(a.FriendlyName), "email") || strings.Contains(strings.ToLower(a.Name), "emailaddress") {
				if len(a.Values) > 0 {
					email = a.Values[0].Value
				}
			}
		}
	}
	if email == "" && assertion.Subject != nil && assertion.Subject.NameID != nil {
		email = assertion.Subject.NameID.Value
	}
	return email
}

func (s *Service) GetIdentityProviderByName(ctx context.Context, name string, projectID string) (IdentityProvider, error) {
	var p IdentityProvider
	var cid, cs, murl *string
	err := s.db.QueryRow(ctx, `
		SELECT id, project_id, type, name, client_id, client_secret, metadata_url, is_active 
		FROM auth.identity_providers 
		WHERE name = $1 AND project_id = $2
	`, name, projectID).Scan(&p.ID, &p.ProjectID, &p.Type, &p.Name, &cid, &cs, &murl, &p.IsActive)
	if cid != nil {
		p.ClientID = *cid
	}
	if cs != nil {
		p.ClientSecret = *cs
	}
	if murl != nil {
		p.MetadataURL = *murl
	}
	return p, err
}

func (s *Service) GetIdentityProviders(ctx context.Context, projectID string) ([]IdentityProvider, error) {
	rows, err := s.db.Query(ctx, "SELECT id, project_id, type, name, client_id, client_secret, metadata_url, is_active FROM auth.identity_providers WHERE project_id = $1", projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []IdentityProvider
	for rows.Next() {
		var p IdentityProvider
		var cid, cs, murl *string
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Type, &p.Name, &cid, &cs, &murl, &p.IsActive); err != nil {
			return nil, err
		}
		if cid != nil {
			p.ClientID = *cid
		}
		if cs != nil {
			p.ClientSecret = *cs
		}
		if murl != nil {
			p.MetadataURL = *murl
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func (s *Service) SaveIdentityProvider(ctx context.Context, p IdentityProvider) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.identity_providers (project_id, type, name, client_id, client_secret, metadata_url, is_active)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7)
		ON CONFLICT (name, project_id) DO UPDATE SET
		type = EXCLUDED.type, client_id = EXCLUDED.client_id, client_secret = EXCLUDED.client_secret,
		metadata_url = EXCLUDED.metadata_url, is_active = EXCLUDED.is_active, updated_at = NOW()
	`, p.ProjectID, p.Type, p.Name, p.ClientID, p.ClientSecret, p.MetadataURL, p.IsActive)
	return err
}

func (s *Service) DeleteIdentityProvider(ctx context.Context, id string, projectID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.identity_providers WHERE id = $1 AND project_id = $2", id, projectID)
	return err
}
