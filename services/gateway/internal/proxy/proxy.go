package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Proxy is a reverse proxy that forwards requests to an upstream service
type Proxy struct {
	upstreamURL string
	client      *http.Client
	log         *zap.Logger
}

// NewProxy creates a new reverse proxy pointing to upstreamURL
func NewProxy(upstreamURL string, log *zap.Logger) *Proxy {
	return &Proxy{
		upstreamURL: upstreamURL,
		log:         log,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// Forward proxies the request as-is to the upstream service
func (p *Proxy) Forward(c *fiber.Ctx) error {
	return p.proxy(c, false)
}

// ForwardWithAuth proxies the request and injects OmniBase auth headers for RLS
func (p *Proxy) ForwardWithAuth(c *fiber.Ctx) error {
	return p.proxy(c, true)
}

// ForwardWebSocket upgrades the connection to WebSocket and proxies to upstream
// In Phase 1, this is a simple HTTP upgrade forward.
// Phase 2 will implement a full WS proxy with reconnection logic.
func (p *Proxy) ForwardWebSocket(c *fiber.Ctx) error {
	// TODO: Implement WebSocket proxying with gorilla/websocket
	// For Phase 1 we redirect to the realtime service directly
	return c.Redirect(p.upstreamURL+c.OriginalURL(), fiber.StatusTemporaryRedirect)
}

func (p *Proxy) proxy(c *fiber.Ctx, injectAuth bool) error {
	// Build upstream URL
	upstream, err := url.Parse(p.upstreamURL)
	if err != nil {
		return fmt.Errorf("invalid upstream URL: %w", err)
	}

	// Combine upstream base with incoming path
	targetURL := *upstream
	targetURL.Path = c.Path()
	targetURL.RawQuery = string(c.Request().URI().QueryString())

	// Create upstream request
	req, err := http.NewRequestWithContext(
		c.Context(),
		c.Method(),
		targetURL.String(),
		c.Request().BodyStream(),
	)
	if err != nil {
		return fmt.Errorf("failed to create upstream request: %w", err)
	}

	// Copy request headers
	c.Request().Header.VisitAll(func(key, value []byte) {
		req.Header.Set(string(key), string(value))
	})

	// Set standard proxy headers
	req.Header.Set("X-Forwarded-For", c.IP())
	req.Header.Set("X-Forwarded-Host", c.Hostname())
	req.Header.Set("X-Real-IP", c.IP())

	// Inject OmniBase auth context for RLS
	if injectAuth {
		if userID, ok := c.Locals("user_id").(string); ok && userID != "" {
			req.Header.Set("X-OmniBase-User-ID", userID)
		}
		if role, ok := c.Locals("user_role").(string); ok {
			req.Header.Set("X-OmniBase-Role", role)
			// PostgREST uses this header to set the role
			req.Header.Set("X-PostgREST-Role", role)
		}
	}

	// Execute upstream request
	resp, err := p.client.Do(req)
	if err != nil {
		p.log.Error("upstream request failed",
			zap.String("url", targetURL.String()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"code":    "upstream_error",
			"message": "Upstream service unavailable",
		})
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, v := range values {
			c.Set(key, v)
		}
	}

	// Stream response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read upstream response: %w", err)
	}

	return c.Status(resp.StatusCode).Send(body)
}
