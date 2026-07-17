# dev-all 四微服务共享定义与辅助函数（供 dev-all / status / logs / stop 点源）
# 点源后须调用: Initialize-DevAllCommon -Root (Split-Path -Parent $PSScriptRoot)

$script:DevAllRoot = $null
$script:DevAllPidFile = $null
$script:DevAllLogDir = $null
$script:DevAllServices = @(
    @{ Name = "user";    Config = "configs/user.yaml";    Main = "./services/user/cmd/main.go";    Port = 5002; KitexPort = 50052; AfterStart = $true }
    @{ Name = "blog";    Config = "configs/blog.yaml";    Main = "./services/blog/cmd/main.go";    Port = 5001; KitexPort = 50051; AfterStart = $false }
    @{ Name = "rpg";     Config = "configs/rpg.yaml";     Main = "./services/rpg/cmd/main.go";     Port = 5003; KitexPort = 50053; AfterStart = $false }
    @{ Name = "gateway"; Config = "configs/gateway.yaml"; Main = "./services/gateway/cmd/main.go"; Port = 8000; AfterStart = $true }
)

function Initialize-DevAllCommon {
    param([Parameter(Mandatory = $true)][string]$Root)
    $script:DevAllRoot = $Root
    $script:DevAllPidFile = Join-Path $Root ".dev-all.pids"
    $script:DevAllLogDir = Join-Path $Root ".dev-logs"
}

function Get-DevAllServices { return $script:DevAllServices }

function Get-DevAllKitexPorts {
    $ports = @()
    foreach ($svc in $script:DevAllServices) {
        if ($svc.KitexPort) { $ports += $svc.KitexPort }
    }
    return $ports | Sort-Object -Unique
}

# 兼容旧名
function Get-DevAllGrpcPorts { return Get-DevAllKitexPorts }

function Test-DevPortListening([int]$Port) {
    $line = netstat -ano | Select-String ":$Port\s" | Select-String "LISTENING" | Select-Object -First 1
    return $null -ne $line
}

function Get-DevPortPid([int]$Port) {
    $line = netstat -ano | Select-String ":$Port\s" | Select-String "LISTENING" | Select-Object -First 1
    if (-not $line) { return $null }
    $parts = ($line -replace '\s+', ' ').ToString().Trim().Split(' ')
    $procId = [int]$parts[-1]
    if ($procId -le 0) { return $null }
    return $procId
}

function Get-DevHealthStatus([int]$Port) {
    try {
        $r = Invoke-RestMethod -Uri "http://127.0.0.1:$Port/api/v1/health" -TimeoutSec 2
        if ($r.code -eq 200 -or $r.status -eq "ok") { return "ok" }
        return "bad"
    } catch {
        return "down"
    }
}

function Wait-DevHealth([int]$Port, [int]$TimeoutSec = 90) {
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        if ((Get-DevHealthStatus $Port) -eq "ok") { return $true }
        Start-Sleep -Milliseconds 600
    }
    return $false
}

function Read-DevAllPidMap {
    $map = @{}
    if (-not (Test-Path $script:DevAllPidFile)) { return $map }
    foreach ($line in Get-Content $script:DevAllPidFile -Encoding UTF8) {
        $line = $line.Trim()
        if ($line -match '^([^=]+)=(\d+)$') {
            $map[$Matches[1]] = [int]$Matches[2]
        }
    }
    return $map
}

function Get-DevServiceLogPath([string]$Name, [switch]$Errors) {
    $suffix = if ($Errors) { ".err.log" } else { ".log" }
    return Join-Path $script:DevAllLogDir "$Name$suffix"
}

function Show-DevLogTail([string]$Name, [int]$Lines = 30, [switch]$Errors) {
    $paths = @()
    if ($Errors) {
        $paths += Get-DevServiceLogPath $Name -Errors
    } else {
        $paths += Get-DevServiceLogPath $Name
        $errPath = Get-DevServiceLogPath $Name -Errors
        if ((Test-Path $errPath) -and ((Get-Item $errPath).Length -gt 0)) {
            $paths += $errPath
        }
    }
    foreach ($path in $paths) {
        if (-not (Test-Path $path)) {
            Write-Host "  (no log $path)" -ForegroundColor DarkGray
            continue
        }
        Write-Host ""
        Write-Host "--- tail $path ---" -ForegroundColor Yellow
        Get-Content $path -Tail $Lines -Encoding UTF8 | ForEach-Object { Write-Host $_ }
    }
}

function Get-DevInfraStatus {
    $checks = @(
        @{ Label = "MySQL"; Port = 3306 }
        @{ Label = "Redis"; Port = 6379 }
        @{ Label = "etcd";  Port = 2379 }
    )
    $missing = @()
    foreach ($c in $checks) {
        if (-not (Test-DevPortListening $c.Port)) {
            $missing += "$($c.Label):$($c.Port)"
        }
    }
    return @{
        Ok      = ($missing.Count -eq 0)
        Missing = $missing
    }
}

function Write-DevAllHelpFooter {
    Write-Host ""
    Write-Host "commands:" -ForegroundColor Cyan
    Write-Host "  status  .\scripts\dev-all-status.ps1"
    Write-Host "  logs    .\scripts\dev-all-logs.ps1 -Service blog -Follow"
    Write-Host "  errors  .\scripts\dev-all-logs.ps1 -Service blog -Errors -Tail 80"
    Write-Host "  stop    .\scripts\dev-all-stop.ps1"
}
