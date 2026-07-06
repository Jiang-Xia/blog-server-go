# 一键启动 Plan 10 四服务（user → blog/rpg → gateway）
# 用法（blog-server-go 根目录）：
#   .\scripts\dev-all.ps1              # 后台启动，日志写入 .dev-logs/
#   .\scripts\dev-all.ps1 -Windows     # 四个独立 PowerShell 窗口（标题含服务名与端口）
#   .\scripts\dev-all-status.ps1         # 查看状态 / 健康检查
#   .\scripts\dev-all-logs.ps1           # 查看日志（-Service blog -Follow）
#   .\scripts\dev-all-stop.ps1         # 停止
param(
    [switch]$Windows,
    [switch]$SkipInfraCheck
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root
. (Join-Path $PSScriptRoot "ps-console-utf8.ps1")
. (Join-Path $PSScriptRoot "dev-all-common.ps1")
Initialize-DevAllCommon -Root $Root

$services = Get-DevAllServices

function Start-DevService($svc) {
    $envBlock = "`$env:CONFIG_PATH='$($svc.Config)'"
    $title = "blog-server-go | $($svc.Name)-service :$($svc.Port)"
    $banner = "[dev-all] $($svc.Name)-service  :$($svc.Port)  CONFIG=$($svc.Config)"

    if ($Windows) {
        $runCmd = @"
$envBlock
`$host.UI.RawUI.WindowTitle = '$title'
Set-Location '$Root'
Write-Host '$banner' -ForegroundColor Cyan
go run $($svc.Main)
"@
        $proc = Start-Process powershell -ArgumentList @("-NoExit", "-Command", $runCmd) -PassThru
        return $proc.Id
    }

    New-Item -ItemType Directory -Force -Path $script:DevAllLogDir | Out-Null
    $outLog = Get-DevServiceLogPath $svc.Name
    $errLog = Get-DevServiceLogPath $svc.Name -Errors
    $hidden = "$envBlock; Set-Location '$Root'; go run $($svc.Main) 1>> '$outLog' 2>> '$errLog'"
    $proc = Start-Process powershell -ArgumentList @("-WindowStyle", "Hidden", "-Command", $hidden) -PassThru
    return $proc.Id
}

# 基础设施预检
if (-not $SkipInfraCheck) {
    $infra = Get-DevInfraStatus
    if (-not $infra.Ok) {
        Write-Host "MySQL/Redis 未监听: $($infra.Missing -join ', ')" -ForegroundColor Red
        Write-Host "请先启动本机 MySQL(3306) 与 Redis(6379)，或加 -SkipInfraCheck 跳过"
        exit 1
    }
}

# 已有进程在跑则提示
$busy = @()
foreach ($svc in $services) {
    if (Test-DevPortListening $svc.Port) { $busy += "$($svc.Name):$($svc.Port)" }
}
if ($busy.Count -gt 0) {
    Write-Host "以下端口已被占用: $($busy -join ', ')" -ForegroundColor Yellow
    Write-Host "  状态: .\scripts\dev-all-status.ps1"
    Write-Host "  停止: .\scripts\dev-all-stop.ps1"
    exit 1
}

Write-Host "启动四服务（MySQL + Redis 已就绪）..." -ForegroundColor Cyan
$pids = @()

foreach ($svc in $services) {
    Write-Host "  -> $($svc.Name) :$($svc.Port)"
    $procId = Start-DevService $svc
    $pids += "$($svc.Name)=$procId"

    if ($svc.AfterStart) {
        if (-not (Wait-DevHealth $svc.Port)) {
            Write-Host "  $($svc.Name) 启动超时" -ForegroundColor Red
            Show-DevLogTail $svc.Name -Lines 40 -Errors
            Show-DevLogTail $svc.Name -Lines 20
            & (Join-Path $PSScriptRoot "dev-all-stop.ps1")
            exit 1
        }
        Write-Host "  $($svc.Name) ready" -ForegroundColor Green
    } else {
        Start-Sleep -Seconds 2
    }
}

foreach ($svc in ($services | Where-Object { -not $_.AfterStart })) {
    if (-not (Wait-DevHealth $svc.Port 60)) {
        Write-Host "  $($svc.Name) 启动超时" -ForegroundColor Red
        Show-DevLogTail $svc.Name -Lines 40 -Errors
        Show-DevLogTail $svc.Name -Lines 20
        & (Join-Path $PSScriptRoot "dev-all-stop.ps1")
        exit 1
    }
    Write-Host "  $($svc.Name) ready" -ForegroundColor Green
}

Set-Content -Path $script:DevAllPidFile -Value ($pids -join "`n") -Encoding UTF8

Write-Host ""
Write-Host "四服务已启动。前端/联调请指向:" -ForegroundColor Green
Write-Host "  API  baseUrl  http://127.0.0.1:8000"
Write-Host "  登录 DEV_LOGIN_BASE=http://127.0.0.1:8000"
Write-Host ""
if (-not $Windows) {
    Write-Host "日志目录: $script:DevAllLogDir\"
} else {
    Write-Host "各窗口标题: blog-server-go | {user,blog,rpg,gateway}-service :端口"
}
Write-DevAllHelpFooter
