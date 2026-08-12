#!/usr/bin/env bash
# =============================================================================
# Nexus ERP — Unified Build Script
# Usage:
#   ./scripts/build.sh            # Linux binary
#   ./scripts/build.sh --windows  # Windows .exe
#   ./scripts/build.sh --all      # Both Linux + Windows
#   ./scripts/build.sh --zip      # Build + package zip
# =============================================================================

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
WEB_DIR="$PROJECT_ROOT/web"
DIST_DIR="$WEB_DIR/dist"

VERSION="${VERSION:-1.0.0}"
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null || echo "unknown")

log_info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
log_success() { echo -e "${GREEN}[OK]${NC}    $*"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ─── Parse Arguments ─────────────────────────────────────────────────────────
BUILD_WINDOWS=false
BUILD_LINUX=true
BUILD_ZIP=false

for arg in "$@"; do
    case "$arg" in
        --windows) BUILD_WINDOWS=true; BUILD_LINUX=false ;;
        --all)     BUILD_WINDOWS=true; BUILD_LINUX=true ;;
        --zip)     BUILD_ZIP=true ;;
        --help|-h)
            echo "Usage: $0 [--windows] [--all] [--zip]"
            exit 0
            ;;
    esac
done

# ─── Step 1: Build Frontend ───────────────────────────────────────────────────
log_info "Step 1/3 — Building Vue 3 frontend..."
if [ ! -d "$WEB_DIR/node_modules" ]; then
    log_info "Installing npm dependencies..."
    cd "$WEB_DIR" && npm install --prefer-offline 2>&1 || \
    (log_warn "offline install failed, trying online..." && npm install)
else
    log_info "node_modules exists, skipping install (use 'npm install' to update)"
fi

cd "$WEB_DIR"
log_info "Running Vite build..."
npm run build

if [ ! -d "$DIST_DIR" ] || [ -z "$(ls -A "$DIST_DIR" 2>/dev/null)" ]; then
    log_error "Frontend build failed: $DIST_DIR is empty or missing"
fi
log_success "Frontend built → $DIST_DIR"
ls -lh "$DIST_DIR"

# ─── Step 2: Go Build ─────────────────────────────────────────────────────────
log_info "Step 2/3 — Building Go backend..."
cd "$PROJECT_ROOT"

log_info "Running go mod tidy..."
go mod tidy

LD_FLAGS="-s -w \
  -X main.Version=${VERSION} \
  -X main.BuildDate=${BUILD_DATE} \
  -X main.GitCommit=${GIT_COMMIT}"

if [ "$BUILD_LINUX" = true ]; then
    log_info "Compiling Linux binary..."
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
        go build -ldflags "$LD_FLAGS" -o nexus-erp .
    chmod +x nexus-erp
    LINUX_SIZE=$(du -sh nexus-erp | cut -f1)
    log_success "Linux binary: nexus-erp (${LINUX_SIZE})"
fi

if [ "$BUILD_WINDOWS" = true ]; then
    log_info "Compiling Windows binary..."
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
        go build -ldflags "$LD_FLAGS" -o nexus-erp.exe .
    WIN_SIZE=$(du -sh nexus-erp.exe | cut -f1)
    log_success "Windows binary: nexus-erp.exe (${WIN_SIZE})"
fi

# ─── Step 3: Package ZIP ──────────────────────────────────────────────────────
if [ "$BUILD_ZIP" = true ]; then
    log_info "Step 3/3 — Packaging distribution ZIP..."
    cd "$PROJECT_ROOT"

    ZIP_NAME="nexus-erp-v${VERSION}-$(date +%Y%m%d)"

    # Always package whatever binaries were built
    FILES_TO_ZIP=()
    [ -f "nexus-erp" ]     && FILES_TO_ZIP+=("nexus-erp")
    [ -f "nexus-erp.exe" ] && FILES_TO_ZIP+=("nexus-erp.exe")
    FILES_TO_ZIP+=(".env.example" "docker-compose.yml" "README.md")

    if [ ${#FILES_TO_ZIP[@]} -eq 0 ]; then
        log_error "No binaries found to package. Run build first."
    fi

    ZIP_FILE="${ZIP_NAME}.zip"
    zip -j "$ZIP_FILE" "${FILES_TO_ZIP[@]}"
    ZIP_SIZE=$(du -sh "$ZIP_FILE" | cut -f1)
    log_success "Package created: ${ZIP_FILE} (${ZIP_SIZE})"
    echo ""
    echo -e "${GREEN}Contents:${NC}"
    unzip -l "$ZIP_FILE"
fi

echo ""
log_success "====== BUILD COMPLETE ======"
echo ""
echo "To run Nexus ERP:"
echo "  1. Set up PostgreSQL and update DATABASE_URL"
echo "  2. Copy .env.example to .env and configure"
echo "  3. ./nexus-erp"
echo "  4. Open http://localhost:8080"
echo "  5. Login: admin / Admin@123456"
echo ""
