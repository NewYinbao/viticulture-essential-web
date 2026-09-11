$ErrorActionPreference = "Stop"
$projectRoot = Split-Path -Parent $PSScriptRoot
Push-Location $projectRoot
try {
    & go test -tags "ee_rule_audit audit_diff" ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed" }
    & go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "Go vet failed" }
    $unformatted = & gofmt -l cmd internal web
    if ($LASTEXITCODE -ne 0) { throw "gofmt failed" }
    if ($unformatted) {
        $unformatted | Write-Error
        throw "Go sources are not gofmt-formatted"
    }
} finally { Pop-Location }
