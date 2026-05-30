$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$env:Path = "C:\Program Files\Go\bin;" + $env:Path
$env:GOCACHE = Join-Path $ProjectDir ".gocache"
$env:GOPATH  = Join-Path $ProjectDir ".gopath"
$LogPath     = Join-Path $ProjectDir "agent.combined.log"

Set-Location $ProjectDir

# Auto-create .env on first run
if (-not (Test-Path "$ProjectDir\.env")) {
    Copy-Item "$ProjectDir\.env.example" "$ProjectDir\.env"
    Write-Host "Created .env from .env.example"
    Write-Host "Fill in PRIVATE_KEY and at least one AI key, then re-run."
    Write-Host ""
    Write-Host "  Free: GEMINI_API_KEY  -> https://aistudio.google.com/app/apikey"
    Write-Host "  Free: GROQ_API_KEY    -> https://console.groq.com"
    exit 0
}

# Warn when no AI model key is set
$env = Get-Content "$ProjectDir\.env" -Raw
$hasKey = ($env -match 'GEMINI_API_KEY=\S') -or
          ($env -match 'GROQ_API_KEY=\S')   -or
          ($env -match 'ANTHROPIC_API_KEY=\S') -or
          ($env -match 'OPENAI_API_KEY=\S')
if (-not $hasKey) {
    Write-Host "WARNING: No AI model keys detected."
    Write-Host "  Free: GEMINI_API_KEY  -> https://aistudio.google.com/app/apikey"
    Write-Host "  Free: GROQ_API_KEY    -> https://console.groq.com"
}

"[$(Get-Date -Format o)] Building FusionAI..." | Out-File -LiteralPath $LogPath -Encoding utf8
& "C:\Program Files\Go\bin\go.exe" build . *>&1 | Out-File -LiteralPath $LogPath -Append -Encoding utf8
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed — check $LogPath"
    exit 1
}

"[$(Get-Date -Format o)] Starting FusionAI..." | Out-File -LiteralPath $LogPath -Append -Encoding utf8
& (Join-Path $ProjectDir "fusion-ai.exe") *>&1 | Out-File -LiteralPath $LogPath -Append -Encoding utf8
"[$(Get-Date -Format o)] Agent exited with code $LASTEXITCODE" | Out-File -LiteralPath $LogPath -Append -Encoding utf8
