package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
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

func (p *Proxy) ForwardWebSocket(c *fiber.Ctx) error {
	upstream, err := url.Parse(p.upstreamURL)
	if err != nil {
		return fmt.Errorf("invalid upstream URL: %w", err)
	}

	requestPath := c.Path()
	rawQuery := string(c.Request().URI().QueryString())
	authHeader := c.Get("Authorization")
	forwardedFor := c.IP()
	forwardedHost := c.Hostname()
	forwardHeaders := http.Header{}
	c.Request().Header.VisitAll(func(key, value []byte) {
		headerKey := string(key)
		if strings.EqualFold(headerKey, "Host") {
			return
		}
		forwardHeaders.Set(headerKey, string(value))
	})

	targetScheme := "ws"
	if upstream.Scheme == "https" {
		targetScheme = "wss"
	}
	targetURL := url.URL{
		Scheme:   targetScheme,
		Host:     upstream.Host,
		Path:     requestPath,
		RawQuery: rawQuery,
	}

	query := targetURL.Query()
	if query.Get("apikey") == "" && strings.HasPrefix(authHeader, "Bearer ") {
		query.Set("apikey", strings.TrimPrefix(authHeader, "Bearer "))
	}
	targetURL.RawQuery = query.Encode()

	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(_ *fasthttp.RequestCtx) bool { return true },
	}

	return upgrader.Upgrade(c.Context(), func(clientConn *websocket.Conn) {
		headers := forwardHeaders.Clone()
		headers.Set("X-Forwarded-For", forwardedFor)
		headers.Set("X-Forwarded-Host", forwardedHost)
		headers.Set("X-Real-IP", forwardedFor)

		upstreamConn, _, dialErr := websocket.DefaultDialer.Dial(targetURL.String(), headers)
		if dialErr != nil {
			p.log.Error("websocket upstream dial failed", zap.String("url", targetURL.String()), zap.Error(dialErr))
			_ = clientConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "upstream unavailable"))
			_ = clientConn.Close()
			return
		}
		defer upstreamConn.Close()
		defer clientConn.Close()

		done := make(chan struct{}, 2)

		go p.pipeWebSocket(clientConn, upstreamConn, done, "client_to_upstream")
		go p.pipeWebSocket(upstreamConn, clientConn, done, "upstream_to_client")

		<-done
	})
}

func (p *Proxy) pipeWebSocket(src, dst *websocket.Conn, done chan<- struct{}, direction string) {
	defer func() { done <- struct{}{} }()
	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			p.log.Debug("websocket proxy read closed", zap.String("direction", direction), zap.Error(err))
			_ = dst.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
		if err := dst.WriteMessage(messageType, message); err != nil {
			p.log.Debug("websocket proxy write closed", zap.String("direction", direction), zap.Error(err))
			return
		}
	}
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
		bytes.NewReader(c.Request().Body()),
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
