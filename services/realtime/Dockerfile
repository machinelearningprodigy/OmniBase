FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY shared/ ./shared/
COPY services/realtime/ ./services/realtime/

WORKDIR /app/services/realtime
RUN go mod download
RUN go build -ldflags="-w -s" -o /app/bin/realtime ./cmd/realtime/...

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=builder /app/bin/realtime /usr/local/bin/omnibase-realtime
EXPOSE 9004
ENTRYPOINT ["omnibase-realtime"]
