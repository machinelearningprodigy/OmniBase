package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
	"go.uber.org/zap"
)

type FunctionRecord struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Slug      string                 `json:"slug"`
	Runtime   string                 `json:"runtime"`
	Source    map[string]interface{} `json:"source"`
	VerifyJWT bool                   `json:"verify_jwt"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type Service struct {
	db     *pgxpool.Pool
	jwt    *jwt.Manager
	log    *zap.Logger
	client *http.Client
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.MustNew(cfg.LogLevel, cfg.Env)
	defer log.Sync()

	db, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer db.Close()

	svc := &Service{
		db:     db,
		jwt:    jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry),
		log:    log,
		client: &http.Client{Timeout: 30 * time.Second},
	}
	if err := svc.ensureTables(context.Background()); err != nil {
		log.Fatal("failed to initialize functions schema", zap.Error(err))
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "functions"})
	})

	api := app.Group("/functions/v1")
	api.Get("/", svc.ListFunctions)
	api.Get("/:slug", svc.GetFunction)
	api.Post("/", svc.RequireAdmin, svc.UpsertFunction)
	api.Delete("/:slug", svc.RequireAdmin, svc.DeleteFunction)
	api.All("/:slug/invoke", svc.InvokeFunction)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down functions service...")
		_ = app.Shutdown()
	}()

	port := fmt.Sprintf("%d", cfg.FunctionsPort)
	log.Info("Functions service starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Functions service stopped", zap.Error(err))
	}
}

func (s *Service) ensureTables(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS omnibase;
		CREATE TABLE IF NOT EXISTS omnibase.functions (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			runtime TEXT NOT NULL,
			source JSONB NOT NULL DEFAULT '{}'::jsonb,
			verify_jwt BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *Service) RequireAdmin(c *fiber.Ctx) error {
	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
	}
	claims, err := s.jwt.Verify(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	if claims.Role == "service_role" {
		return c.Next()
	}

	var isAdmin bool
	if err := s.db.QueryRow(c.Context(), "SELECT is_super_admin FROM auth.users WHERE id = $1", claims.UserID).Scan(&isAdmin); err != nil || !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}
	return c.Next()
}

func (s *Service) ListFunctions(c *fiber.Ctx) error {
	rows, err := s.db.Query(c.Context(), `
		SELECT id::text, name, slug, runtime, source, verify_jwt, created_at::text, updated_at::text
		FROM omnibase.functions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	items := make([]FunctionRecord, 0)
	for rows.Next() {
		var item FunctionRecord
		var sourceBytes []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Runtime, &sourceBytes, &item.VerifyJWT, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		_ = json.Unmarshal(sourceBytes, &item.Source)
		items = append(items, item)
	}
	return c.JSON(items)
}

func (s *Service) GetFunction(c *fiber.Ctx) error {
	item, err := s.loadFunction(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Function not found"})
	}
	return c.JSON(item)
}

func (s *Service) UpsertFunction(c *fiber.Ctx) error {
	var body struct {
		Name      string                 `json:"name"`
		Slug      string                 `json:"slug"`
		Runtime   string                 `json:"runtime"`
		Source    map[string]interface{} `json:"source"`
		VerifyJWT *bool                  `json:"verify_jwt"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if body.Name == "" || body.Slug == "" || body.Runtime == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name, slug, and runtime are required"})
	}
	if !isSupportedRuntime(body.Runtime) {
		return c.Status(400).JSON(fiber.Map{"error": "supported runtimes: static-json, webhook, javascript, python"})
	}

	verifyJWT := true
	if body.VerifyJWT != nil {
		verifyJWT = *body.VerifyJWT
	}
	sourceBytes, _ := json.Marshal(body.Source)

	_, err := s.db.Exec(c.Context(), `
		INSERT INTO omnibase.functions (id, name, slug, runtime, source, verify_jwt, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			runtime = EXCLUDED.runtime,
			source = EXCLUDED.source,
			verify_jwt = EXCLUDED.verify_jwt,
			updated_at = NOW()
	`, uuid.New().String(), body.Name, body.Slug, body.Runtime, sourceBytes, verifyJWT)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	item, err := s.loadFunction(c.Context(), body.Slug)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (s *Service) DeleteFunction(c *fiber.Ctx) error {
	tag, err := s.db.Exec(c.Context(), `DELETE FROM omnibase.functions WHERE slug = $1`, c.Params("slug"))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Function not found"})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (s *Service) InvokeFunction(c *fiber.Ctx) error {
	item, err := s.loadFunction(c.Context(), c.Params("slug"))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Function not found"})
	}

	if item.VerifyJWT {
		auth := c.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header required"})
		}
		if _, err := s.jwt.Verify(strings.TrimPrefix(auth, "Bearer ")); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}
	}

	switch item.Runtime {
	case "static-json":
		return serveStaticJSON(c, item)
	case "webhook":
		return s.invokeWebhook(c, item)
	case "javascript", "python":
		return s.invokeProcessRuntime(c, item)
	default:
		return c.Status(500).JSON(fiber.Map{"error": "unsupported runtime"})
	}
}

func serveStaticJSON(c *fiber.Ctx, item *FunctionRecord) error {
	statusCode := fiber.StatusOK
	if rawStatus, ok := item.Source["status"].(float64); ok {
		statusCode = int(rawStatus)
	}
	if headers, ok := item.Source["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			c.Set(key, fmt.Sprint(value))
		}
	}
	if body, ok := item.Source["body"]; ok {
		return c.Status(statusCode).JSON(body)
	}
	return c.Status(statusCode).JSON(item.Source)
}

func (s *Service) invokeWebhook(c *fiber.Ctx, item *FunctionRecord) error {
	targetURL, _ := item.Source["url"].(string)
	if targetURL == "" {
		return c.Status(500).JSON(fiber.Map{"error": "webhook url is not configured"})
	}
	method := c.Method()
	if configuredMethod, ok := item.Source["method"].(string); ok && configuredMethod != "" {
		method = strings.ToUpper(configuredMethod)
	}

	req, err := http.NewRequestWithContext(c.Context(), method, targetURL, bytes.NewReader(c.Body()))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	c.Request().Header.VisitAll(func(k, v []byte) {
		key := string(k)
		if strings.EqualFold(key, "Host") || strings.EqualFold(key, "Content-Length") {
			return
		}
		req.Header.Set(key, string(v))
	})
	resp, err := s.client.Do(req)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Append(key, value)
		}
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(502).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(resp.StatusCode).Send(payload)
}

func (s *Service) invokeProcessRuntime(c *fiber.Ctx, item *FunctionRecord) error {
	workdir, err := os.MkdirTemp("", "omnibase-fn-*")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer os.RemoveAll(workdir)

	payload := map[string]interface{}{
		"method":  c.Method(),
		"path":    c.Path(),
		"query":   c.Queries(),
		"headers": flattenHeaders(c),
		"body":    parseBody(c),
	}
	payloadBytes, _ := json.Marshal(payload)

	var cmd *exec.Cmd
	switch item.Runtime {
	case "javascript":
		cmd, err = prepareJavaScriptCommand(workdir, item.Source)
	case "python":
		cmd, err = preparePythonCommand(workdir, item.Source)
	}
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	timeout := 15 * time.Second
	if rawTimeout, ok := item.Source["timeout_seconds"].(float64); ok && rawTimeout > 0 && rawTimeout <= 120 {
		timeout = time.Duration(rawTimeout) * time.Second
	}
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()
	cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
	cmd.Dir = workdir
	cmd.Stdin = bytes.NewReader(payloadBytes)
	cmd.Env = buildFunctionEnv(item.Source)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "function execution failed",
			"details": strings.TrimSpace(stderr.String()),
		})
	}

	if stderr.Len() > 0 {
		c.Set("X-OmniBase-Function-Stderr", truncate(stderr.String(), 256))
	}

	resultBytes := bytes.TrimSpace(stdout.Bytes())
	if len(resultBytes) == 0 {
		return c.SendStatus(fiber.StatusNoContent)
	}

	var result interface{}
	if json.Unmarshal(resultBytes, &result) == nil {
		return c.JSON(result)
	}
	return c.Send(resultBytes)
}

func prepareJavaScriptCommand(workdir string, source map[string]interface{}) (*exec.Cmd, error) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return nil, fmt.Errorf("node runtime not found in PATH")
	}
	code, _ := source["code"].(string)
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("javascript source.code is required")
	}
	handler := "handler"
	if rawHandler, ok := source["handler"].(string); ok && rawHandler != "" {
		handler = rawHandler
	}

	script := fmt.Sprintf(`%s
const payload = JSON.parse(require("fs").readFileSync(0, "utf8") || "{}");
(async () => {
  const fn = globalThis[%q];
  if (typeof fn !== "function") {
    throw new Error("handler function %s not found");
  }
  const result = await fn(payload);
  process.stdout.write(JSON.stringify(result ?? null));
})().catch((err) => {
  console.error(err && err.stack ? err.stack : String(err));
  process.exit(1);
});
`, code, handler, handler)
	file := filepath.Join(workdir, "index.cjs")
	if err := os.WriteFile(file, []byte(script), 0o600); err != nil {
		return nil, err
	}
	return exec.Command(nodePath, file), nil
}

func preparePythonCommand(workdir string, source map[string]interface{}) (*exec.Cmd, error) {
	pythonPath, err := exec.LookPath("python")
	if err != nil {
		return nil, fmt.Errorf("python runtime not found in PATH")
	}
	code, _ := source["code"].(string)
	if strings.TrimSpace(code) == "" {
		return nil, fmt.Errorf("python source.code is required")
	}
	handler := "handler"
	if rawHandler, ok := source["handler"].(string); ok && rawHandler != "" {
		handler = rawHandler
	}

	script := fmt.Sprintf(`import json, sys
%s
payload = json.loads(sys.stdin.read() or "{}")
fn = globals().get(%q)
if not callable(fn):
    raise RuntimeError("handler function %s not found")
result = fn(payload)
sys.stdout.write(json.dumps(result))
`, code, handler, handler)
	file := filepath.Join(workdir, "main.py")
	if err := os.WriteFile(file, []byte(script), 0o600); err != nil {
		return nil, err
	}
	return exec.Command(pythonPath, file), nil
}

func buildFunctionEnv(source map[string]interface{}) []string {
	env := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"TEMP=" + os.Getenv("TEMP"),
		"TMP=" + os.Getenv("TMP"),
		"OMNIBASE_FUNCTION=true",
	}
	if rawEnv, ok := source["env"].(map[string]interface{}); ok {
		for key, value := range rawEnv {
			env = append(env, key+"="+fmt.Sprint(value))
		}
	}
	return env
}

func flattenHeaders(c *fiber.Ctx) map[string]string {
	headers := map[string]string{}
	c.Request().Header.VisitAll(func(k, v []byte) {
		headers[string(k)] = string(v)
	})
	return headers
}

func parseBody(c *fiber.Ctx) interface{} {
	if len(c.Body()) == 0 {
		return nil
	}
	var payload interface{}
	if json.Unmarshal(c.Body(), &payload) == nil {
		return payload
	}
	return string(c.Body())
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func isSupportedRuntime(runtime string) bool {
	switch runtime {
	case "static-json", "webhook", "javascript", "python":
		return true
	default:
		return false
	}
}

func (s *Service) loadFunction(ctx context.Context, slug string) (*FunctionRecord, error) {
	var item FunctionRecord
	var sourceBytes []byte
	err := s.db.QueryRow(ctx, `
		SELECT id::text, name, slug, runtime, source, verify_jwt, created_at::text, updated_at::text
		FROM omnibase.functions
		WHERE slug = $1
	`, slug).Scan(&item.ID, &item.Name, &item.Slug, &item.Runtime, &sourceBytes, &item.VerifyJWT, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(sourceBytes, &item.Source)
	return &item, nil
}
