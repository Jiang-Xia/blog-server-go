# 查看 dev-all 四微服务运行状态、端口、健康检查
# 用法：
#   .\scripts\dev-all-status.ps1
#   .\scripts\dev-all-status.ps1 -Json
param([switch]$Json)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "ps-console-utf8.ps1")
. (Join-Path $PSScriptRoot "dev-all-common.ps1")
$Root = Split-Path -Parent $PSScriptRoot
Initialize-DevAllCommon -Root $Root

$pidMap = Read-DevAllPidMap
$rows = @()

foreach ($svc in Get-DevAllServices) {
    $listening = Test-DevPortListening $svc.Port
    $portPid = Get-DevPortPid $svc.Port
    if ($listening) {
        $health = Get-DevHealthStatus $svc.Port
    } else {
        $health = "-"
    }
    $launcherPid = $null
    if ($pidMap.ContainsKey($svc.Name)) {
        $launcherPid = $pidMap[$svc.Name]
    }

    $rows += [pscustomobject]@{
        Service     = $svc.Name
        Port        = $svc.Port
        Listening   = $listening
        Health      = $health
        PortPid     = $portPid
        LauncherPid = $launcherPid
        Config      = $svc.Config
        HealthUrl   = "http://127.0.0.1:$($svc.Port)/api/v1/health"
        Log         = (Get-DevServiceLogPath $svc.Name)
        ErrLog      = (Get-DevServiceLogPath $svc.Name -Errors)
    }
}

$infra = Get-DevInfraStatus

if ($Json) {
    $rows | ConvertTo-Json -Depth 3
    exit 0
}

Write-Host ""
Write-Host "blog-server-go dev-all status" -ForegroundColor Cyan
Write-Host ("-" * 72)

$fmt = "{0,-10} {1,5} {2,10} {3,8} {4,8} {5}"
Write-Host ($fmt -f "SERVICE", "PORT", "LISTEN", "HEALTH", "PID", "CONFIG") -ForegroundColor DarkGray

foreach ($r in $rows) {
    if ($r.Listening) { $listen = "yes" } else { $listen = "no" }
    if ($r.Health -eq "ok") { $color = "Green" }
    elseif ($r.Health -eq "down") { $color = "Red" }
    elseif ($r.Health -eq "bad") { $color = "Yellow" }
    else { $color = "DarkGray" }

    if ($null -ne $r.PortPid) { $pidText = "$($r.PortPid)" } else { $pidText = "-" }
    $line = $fmt -f $r.Service, $r.Port, $listen, $r.Health, $pidText, $r.Config
    Write-Host $line -ForegroundColor $color
}

Write-Host ""
if ($infra.Ok) {
    Write-Host "infra: MySQL :3306 / Redis :6379 listening" -ForegroundColor Green
} else {
    Write-Host "infra not ready: $($infra.Missing -join ', ')" -ForegroundColor Red
    Write-Host "  start MySQL/Redis before dev-all"
}

$up = @($rows | Where-Object { $_.Health -eq "ok" }).Count
$total = $rows.Count
Write-Host ""
if ($up -eq $total) {
    $summaryColor = "Green"
} else {
    $summaryColor = "Yellow"
}
Write-Host "health: $up / $total    gateway: http://127.0.0.1:8000/api/v1/health" -ForegroundColor $summaryColor

if ($up -lt $total) {
    Write-Host ""
    Write-Host "troubleshoot:" -ForegroundColor Yellow
    foreach ($r in ($rows | Where-Object { $_.Health -ne "ok" })) {
        Write-Host "  $($r.Service): .\scripts\dev-all-logs.ps1 -Service $($r.Service) -Errors -Tail 50"
    }
}

Write-DevAllHelpFooter
