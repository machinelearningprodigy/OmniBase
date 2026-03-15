FROM golang:1.24-alpine AS builder
ENV GOTOOLCHAIN=auto

WORKDIR /app
COPY shared/ ./shared/
COPY services/auth/ ./services/auth/

WORKDIR /app/services/auth
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/auth ./cmd/auth/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app/bin/auth /usr/local/bin/omnibase-auth
EXPOSE 9001
HEALTHCHECK --interval=10s --timeout=5s CMD wget -qO- http://localhost:9001/health || exit 1
ENTRYPOINT ["omnibase-auth"]
