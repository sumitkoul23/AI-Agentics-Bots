.PHONY: build android android-arm linux darwin darwin-intel windows all clean run help

build: ## Build for current platform
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bodhi-hub .

run: build ## Build and run
	./bodhi-hub

# ── Android ──────────────────────────────────────────────────────────────────
android: ## ARM64 — modern Android phones via Termux
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-android-arm64 .
	@echo "✅  dist/bodhi-hub-android-arm64"

android-arm: ## ARM 32-bit — older Android devices
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-android-arm32 .
	@echo "✅  dist/bodhi-hub-android-arm32"

# ── Desktop ───────────────────────────────────────────────────────────────────
linux: ## Linux x86_64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-linux-amd64 .

linux-arm: ## Linux ARM64 (Raspberry Pi, ARM server)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-linux-arm64 .

darwin: ## macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-macos-arm64 .

darwin-intel: ## macOS Intel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-macos-amd64 .

windows: ## Windows 64-bit
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-hub-windows-amd64.exe .

# ── All platforms ─────────────────────────────────────────────────────────────
all: ## Build all targets into dist/
	mkdir -p dist
	$(MAKE) linux linux-arm darwin darwin-intel windows android android-arm
	@echo ""
	@echo "Built $(shell ls dist/ | wc -l | tr -d ' ') binaries in dist/"
	@ls -lh dist/

clean: ## Remove build artifacts
	rm -f bodhi-hub dist/bodhi-hub-* dist/*.exe

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n",$$1,$$2}'
