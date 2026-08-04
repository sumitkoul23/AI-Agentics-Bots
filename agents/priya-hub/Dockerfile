# ── SkyAgents-hub (Bodhi) — multi-stage Docker build ──────────────────────────
# Stage 1: build the Go binary
FROM golang:1.23-alpine AS builder

WORKDIR /src

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bodhi-hub .

# ── Stage 2: minimal runtime image ────────────────────────────────────────────
FROM alpine:3.20

# cloudflared (IP tunnel) — installed so TUNNEL=cloudflared works out of the box
RUN apk add --no-cache ca-certificates curl && \
    ARCH=$(uname -m) && \
    if [ "$ARCH" = "aarch64" ]; then CF_ARCH=arm64; else CF_ARCH=amd64; fi && \
    curl -fsSL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CF_ARCH}" \
         -o /usr/local/bin/cloudflared && \
    chmod +x /usr/local/bin/cloudflared

WORKDIR /app

COPY --from=builder /bodhi-hub ./bodhi-hub

# /training is mounted from the host (see docker-compose.yml)
# /data    is a named volume for persisted agent memory
RUN mkdir -p /training /data

EXPOSE 8080

ENV PORT=8080 \
    TUNNEL=cloudflared \
    OLLAMA_HOST=http://ollama:11434 \
    OLLAMA_MODEL= \
    NLP_TRAIN_DIR=/training \
    NLP_TRAIN_ENDPOINT=1 \
    REASONING_MODE=chain-of-thought \
    CLI=

ENTRYPOINT ["/app/bodhi-hub"]
