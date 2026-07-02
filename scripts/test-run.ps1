# 行业规范本地全量测试（Windows）
# 用法：
#   .\scripts\test-run.ps1              # docker 测试库 + 四层测试
#   .\scripts\test-run.ps1 -SkipDocker  # 使用本机已有 MySQL/Redis（3306/6379）
#   .\scripts\test-run.ps1 -UnitOnly    # 仅单元+覆盖率
param(
    [switch]$SkipDocker,
    [switch]$UnitOnly
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$env:CI_JWT_SECRET = if ($env:CI_JWT_SECRET) { $env:CI_JWT_SECRET } else { "ci-integration-test-secret" }
$env:DEV_LOGIN_BASE = if ($env:DEV_LOGIN_BASE) { $env:DEV_LOGIN_BASE } else { "http://127.0.0.1:8000" }
$env:TEST_BASE = if ($env:TEST_BASE) { $env:TEST_BASE } else { $env:DEV_LOGIN_BASE }

if (-not $SkipDocker) {
    $env:CI_MYSQL_HOST = "127.0.0.1"
    $env:CI_MYSQL_PORT = "3307"
    $env:CI_MYSQL_USER = "root"
    $env:CI_MYSQL_PASSWORD = "testpass"
    $env:CI_MYSQL_DATABASE = "x_my_blog"
    $env:CI_REDIS_ADDR = "127.0.0.1:6380"
    $env:CI_REDIS_DB = "2"
} else {
    $env:CI_MYSQL_PORT = if ($env:CI_MYSQL_PORT) { $env:CI_MYSQL_PORT } else { "3306" }
    $env:CI_REDIS_ADDR = if ($env:CI_REDIS_ADDR) { $env:CI_REDIS_ADDR } else { "127.0.0.1:6379" }
    $env:CI_REDIS_DB = if ($env:CI_REDIS_DB) { $env:CI_REDIS_DB } else { "1" }
}

function Stop-All {
    & "$PSScriptRoot\ci\stop-services.ps1" 2>$null
    if (-not $SkipDocker) {
        docker compose -f deploy/docker/docker-compose.test.yml down -v 2>$null
    }
}

try {
    if ($UnitOnly) {
        & "$PSScriptRoot\ci\check-coverage.ps1"
        if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
        exit 0
    }

    if (-not $SkipDocker) {
        Write-Host ">>> docker compose test infra" -ForegroundColor Cyan
        docker compose -f deploy/docker/docker-compose.test.yml up -d --wait
        if ($LASTEXITCODE -ne 0) { throw "docker compose failed" }
    }

    # wait-for 需 Git Bash / WSL；无 bash 时简单 sleep
    if (Get-Command bash -ErrorAction SilentlyContinue) {
        bash scripts/ci/wait-for.sh
    } else {
        Write-Host "wait 15s for mysql/redis..."
        Start-Sleep -Seconds 15
    }

    go run ./scripts/ci/env.go ./scripts/ci/prepare_config.go
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    go run ./scripts/ci/env.go ./scripts/ci/migrate_schemas.go
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    go run ./scripts/ci/env.go ./scripts/ci/seed_test_data.go
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

    & "$PSScriptRoot\ci\start-services.ps1"

    $env:RUN_UNIT = "1"
    $env:RUN_SMOKE = "1"
    $env:RUN_INTEGRATION = "1"
    $env:RUN_E2E = "1"
    & "$PSScriptRoot\ci\run-layer-tests.ps1"

    Write-Host ">>> test-run.ps1 OK" -ForegroundColor Green
}
finally {
    Stop-All
}
