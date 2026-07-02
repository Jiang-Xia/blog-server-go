# 按分层顺序跑测试（Windows，需服务已启动时跑 smoke/integration/e2e）
$ErrorActionPreference = "Stop"
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
Set-Location $Root

$env:DEV_LOGIN_BASE = if ($env:DEV_LOGIN_BASE) { $env:DEV_LOGIN_BASE } else { "http://127.0.0.1:8000" }
$env:TEST_BASE = if ($env:TEST_BASE) { $env:TEST_BASE } else { $env:DEV_LOGIN_BASE }

function Run-GoTest([string]$Label, [string[]]$Args) {
    Write-Host "=== $Label ===" -ForegroundColor Cyan
    & go @Args
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}

$runUnit = if ($env:RUN_UNIT -eq "0") { $false } else { $true }
$runSmoke = if ($env:RUN_SMOKE -eq "0") { $false } else { $true }
$runIntegration = if ($env:RUN_INTEGRATION -eq "0") { $false } else { $true }
$runE2E = if ($env:RUN_E2E -eq "0") { $false } else { $true }

if ($runUnit) {
    & "$PSScriptRoot\check-coverage.ps1"
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
}
if ($runSmoke) {
    Run-GoTest "smoke" @("test", "-tags=smoke", "./test/smoke/...", "-count=1", "-v")
}
if ($runIntegration) {
    Run-GoTest "integration" @("test", "-tags=integration", "./test/integration/...", "-count=1", "-v")
}
if ($runE2E) {
    Run-GoTest "e2e" @("test", "-tags=e2e", "./test/e2e/...", "-count=1", "-v")
}

Write-Host "=== all requested test layers passed ===" -ForegroundColor Green
