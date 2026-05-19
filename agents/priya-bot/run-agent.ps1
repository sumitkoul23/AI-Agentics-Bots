# run-agent.ps1 — Launch Priya on Windows / PowerShell
Set-Location $PSScriptRoot

if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env — add your PRIVATE_KEY and ANTHROPIC_API_KEY before running." -ForegroundColor Yellow
    exit 1
}

Write-Host "Fetching dependencies..." -ForegroundColor Cyan
go mod tidy

Write-Host "Building Priya..." -ForegroundColor Cyan
go build -o priya-bot.exe .

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed." -ForegroundColor Red
    exit 1
}

Write-Host "Priya is starting..." -ForegroundColor Green
.\priya-bot.exe
