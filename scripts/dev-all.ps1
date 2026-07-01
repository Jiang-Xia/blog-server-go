# 一键启动 Plan 10 四服务（user → blog/rpg → gateway）
# 用法（blog-server-go 根目录）：
#   .\scripts\dev-all.ps1              # 后台启动，日志写入 .dev-logs/
#   .\scripts\dev-all.ps1 -Windows     # 四个独立 PowerShell 窗口（方便看日志）
#   .\scripts\dev-all-stop.ps1         # 停止
param(
    [switch]$Windows
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$PidFile = Join-Path $Root ".dev-all.pids"
$LogDir = Join-Path $Root ".dev-logs"

$services = @(
    @{ Name = "user";    Config = "configs/user.yaml";    Main = "./services/monolith/cmd/user/main.go";    Port = 5002; AfterStart = $true }
    @{ Name = "blog";    Config = "configs/blog.yaml";    Main = "./services/monolith/cmd/blog/main.go";    Port = 5001 }
    @{ Name = "rpg";     Config = "configs/rpg.yaml";     Main = "./services/monolith/cmd/rpg/main.go";     Port = 5003 }
    @{ Name = "gateway"; Config = "configs/gateway.yaml"; Main = "./services/gateway/cmd/main.go";          Port = 8000; AfterStart = $true }
)

function Test-PortListening([int]$Port) {
    $line = netstat -ano | Select-String ":$Port\s" | Select-String "LISTENING" | Select-Object -First 1
    return $null -ne $line
}

function Wait-Health([int]$Port, [int]$TimeoutSec = 90) {
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        try {
            $r = Invoke-RestMethod -Uri "http://127.0.0.1:$Port/api/v1/health" -TimeoutSec 3
            if ($r.code -eq 200 -or $r.status -eq "ok") { return $true }
        } catch { }
        Start-Sleep -Milliseconds 600
    }
    return $false
}

function Start-DevService($svc) {
    $envBlock = "`$env:CONFIG_PATH='$($svc.Config)'"
    $runCmd = "$envBlock; Set-Location '$Root'; go run $($svc.Main)"

    if ($Windows) {
        $proc = Start-Process powershell -ArgumentList @("-NoExit", "-Command", $runCmd) -PassThru
        return $proc.Id
    }

    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
    $outLog = Join-Path $LogDir "$($svc.Name).log"
    $errLog = Join-Path $LogDir "$($svc.Name).err.log"
    $hidden = "$envBlock; Set-Location '$Root'; go run $($svc.Main) 1>> '$outLog' 2>> '$errLog'"
    $proc = Start-Process powershell -ArgumentList @("-WindowStyle", "Hidden", "-Command", $hidden) -PassThru
    return $proc.Id
}

# 已有进程在跑则提示
$busy = @()
foreach ($svc in $services) {
    if (Test-PortListening $svc.Port) { $busy += $svc.Name }
}
if ($busy.Count -gt 0) {
    Write-Host "以下端口已被占用: $($busy -join ', ')。请先执行 .\scripts\dev-all-stop.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host "启动四服务（需本机 MySQL + Redis 已就绪）..." -ForegroundColor Cyan
$pids = @()

foreach ($svc in $services) {
    Write-Host "  -> $($svc.Name) :$($svc.Port)"
    $pid = Start-DevService $svc
    $pids += "$($svc.Name)=$pid"

    if ($svc.AfterStart) {
        if (-not (Wait-Health $svc.Port)) {
            Write-Host "  $($svc.Name) 启动超时，请查看 $LogDir\$($svc.Name).err.log" -ForegroundColor Red
            & (Join-Path $PSScriptRoot "dev-all-stop.ps1")
            exit 1
        }
        Write-Host "  $($svc.Name) ready" -ForegroundColor Green
    } else {
        Start-Sleep -Seconds 2
    }
}

# 等待 blog/rpg
foreach ($svc in ($services | Where-Object { -not $_.AfterStart })) {
    if (-not (Wait-Health $svc.Port 60)) {
        Write-Host "  $($svc.Name) 启动超时" -ForegroundColor Red
        & (Join-Path $PSScriptRoot "dev-all-stop.ps1")
        exit 1
    }
    Write-Host "  $($svc.Name) ready" -ForegroundColor Green
}

Set-Content -Path $PidFile -Value ($pids -join "`n") -Encoding UTF8

Write-Host ""
Write-Host "四服务已启动。前端/联调请指向:" -ForegroundColor Green
Write-Host "  API  baseUrl  http://127.0.0.1:8000"
Write-Host "  登录 DEV_LOGIN_BASE=http://127.0.0.1:8000"
Write-Host ""
Write-Host "停止: .\scripts\dev-all-stop.ps1"
if (-not $Windows) {
    Write-Host "日志: $LogDir\"
}
