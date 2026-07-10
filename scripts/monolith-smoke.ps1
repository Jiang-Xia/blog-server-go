# 单体 :5000 冒烟脚本（Plan 22）
# 用法：先 make dev / .\scripts\dev.ps1，再 .\scripts\monolith-smoke.ps1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$base = if ($env:TEST_BASE) { $env:TEST_BASE.TrimEnd('/') } else { "http://127.0.0.1:5000" }
$api = "$base/api/v1"

$env:DEV_LOGIN_BASE = $base
$token = (go run scripts/dev_login.go --token-only 2>$null).Trim()
if (-not $token) { throw "dev_login 失败，请确认单体已启动且 MySQL/Redis 可用" }

$headers = @{ Authorization = "Bearer $token" }
$passed = 0
$failed = 0

function Test-Api {
    param([string]$Name, [string]$Url, [bool]$Auth = $false, [string]$Method = "GET", [string]$Body = $null)
    try {
        $h = if ($Auth) { $headers } else { @{} }
        if ($Method -eq "POST") {
            $r = Invoke-RestMethod -Uri $Url -Method POST -Body $Body -ContentType "application/json" -Headers $h
        } else {
            $r = Invoke-RestMethod -Uri $Url -Headers $h
        }
        if ($r.code -eq 200) {
            Write-Host "[OK] $Name"
            $script:passed++
        } else {
            Write-Host "[FAIL] $Name code=$($r.code) msg=$($r.message)"
            $script:failed++
        }
    } catch {
        Write-Host "[FAIL] $Name $($_.Exception.Message)"
        $script:failed++
    }
}

Write-Host "monolith smoke base=$base"
Test-Api "health" "$base/health"
Test-Api "pub/stats" "$api/pub/stats"
Test-Api "rag/status" "$api/rag/status"
Test-Api "rag/quota" "$api/rag/quota" -Auth $true
Test-Api "scheduled-task/tasks" "$api/scheduled-task/tasks" -Auth $true
Test-Api "user/info" "$api/user/info" -Auth $true
Test-Api "article/list" "$api/article/list" -Method POST -Body "{}"

Write-Host "`n结果: passed=$passed failed=$failed"
if ($failed -gt 0) { exit 1 }
