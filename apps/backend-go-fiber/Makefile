BINARY_NAME := hometab
BIN_DIR := bin
FRONTEND_DIR := ../frontend-react-vite
EMBED_DIR := internal/frontend/dist

VERSION ?= 1.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
LDFLAGS := -s -w -X 'hometab/pkg/buildinfo.Version=$(VERSION)' -X 'hometab/pkg/buildinfo.BuildTime=$(BUILD_TIME)'

.PHONY: all build dev clean frontend embed

all: build

# Full production build: frontend → embed → Go binary → bin/
build: frontend embed
	@mkdir -p $(BIN_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server/
	@echo "✅ Built $(BIN_DIR)/$(BINARY_NAME) ($$(du -h $(BIN_DIR)/$(BINARY_NAME) | cut -f1))"
	@echo "   Version: $(VERSION) | Built: $(BUILD_TIME)"

# Build frontend dist
frontend:
	@echo "📦 Building frontend..."
	cd $(FRONTEND_DIR) && npm ci --prefer-offline --no-audit && npm run build

# Copy frontend dist into embed directory
embed:
	@echo "📋 Embedding frontend dist..."
	rm -rf $(EMBED_DIR)
	mkdir -p $(EMBED_DIR)
	cp -R $(FRONTEND_DIR)/dist/. $(EMBED_DIR)/

# Dev mode: just build Go binary (no frontend embed)
dev:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/server/

# Run in dev mode with CORS enabled
run-dev:
	HOME_CORS_ENABLED=true go run ./cmd/server/ --no-open

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(EMBED_DIR)
	mkdir -p $(EMBED_DIR)
	touch $(EMBED_DIR)/.gitkeep
