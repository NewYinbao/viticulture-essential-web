param([string]$Output = "dist/Viticulture.exe")
$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    $outputPath = [IO.Path]::GetFullPath((Join-Path $projectRoot $Output))
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputPath) | Out-Null
    & go build -trimpath -o $outputPath ./cmd/viticulture
    if ($LASTEXITCODE -ne 0) { throw "Go build failed" }
    Write-Host "Built $outputPath"
} finally { Pop-Location }
