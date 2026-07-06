# 查看 dev-all 后台模式日志（.dev-logs/）
# 用法：
#   .\scripts\dev-all-logs.ps1                          # 四服务各 tail 30 行
#   .\scripts\dev-all-logs.ps1 -Service blog -Follow    # 跟踪 blog 日志
#   .\scripts\dev-all-logs.ps1 -Service blog -Errors    # 仅 stderr
#   .\scripts\dev-all-logs.ps1 -Service blog -Errors -Follow
param(
    [ValidateSet("user", "blog", "rpg", "gateway", "all")]
    [string]$Service = "all",
    [switch]$Follow,
    [switch]$Errors,
    [int]$Tail = 30
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "ps-console-utf8.ps1")
. (Join-Path $PSScriptRoot "dev-all-common.ps1")
$Root = Split-Path -Parent $PSScriptRoot
Initialize-DevAllCommon -Root $Root

$names = if ($Service -eq "all") {
    (Get-DevAllServices | ForEach-Object { $_.Name })
} else {
    @($Service)
}

if ($Follow -and $names.Count -gt 1) {
    Write-Host "-Follow requires single -Service (e.g. blog), not all" -ForegroundColor Yellow
    exit 1
}

if (-not (Test-Path $script:DevAllLogDir)) {
    Write-Host "log dir missing: $script:DevAllLogDir" -ForegroundColor Yellow
    Write-Host "background mode writes .dev-logs/; use -Windows mode or dev-all.ps1 -Windows"
    exit 1
}

foreach ($name in $names) {
    if ($Errors) {
        $path = Get-DevServiceLogPath $name -Errors
        $label = "$name.err.log"
    } else {
        $path = Get-DevServiceLogPath $name
        $label = "$name.log"
    }

    if (-not (Test-Path $path)) {
        Write-Host "[$label] not found (service not started or no output yet)" -ForegroundColor DarkGray
        continue
    }

    if ($Follow) {
        Write-Host "following $path (Ctrl+C to exit)" -ForegroundColor Cyan
        Get-Content $path -Wait -Tail $Tail -Encoding UTF8
        return
    }

    Write-Host ""
    Write-Host "=== $label (last $Tail) ===" -ForegroundColor Cyan
    Get-Content $path -Tail $Tail -Encoding UTF8 | ForEach-Object { Write-Host $_ }

    if (-not $Errors) {
        $errPath = Get-DevServiceLogPath $name -Errors
        if ((Test-Path $errPath) -and ((Get-Item $errPath).Length -gt 0)) {
            Write-Host ""
            Write-Host "=== $name.err.log (last $Tail) ===" -ForegroundColor Yellow
            Get-Content $errPath -Tail $Tail -Encoding UTF8 | ForEach-Object { Write-Host $_ }
        }
    }
}
