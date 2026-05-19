Set-Location $PSScriptRoot
if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "Created .env — add PRIVATE_KEY and ANTHROPIC_API_KEY." -ForegroundColor Yellow
    exit 1
}
Write-Host "Building Priya Hub..." -ForegroundColor Cyan
go mod tidy
go build -o priya-hub.exe .
if ($LASTEXITCODE -ne 0) { Write-Host "Build failed." -ForegroundColor Red; exit 1 }
Write-Host "Priya Hub starting..." -ForegroundColor Green
.\priya-hub.exe
