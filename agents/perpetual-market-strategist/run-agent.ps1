$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$env:Path = "C:\Program Files\Go\bin;" + $env:Path
$env:GOCACHE = Join-Path $ProjectDir ".gocache"
$env:GOPATH = Join-Path $ProjectDir ".gopath"
$LogPath = Join-Path $ProjectDir "agent.combined.log"

Set-Location $ProjectDir
"[$(Get-Date -Format o)] Building agent..." | Out-File -LiteralPath $LogPath -Encoding utf8
& "C:\Program Files\Go\bin\go.exe" build . *>&1 | Out-File -LiteralPath $LogPath -Append -Encoding utf8
"[$(Get-Date -Format o)] Starting agent..." | Out-File -LiteralPath $LogPath -Append -Encoding utf8
& (Join-Path $ProjectDir "perpetual-market-strategist.exe") *>&1 | Out-File -LiteralPath $LogPath -Append -Encoding utf8
"[$(Get-Date -Format o)] Agent exited with code $LASTEXITCODE" | Out-File -LiteralPath $LogPath -Append -Encoding utf8
