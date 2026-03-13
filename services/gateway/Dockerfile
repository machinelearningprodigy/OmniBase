FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY services/gateway/ ./services/gateway/

WORKDIR /app/services/gateway
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/gateway ./cmd/gateway/...

# ── Final Image ──────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget

COPY --from=builder /app/bin/gateway /usr/local/bin/omnibase-gateway

EXPOSE 8000

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget -qO- http://localhost:8000/health || exit 1

ENTRYPOINT ["omnibase-gateway"]
