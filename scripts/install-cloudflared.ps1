# Pinned official GitHub release asset; no PATH, services or firewall changes.
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$root = Split-Path -Parent $PSScriptRoot
$dir = Join-Path $root "dist/tools"
$version = "2026.9.0"
$url = "https://github.com/cloudflare/cloudflared/releases/download/2026.9.0/cloudflared-windows-amd64.exe"
$expected = "547057326266f0e1c7d50d102dbd22ff283d740c055bd61e94f10e2c606f89af"
New-Item -ItemType Directory -Force $dir | Out-Null
$tmp = Join-Path $dir (([guid]::NewGuid().ToString()) + ".download")
try {
 Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $tmp -TimeoutSec 180
 $actual = (Get-FileHash -Algorithm SHA256 $tmp).Hash.ToLowerInvariant()
 if ($actual -ne $expected) { throw "SHA256 mismatch; downloaded file will NOT be installed" }
 $target = Join-Path $dir "cloudflared.exe"
 if (Test-Path $target) {
  Copy-Item $target (Join-Path $dir ("cloudflared-backup-" + [guid]::NewGuid() + ".exe"))
 }
 Move-Item -Force $tmp $target
 Write-Host "Installed official cloudflared $version (Windows amd64), SHA256 $actual"
} finally { if (Test-Path $tmp) { Remove-Item $tmp } }
