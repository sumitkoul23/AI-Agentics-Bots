# run-agent.ps1 – Launch the Mobile Assistant Bot on Windows / PowerShell
Set-Location $PSScriptRoot

if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env from .env.example – add your PRIVATE_KEY before running." -ForegroundColor Yellow
    exit 1
}

Write-Host "Building Mobile Assistant Bot..." -ForegroundColor Cyan
go build -o mobile-assistant-bot.exe .

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}

Write-Host "Starting agent..." -ForegroundColor Green
.\mobile-assistant-bot.exe
