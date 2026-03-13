.PHONY: dev build test clean deps docker-up docker-down dashboard sdk

# ─── Development ─────────────────────────────────────────────────────────────

## Start the full development stack (requires Docker)
dev:
	docker compose up -d
	@echo "✓ OmniBase stack started"
	@echo "  Dashboard:     http://localhost:3001"
	@echo "  API Gateway:   http://localhost:8000"
	@echo "  MinIO Console: http://localhost:9001"

## Stop the development stack
down:
	docker compose down

## View logs from all services
logs:
	docker compose logs -f

## View logs from a specific service (make logs-gateway)
logs-%:
	docker compose logs -f $*

# ─── Building ─────────────────────────────────────────────────────────────────

## Build all Go services
build: build-gateway build-auth build-storage build-realtime

build-gateway:
	@echo "Building gateway..."
	cd services/gateway && go build -o ../../bin/omnibase-gateway ./cmd/gateway/...

build-auth:
	@echo "Building auth service..."
	cd services/auth && go build -o ../../bin/omnibase-auth ./cmd/auth/...

build-storage:
	@echo "Building storage service..."
	cd services/storage && go build -o ../../bin/omnibase-storage ./cmd/storage/...

build-realtime:
	@echo "Building realtime service..."
	cd services/realtime && go build -o ../../bin/omnibase-realtime ./cmd/realtime/...

# ─── Dependencies ─────────────────────────────────────────────────────────────

## Download all Go dependencies
deps:
	cd shared && go mod tidy
	cd services/gateway && go mod tidy
	cd services/auth && go mod tidy
	cd services/storage && go mod tidy
	cd services/realtime && go mod tidy
	cd cli && go mod tidy 2>/dev/null || true
	@echo "✓ All Go dependencies downloaded"

## Install dashboard dependencies
dashboard-deps:
	cd dashboard && npm install

## Install SDK dependencies  
sdk-deps:
	cd sdk/js && npm install

# ─── Testing ─────────────────────────────────────────────────────────────────

## Run all tests
test: test-go test-sdk

test-go:
	go test ./...

test-sdk:
	cd sdk/js && npm test

# ─── Code Quality ─────────────────────────────────────────────────────────────

## Run linter on all Go code
lint:
	golangci-lint run ./...

## Format all Go code
fmt:
	gofmt -w -s .

# ─── Single Binary Mode ───────────────────────────────────────────────────────

## Build the single-binary OmniBase (SQLite mode, no Docker needed)
single-binary:
	go build -tags sqlite -o bin/omnibase ./cmd/omnibase/...
	@echo "✓ Single binary built at bin/omnibase"
	@echo "  Run: ./bin/omnibase serve"

# ─── Dashboard ────────────────────────────────────────────────────────────────

## Start the dashboard in dev mode
dashboard:
	cd dashboard && npm run dev -- --port 3001

## Build the dashboard for production
dashboard-build:
	cd dashboard && npm run build

# ─── SDK ──────────────────────────────────────────────────────────────────────

## Build the JavaScript SDK
sdk:
	cd sdk/js && npm run build

# ─── Cleanup ─────────────────────────────────────────────────────────────────

## Remove build artifacts
clean:
	rm -rf bin/
	rm -rf dashboard/build/
	rm -rf sdk/js/dist/
	docker compose down -v 2>/dev/null || true

# ─── Utilities ────────────────────────────────────────────────────────────────

## Copy .env.example to .env (won't overwrite existing)
env:
	@if [ ! -f .env ]; then cp .env.example .env && echo "✓ Created .env from .env.example"; else echo ".env already exists, skipping"; fi

## Generate JWT keys for a new installation
gen-keys:
	@echo "Generating OmniBase JWT keys..."
	@echo "JWT_SECRET=$$(openssl rand -base64 32)"
	@echo "ANON_KEY and SERVICE_ROLE_KEY: Run 'omnibase gen keys' after install"

## Show help
help:
	@grep -E '^##' Makefile | sed 's/## //'
