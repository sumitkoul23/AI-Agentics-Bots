.PHONY: build android android-arm linux darwin windows run clean

build:
	CGO_ENABLED=0 go build -o priya-app .

run: build
	./priya-app

android:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o priya-app-android-arm64 .
	@echo ""
	@echo "✅  priya-app-android-arm64 ready"
	@echo ""
	@echo "── Install on Android ──────────────────────────────────────────────"
	@echo "Requires: priya-hub running on the same device or LAN"
	@echo ""
	@echo "1. Copy binary to phone (adb / Bluetooth / Files)"
	@echo "2. In Termux:"
	@echo "   chmod +x ~/priya-app && PRIYA_HUB_URL=http://localhost:8080 ./priya-app"
	@echo "3. Open Chrome → http://localhost:9090"
	@echo "4. Tap ⋮ → Add to Home Screen → pure app experience"
	@echo "────────────────────────────────────────────────────────────────────"

android-arm:
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -o priya-app-android-arm .

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o priya-app-linux .

darwin:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o priya-app-macos .

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o priya-app.exe .

clean:
	rm -f priya-app priya-app-* *.exe
