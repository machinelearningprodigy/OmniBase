FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY shared/ ./shared/
COPY services/storage/ ./services/storage/

WORKDIR /app/services/storage
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/storage ./cmd/storage/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app/bin/storage /usr/local/bin/omnibase-storage
EXPOSE 9003
HEALTHCHECK --interval=10s --timeout=5s CMD wget -qO- http://localhost:9003/health || exit 1
ENTRYPOINT ["omnibase-storage"]
