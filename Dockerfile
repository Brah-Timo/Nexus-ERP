# =============================================================================
# Nexus ERP — Multi-stage Dockerfile
# Stage 1: Build Vue frontend
# Stage 2: Build Go binary with embedded frontend
# Stage 3: Minimal runtime image
# =============================================================================

# ─── Stage 1: Frontend build ─────────────────────────────────────────────────
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

COPY web/package.json web/package-lock.json* ./
RUN npm ci --prefer-offline || npm install

COPY web/ .
RUN npm run build

# ─── Stage 2: Go build ───────────────────────────────────────────────────────
FROM golang:1.21-alpine AS go-builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Download dependencies first (layer cache)
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .
# Copy compiled frontend into expected embed path
COPY --from=frontend-builder /app/web/dist ./web/dist

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o nexus-erp .

# ─── Stage 3: Minimal runtime ─────────────────────────────────────────────────
FROM scratch

# Copy timezone data and CA certs from builder
COPY --from=go-builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=go-builder /app/nexus-erp /nexus-erp

EXPOSE 8080

ENTRYPOINT ["/nexus-erp"]
