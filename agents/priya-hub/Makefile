.PHONY: build android android-arm linux darwin windows clean run

build: ## Build for current platform
	CGO_ENABLED=0 go build -o priya-hub .

run: build ## Build and run
	./priya-hub

# ── Android ──────────────────────────────────────────────────────────────────
android: ## ARM64 — modern Android phones (Pixel, Samsung S-series, etc.) via Termux
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o priya-hub-android-arm64 .
	@echo ""
	@echo "✅  Built: priya-hub-android-arm64"
	@echo ""
	@echo "── Deploy to Android ──────────────────────────────────────────────"
	@echo "1. Install Termux from F-Droid (https://f-droid.org)"
	@echo "2. adb push priya-hub-android-arm64 /sdcard/Download/"
	@echo "   (or share via any file manager)"
	@echo "3. In Termux:"
	@echo "   cp /sdcard/Download/priya-hub-android-arm64 ~/priya-hub"
	@echo "   chmod +x ~/priya-hub"
	@echo "   ./priya-hub"
	@echo "4. Open Chrome → http://localhost:8080"
	@echo "5. Tap ⋮ → 'Add to Home screen' to install as PWA"
	@echo "──────────────────────────────────────────────────────────────────"

android-arm: ## ARM 32-bit — older Android devices
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -o priya-hub-android-arm .
	@echo "✅  Built: priya-hub-android-arm (32-bit)"

# ── Desktop platforms ─────────────────────────────────────────────────────────
linux: ## Linux AMD64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o priya-hub-linux .

darwin: ## macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o priya-hub-macos .

darwin-intel: ## macOS Intel
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o priya-hub-macos-intel .

windows: ## Windows AMD64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o priya-hub.exe .

# ── All platforms ─────────────────────────────────────────────────────────────
all: ## Build all targets
	$(MAKE) build android android-arm linux darwin windows

clean: ## Remove all build artifacts
	rm -f priya-hub priya-hub-* *.exe

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n",$$1,$$2}'
