# 一键启动 Plan 10 四服务（user → blog/rpg → gateway）
# 用法（blog-server-go 根目录）：
#   .\scripts\dev-all.ps1              # 先 go build 再后台启动（推荐）
#   .\scripts\dev-all.ps1 -Windows     # 四个独立 PowerShell 窗口
#   .\scripts\dev-all.ps1 -GoRun       # 强制 go run（首次编译很慢，易超时）
#   .\scripts\dev-all-status.ps1
#   .\scripts\dev-all-logs.ps1
#   .\scripts\dev-all-stop.ps1
param(
    [switch]$Windows,
    [switch]$SkipInfraCheck,
    [switch]$GoRun,
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root
. (Join-Path $PSScriptRoot "ps-console-utf8.ps1")
. (Join-Path $PSScriptRoot "dev-all-common.ps1")
Initialize-DevAllCommon -Root $Root

$services = Get-DevAllServices
$binDir = Join-Path $Root "bin"

function Clear-DevServiceLogs {
    New-Item -ItemType Directory -Force -Path $script:DevAllLogDir | Out-Null
    foreach ($svc in $services) {
        foreach ($path in @(
            (Get-DevServiceLogPath $svc.Name),
            (Get-DevServiceLogPath $svc.Name -Errors)
        )) {
            try {
                Set-Content -Path $path -Value "" -Encoding UTF8 -ErrorAction Stop
            } catch {
                $bak = "$path.bak-$(Get-Date -Format 'HHmmss')"
                try {
                    Move-Item -LiteralPath $path -Destination $bak -Force -ErrorAction Stop
                } catch {
                    Write-Host "警告: 无法清空日志 $path（可能被占用），继续启动" -ForegroundColor Yellow
                }
            }
        }
    }
}

function Build-DevAllBinaries {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    Write-Host "编译四服务二进制到 bin/ ..." -ForegroundColor Cyan
    foreach ($svc in $services) {
        $out = Join-Path $binDir "$($svc.Name).exe"
        Write-Host "  go build $($svc.Name) -> $out"
        & go build -o $out $svc.Main
        if ($LASTEXITCODE -ne 0) {
            throw "go build $($svc.Name) failed"
        }
    }
}

function Start-DevService($svc) {
    $configAbs = Join-Path $Root $svc.Config
    $envBlock = "`$env:CONFIG_PATH='$configAbs'"
    $title = "blog-server-go | $($svc.Name)-service :$($svc.Port)"
    $banner = "[dev-all] $($svc.Name)-service  :$($svc.Port)  CONFIG=$($svc.Config)"
    $binPath = Join-Path $binDir "$($svc.Name).exe"
    $useBin = (-not $GoRun) -and (Test-Path $binPath)

    if ($Windows) {
        if ($useBin) {
            $runCmd = @"
$envBlock
`$host.UI.RawUI.WindowTitle = '$title'
Set-Location '$Root'
Write-Host '$banner (bin)' -ForegroundColor Cyan
& '$binPath'
"@
        } else {
            $runCmd = @"
$envBlock
`$host.UI.RawUI.WindowTitle = '$title'
Set-Location '$Root'
Write-Host '$banner (go run)' -ForegroundColor Cyan
go run $($svc.Main)
"@
        }
        $proc = Start-Process powershell -ArgumentList @("-NoExit", "-Command", $runCmd) -PassThru
        return $proc.Id
    }

    New-Item -ItemType Directory -Force -Path $script:DevAllLogDir | Out-Null
    $outLog = Get-DevServiceLogPath $svc.Name
    $errLog = Get-DevServiceLogPath $svc.Name -Errors
    if ($useBin) {
        $hidden = "$envBlock; Set-Location '$Root'; & '$binPath' 1>> '$outLog' 2>> '$errLog'"
    } else {
        $hidden = "$envBlock; Set-Location '$Root'; go run $($svc.Main) 1>> '$outLog' 2>> '$errLog'"
    }
    $proc = Start-Process powershell -ArgumentList @("-WindowStyle", "Hidden", "-Command", $hidden) -PassThru
    return $proc.Id
}

# 基础设施预检
if (-not $SkipInfraCheck) {
    $infra = Get-DevInfraStatus
    if (-not $infra.Ok) {
        Write-Host "MySQL/Redis/Nacos 未监听: $($infra.Missing -join ', ')" -ForegroundColor Red
        Write-Host "请先启动本机 MySQL(3306)、Redis(6379)、Nacos(8848)，或加 -SkipInfraCheck 跳过"
        Write-Host "Nacos 推荐 Docker:" -ForegroundColor DarkGray
        Write-Host "  .\scripts\start-nacos.ps1" -ForegroundColor DarkGray
        Write-Host "  或: docker run -d --name blog-nacos -e MODE=standalone -e NACOS_AUTH_ENABLE=false -e JVM_XMS=256m -e JVM_XMX=256m -p 8848:8848 -p 9848:9848 nacos/nacos-server:v2.3.2" -ForegroundColor DarkGray
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

Clear-DevServiceLogs

if (-not $GoRun) {
    if (-not $SkipBuild) {
        Build-DevAllBinaries
    } else {
        foreach ($svc in $services) {
            $binPath = Join-Path $binDir "$($svc.Name).exe"
            if (-not (Test-Path $binPath)) {
                Write-Host "缺少 $binPath，请去掉 -SkipBuild 或先 make build" -ForegroundColor Red
                exit 1
            }
        }
    }
} else {
    Write-Host "使用 go run（首次编译可能超过 2 分钟）..." -ForegroundColor Yellow
}

$healthTimeout = if ($GoRun) { 180 } else { 60 }

Write-Host "启动四服务（MySQL + Redis + Nacos 已就绪）..." -ForegroundColor Cyan
$pids = @()

foreach ($svc in $services) {
    Write-Host "  -> $($svc.Name) :$($svc.Port)"
    $procId = Start-DevService $svc
    $pids += "$($svc.Name)=$procId"

    if ($svc.AfterStart) {
        if (-not (Wait-DevHealth $svc.Port $healthTimeout)) {
            Write-Host "  $($svc.Name) 启动超时" -ForegroundColor Red
            Show-DevLogTail $svc.Name -Lines 40 -Errors
            Show-DevLogTail $svc.Name -Lines 20
            & (Join-Path $PSScriptRoot "dev-all-stop.ps1")
            exit 1
        }
        Write-Host "  $($svc.Name) ready" -ForegroundColor Green
    } else {
        Start-Sleep -Seconds 1
    }
}

foreach ($svc in ($services | Where-Object { -not $_.AfterStart })) {
    if (-not (Wait-DevHealth $svc.Port $healthTimeout)) {
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
