.PHONY: build android android-arm linux linux-arm darwin darwin-intel windows all run clean help

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bodhi-app .

run: build
	./bodhi-app

android:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-android-arm64 .
	@echo "✅  dist/bodhi-app-android-arm64"

android-arm:
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-android-arm32 .

linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-linux-amd64 .

linux-arm:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-linux-arm64 .

darwin:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-macos-arm64 .

darwin-intel:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-macos-amd64 .

windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/bodhi-app-windows-amd64.exe .

all:
	mkdir -p dist
	$(MAKE) linux linux-arm darwin darwin-intel windows android android-arm
	@ls -lh dist/

clean:
	rm -f bodhi-app dist/bodhi-app-* dist/*.exe

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  %-16s %s\n",$$1,$$2}'
