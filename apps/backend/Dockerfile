FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache deps separately for faster rebuilds
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ─── Runtime image ─────────────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/server ./server

# Create uploads dir owned by app so the named volume initialises with correct ownership.
RUN mkdir -p /app/uploads && chown -R app:app /app/uploads

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]
