package identity

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListIdentityProviders(c *fiber.Ctx) error {
	projectID := c.Get("X-OmniBase-Project-ID")
	providers, err := h.svc.GetIdentityProviders(c.Context(), projectID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(providers)
}

func (h *Handler) SaveIdentityProvider(c *fiber.Ctx) error {
	var body IdentityProvider
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}
	body.ProjectID = c.Get("X-OmniBase-Project-ID")
	if err := h.svc.SaveIdentityProvider(c.Context(), body); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) DeleteIdentityProvider(c *fiber.Ctx) error {
	projectID := c.Get("X-OmniBase-Project-ID")
	if err := h.svc.DeleteIdentityProvider(c.Context(), c.Params("id"), projectID); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// GetSAMLAuthorize initiates an actual SSO Request Flow
func (h *Handler) GetSAMLAuthorize(c *fiber.Ctx) error {
	providerName := c.Query("provider")
	projectID := c.Query("project_id")
	if providerName == "" || projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing provider or project_id"})
	}

	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
	redirectURL, err := h.svc.GenerateSAMLRequest(c.Context(), providerName, projectID, baseURL)
	if err != nil {
		return h.HandleAuthError(c, err)
	}

	return c.Redirect(redirectURL, fiber.StatusFound)
}

// PostSAMLAssertion catches the Identity Assertion Binding back to OmniBase
func (h *Handler) PostSAMLAssertion(c *fiber.Ctx) error {
	projectID := c.Params("project_id")
	providerName := c.Params("provider")
	samlResponse := c.FormValue("SAMLResponse")

	if samlResponse == "" || projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing SAMLResponse or project_id"})
	}

	resp, err := h.svc.ProcessSAMLResponse(c.Context(), providerName, projectID, samlResponse)
	if err != nil {
		return h.HandleAuthError(c, err)
	}

	h.LogAction(c, resp.User.ID, "saml_sso_login", map[string]any{"provider": providerName, "project_id": projectID})
	return c.JSON(resp)
}
