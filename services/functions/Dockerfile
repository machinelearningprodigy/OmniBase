FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY shared/ ./shared/
COPY services/functions/ ./services/functions/

WORKDIR /app/services/functions
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/functions ./cmd/functions/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates nodejs python3 py3-pip
COPY --from=builder /app/bin/functions /usr/local/bin/omnibase-functions
EXPOSE 9006
HEALTHCHECK --interval=10s --timeout=5s CMD wget -qO- http://localhost:9006/health || exit 1
ENTRYPOINT ["omnibase-functions"]
