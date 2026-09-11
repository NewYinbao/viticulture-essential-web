param([string]$Go = "")
$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    New-Item -ItemType Directory -Force -Path (Join-Path $projectRoot "dist") | Out-Null
    if (-not $Go) {
        $nativeGo = Get-Command go -CommandType Application -ErrorAction SilentlyContinue
        if ($nativeGo) { $Go = $nativeGo.Source }
    }
    if ($Go) {
        $previousTargetOS = $env:GOOS
        $previousTargetArch = $env:GOARCH
        try {
            $env:GOOS = "windows"
            $env:GOARCH = "amd64"
            & $Go build -trimpath -o "dist/Viticulture-Continuation.exe" ./cmd/viticulture
            if ($LASTEXITCODE -ne 0) { throw "Continuation build failed" }
        } finally {
            $env:GOOS = $previousTargetOS
            $env:GOARCH = $previousTargetArch
        }
    } else {
        $linuxProjectRoot = & wsl.exe --exec wslpath -a -u $projectRoot
        if ($LASTEXITCODE -ne 0) { throw "Cannot resolve the project path in WSL" }
        & wsl.exe --cd $linuxProjectRoot --exec env GOOS=windows GOARCH=amd64 /opt/viticulture-toolchain/go/bin/go build -trimpath -o dist/Viticulture-Continuation.exe ./cmd/viticulture
        if ($LASTEXITCODE -ne 0) { throw "Build failed; provide a Windows Go path with -Go" }
    }
    Write-Host "Built dist/Viticulture-Continuation.exe; no server started."
} finally { Pop-Location }
