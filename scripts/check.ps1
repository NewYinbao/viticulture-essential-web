$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    & go test -tags "ee_rule_audit audit_diff" ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed" }
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "Go vet failed" }
} finally { Pop-Location }
