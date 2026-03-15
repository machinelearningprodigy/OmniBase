package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"go.uber.org/zap"
)

const (
	userIDKey    = "user_id"
	userEmailKey = "user_email"
	userRoleKey  = "user_role"
	projectIDKey = "project_id"
	claimsKey    = "jwt_claims"
)

// InjectUserContext extracts JWT from Authorization header and injects user info
// into the request context. Does NOT reject unauthenticated requests — anon access
// is controlled by PostgREST/RLS policies.
func InjectUserContext(jm *jwt.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			// Anonymous request — set anon role
			c.Locals(userRoleKey, "anon")
			c.Set("X-OmniBase-Role", "anon")
			return c.Next()
		}

		claims, err := jm.Verify(token)
		if err != nil {
			// Invalid token — still allow as anon (RLS will restrict)
			c.Locals(userRoleKey, "anon")
			c.Set("X-OmniBase-Role", "anon")
			return c.Next()
		}

		c.Locals(userIDKey, claims.UserID)
		c.Locals(userEmailKey, claims.Email)
		c.Locals(userRoleKey, claims.Role)
		c.Locals(projectIDKey, claims.ProjectID)
		c.Locals(claimsKey, claims)

		// Inject user info as headers for internal services
		c.Set("X-OmniBase-User-ID", claims.UserID)
		c.Set("X-OmniBase-Role", claims.Role)
		c.Set("X-OmniBase-Project-ID", claims.ProjectID)

		return c.Next()
	}
}

// RequireAuth rejects requests without a valid JWT
func RequireAuth(jm *jwt.Manager, log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "missing_token",
				"message": "Authorization header required",
			})
		}

		claims, err := jm.Verify(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"code":    "invalid_token",
				"message": err.Error(),
			})
		}

		c.Locals(userIDKey, claims.UserID)
		c.Locals(userRoleKey, claims.Role)
		c.Locals(claimsKey, claims)
		return c.Next()
	}
}

// RequireServiceRole allows only requests with service_role JWT
func RequireServiceRole(jm *jwt.Manager, log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    "forbidden",
				"message": "Service role key required",
			})
		}

		claims, err := jm.Verify(token)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    "forbidden",
				"message": "Invalid service role key",
			})
		}

		if claims.Role != "service_role" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"code":    "forbidden",
				"message": "Service role required for this endpoint",
			})
		}

		c.Locals(claimsKey, claims)
		return c.Next()
	}
}

// extractToken gets the bearer token from Authorization header or apikey query param
func extractToken(c *fiber.Ctx) string {
	// Check Authorization: Bearer <token>
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check X-OmniBase-Key header (anon key or service role key)
	if key := c.Get("X-OmniBase-Key"); key != "" {
		return key
	}

	// Check apikey query param (PostgREST compatibility)
	if key := c.Query("apikey"); key != "" {
		return key
	}

	return ""
}
