# run-agent.ps1 — Launch FusionAI on Windows / PowerShell
Set-Location $PSScriptRoot

# ── Pre-flight checks ─────────────────────────────────────────────────────────
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Go is not installed. Download from https://go.dev/dl/" -ForegroundColor Red
    exit 1
}

if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host ""
    Write-Host "  Created .env from .env.example" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  Before starting, open .env and fill in:" -ForegroundColor Yellow
    Write-Host "    PRIVATE_KEY      — your Teneo wallet private key" -ForegroundColor Yellow
    Write-Host "    GEMINI_API_KEY   — free at https://aistudio.google.com/app/apikey" -ForegroundColor Yellow
    Write-Host "    GROQ_API_KEY     — free at https://console.groq.com  (optional)" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

# ── Check for at least one AI model key ───────────────────────────────────────
$env_content = Get-Content ".env" -Raw
$hasModel = ($env_content -match "GEMINI_API_KEY=.+") -or
            ($env_content -match "GROQ_API_KEY=.+")   -or
            ($env_content -match "ANTHROPIC_API_KEY=.+") -or
            ($env_content -match "OPENAI_API_KEY=.+")

if (-not $hasModel) {
    Write-Host ""
    Write-Host "  WARNING: No AI model API key set in .env" -ForegroundColor Yellow
    Write-Host "  Add at least one free key:" -ForegroundColor Yellow
    Write-Host "    GEMINI_API_KEY  →  https://aistudio.google.com/app/apikey" -ForegroundColor Cyan
    Write-Host "    GROQ_API_KEY    →  https://console.groq.com" -ForegroundColor Cyan
    Write-Host ""
}

# ── Build ─────────────────────────────────────────────────────────────────────
Write-Host ""
Write-Host "  FusionAI — building..." -ForegroundColor Cyan
go build -o fusion-ai.exe .

if ($LASTEXITCODE -ne 0) {
    Write-Host "  Build failed. Check errors above." -ForegroundColor Red
    exit 1
}

Write-Host "  Build OK" -ForegroundColor Green
Write-Host ""

# ── Launch ────────────────────────────────────────────────────────────────────
Write-Host "  Starting FusionAI on Teneo network..." -ForegroundColor Green
Write-Host "  Models: Gemini (free) · Groq (free) · Claude · GPT-4o · Ollama" -ForegroundColor DarkGray
Write-Host "  Commands: /model /models /code /analyze /write /math /help" -ForegroundColor DarkGray
Write-Host ""
.\fusion-ai.exe
