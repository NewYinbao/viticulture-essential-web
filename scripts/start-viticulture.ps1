param(
 [ValidateSet("Ask", "LAN", "Public")][string]$Mode = "Ask",
 [string]$Address = "0.0.0.0:3013",
 [string]$Data = "runtime/ee-data"
)
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location $root
if (!(Test-Path "dist/Viticulture.exe")) { throw "Build first: powershell -File scripts/build.ps1" }
if ($Mode -eq "Ask") {
 Write-Host "Enable Cloudflare public link?  [1] LAN (default)  [2] Public  [y/N]"
 Write-Host "Public mode sends traffic through Cloudflare; share URL and random key privately."
 $answer = [Console]::ReadLine()
 $Mode = "LAN"
 if ($answer -and $answer.Trim().ToLowerInvariant() -in @("y", "yes", "2")) { $Mode = "Public" }
}
$arguments = @("-addr", $Address, "-data", $Data)
if ($Mode -eq "Public") {
 if (!(Test-Path "dist/tools/cloudflared.exe")) {
  throw "cloudflared is missing. Install explicitly: powershell -NoProfile -ExecutionPolicy Bypass -File scripts/install-cloudflared.ps1 . No server/tunnel started."
 }
 $arguments += @("-public", "-cloudflared", (Join-Path $root "dist/tools/cloudflared.exe"))
}
Write-Host "Open http://localhost:3013 (default port); friends use the LAN IP or printed HTTPS URL."
Write-Host "Mode: $Mode. Existing saves are not migrated. Stop the old server yourself before reusing its save directory."
& (Join-Path $root "dist/Viticulture.exe") @arguments
exit $LASTEXITCODE
