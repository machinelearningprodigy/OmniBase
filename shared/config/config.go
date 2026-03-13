package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds ALL configuration for every OmniBase service.
// Each service reads only the fields it needs.
// Future phases add fields here without breaking existing services.
type Config struct {
	// ─── Server ───────────────────────────────────────
	Env      string // "development" | "staging" | "production"
	LogLevel string // "debug" | "info" | "warn" | "error"

	// ─── Gateway ──────────────────────────────────────
	GatewayPort    int
	GatewayHost    string
	RateLimitRPS   int
	AllowedOrigins []string

	// ─── Services (internal URLs) ─────────────────────
	AuthServiceURL     string
	DatabaseServiceURL string
	StorageServiceURL   string
	RealtimeServiceURL  string
	FunctionsServiceURL string

	// ─── Database ─────────────────────────────────────
	DatabaseURL      string // Full postgres DSN
	DatabaseHost     string
	DatabasePort     int
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	DatabaseSSLMode  string
	MaxOpenConns     int
	MaxIdleConns     int
	ConnMaxLifetime  time.Duration

	// ─── Auth ─────────────────────────────────────────
	JWTSecret          string
	JWTAccessExpiry    time.Duration
	JWTRefreshExpiry   time.Duration
	AuthPort           int
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPass           string
	SMTPSenderName     string
	SMTPSenderEmail    string
	SiteURL            string
	APIExternalURL     string

	// OAuth2 Providers (Phase 1: Google + GitHub)
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string

	// ─── Storage ──────────────────────────────────────
	StoragePort       int
	StorageBackend    string // "minio" | "local" | "s3"
	MinIOEndpoint     string
	MinIOAccessKey    string
	MinIOSecretKey    string
	MinIOBucket       string
	MinIOUseSSL       bool
	LocalStoragePath  string
	MaxUploadSizeMB   int64

	// ─── Realtime ─────────────────────────────────────
	RealtimePort         int
	ReplicationSlot      string
	ReplicationPublisher string

	// ─── Functions ────────────────────────────────────
	FunctionsPort int

	// ─── Cache (Valkey/Redis) ─────────────────────────
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// ─── NATS ─────────────────────────────────────────
	NATSUrl string

	// ─── Database Meta ────────────────────────────────
	DatabaseMetaPort int
	PostgRESTURL     string

	// ─── Observability ────────────────────────────────
	OTELEndpoint    string
	OTELServiceName string
}

// Load reads config from environment variables and optional .env files.
// Priority: OS env > .env.local > .env
func Load() (*Config, error) {
	// Load .env files (non-fatal if missing)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := &Config{
		Env:      getEnv("OMNIBASE_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		GatewayPort:  getEnvInt("GATEWAY_PORT", 8000),
		GatewayHost:  getEnv("GATEWAY_HOST", "0.0.0.0"),
		RateLimitRPS: getEnvInt("RATE_LIMIT_RPS", 1000),

		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:9001"),
		DatabaseServiceURL: getEnv("DATABASE_SERVICE_URL", "http://localhost:9002"),
		StorageServiceURL:  getEnv("STORAGE_SERVICE_URL", "http://localhost:9003"),
		RealtimeServiceURL: getEnv("REALTIME_SERVICE_URL", "http://localhost:9004"),
		FunctionsServiceURL: getEnv("FUNCTIONS_SERVICE_URL", "http://localhost:9006"),

		DatabaseURL:      getEnv("DATABASE_URL", ""),
		DatabaseHost:     getEnv("POSTGRES_HOST", "localhost"),
		DatabasePort:     getEnvInt("POSTGRES_PORT", 5432),
		DatabaseUser:     getEnv("POSTGRES_USER", "omnibase"),
		DatabasePassword: getEnv("POSTGRES_PASSWORD", ""),
		DatabaseName:     getEnv("POSTGRES_DB", "omnibase"),
		DatabaseSSLMode:  getEnv("POSTGRES_SSL_MODE", "disable"),
		MaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime:  getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  getEnvDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
		JWTRefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 720*time.Hour),
		AuthPort:         getEnvInt("AUTH_PORT", 9001),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnvInt("SMTP_PORT", 587),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPass:         getEnv("SMTP_PASS", ""),
		SMTPSenderName:   getEnv("SMTP_SENDER_NAME", "OmniBase"),
		SMTPSenderEmail:  getEnv("SMTP_SENDER_EMAIL", "noreply@omnibase.dev"),
		SiteURL:          getEnv("SITE_URL", "http://localhost:3000"),
		APIExternalURL:   getEnv("API_EXTERNAL_URL", "http://localhost:8000"),

		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),

		StoragePort:      getEnvInt("STORAGE_PORT", 9003),
		StorageBackend:   getEnv("STORAGE_BACKEND", "minio"),
		MinIOEndpoint:    getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:   getEnv("MINIO_ACCESS_KEY", "omnibase"),
		MinIOSecretKey:   getEnv("MINIO_SECRET_KEY", "omnibase123"),
		MinIOBucket:      getEnv("MINIO_BUCKET", "omnibase-storage"),
		MinIOUseSSL:      getEnvBool("MINIO_USE_SSL", false),
		LocalStoragePath: getEnv("LOCAL_STORAGE_PATH", "./data/storage"),
		MaxUploadSizeMB:  int64(getEnvInt("MAX_UPLOAD_SIZE_MB", 50)),

		RealtimePort:         getEnvInt("REALTIME_PORT", 9004),
		ReplicationSlot:      getEnv("REPLICATION_SLOT", "omnibase_realtime"),
		ReplicationPublisher: getEnv("REPLICATION_PUBLISHER", "omnibase_publication"),

		FunctionsPort: getEnvInt("FUNCTIONS_PORT", 9006),

		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		NATSUrl: getEnv("NATS_URL", "nats://localhost:4222"),

		DatabaseMetaPort: getEnvInt("DATABASE_META_PORT", 9005),
		PostgRESTURL:     getEnv("POSTGREST_URL", "http://localhost:3000"),

		OTELEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceName: getEnv("OTEL_SERVICE_NAME", "omnibase"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.JWTSecret == "" && c.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be set in production")
	}
	return nil
}

// DSN builds a postgres connection string from individual fields if DATABASE_URL is not set
func (c *Config) DSN() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DatabaseHost, c.DatabasePort, c.DatabaseUser, c.DatabasePassword, c.DatabaseName, c.DatabaseSSLMode)
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultVal
}
