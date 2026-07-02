# 单元测试覆盖率门禁（Windows）；默认 pkg 总覆盖率 ≥ 40%
param(
    [int]$MinCoverage = $(if ($env:MIN_PKG_COVERAGE) { [int]$env:MIN_PKG_COVERAGE } else { 40 })
)

$ErrorActionPreference = "Stop"
$profile = if ($env:COVERAGE_PROFILE) { $env:COVERAGE_PROFILE } else { "coverage.out" }

go test ./pkg/crypto ./pkg/jwtauth ./pkg/pagination ./pkg/errcode ./pkg/timeutil -count=1 -covermode=atomic -coverprofile=$profile
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$line = go tool cover -func=$profile | Select-String "^total:"
if (-not $line) {
    Write-Error "cannot parse coverage"
    exit 1
}
$total = [double]($line -replace '.*total:\s+\(statements\)\s+','' -replace '%','')
Write-Host "pkg coverage: $total% (min $MinCoverage%)"
if ($total -lt $MinCoverage) {
    Write-Error "coverage below threshold $MinCoverage%"
    exit 1
}
