FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY shared/ ./shared/
COPY services/database/ ./services/database/

WORKDIR /app/services/database
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/database ./cmd/database/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app/bin/database /usr/local/bin/omnibase-database
EXPOSE 9005
HEALTHCHECK --interval=10s --timeout=5s CMD wget -qO- http://localhost:9005/health || exit 1
ENTRYPOINT ["omnibase-database"]
