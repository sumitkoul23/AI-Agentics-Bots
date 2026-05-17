$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Binary = Join-Path $ProjectDir "category-agent-portfolio.exe"
$env:Path = "C:\Program Files\Go\bin;" + $env:Path
$env:GOCACHE = Join-Path $ProjectDir ".gocache"
$env:GOPATH = "C:\Users\Art-E Mediatech\AppData\Local\Teneo CLI\perpetual-market-strategist\.gopath"

Set-Location $ProjectDir
& "C:\Program Files\Go\bin\go.exe" build .

$metadataFiles = Get-ChildItem -LiteralPath (Join-Path $ProjectDir "agents") -Filter "*.json" | Sort-Object Name
foreach ($file in $metadataFiles) {
    $slug = [IO.Path]::GetFileNameWithoutExtension($file.Name)
    $out = Join-Path $ProjectDir "$slug.out.log"
    $err = Join-Path $ProjectDir "$slug.err.log"
    Start-Process -FilePath $Binary -ArgumentList ('"{0}"' -f $file.FullName) -WorkingDirectory $ProjectDir -WindowStyle Hidden -RedirectStandardOutput $out -RedirectStandardError $err | Out-Null
}
