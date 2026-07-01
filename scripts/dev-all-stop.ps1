# 停止 dev-all.ps1 启动的四服务（按端口 5001–5003、8000、50052）
$ErrorActionPreference = "SilentlyContinue"
$Root = Split-Path -Parent $PSScriptRoot
$PidFile = Join-Path $Root ".dev-all.pids"

$ports = @(5001, 5002, 5003, 8000, 50052)
$killed = @()

foreach ($port in $ports) {
    $lines = netstat -ano | Select-String ":$port\s" | Select-String "LISTENING"
    foreach ($line in $lines) {
        $parts = ($line -replace '\s+', ' ').ToString().Trim().Split(' ')
        $procId = [int]$parts[-1]
        if ($procId -gt 0 -and $killed -notcontains $procId) {
            Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
            $killed += $procId
            Write-Host "已停止 PID $procId (端口 $port)"
        }
    }
}

if (Test-Path $PidFile) { Remove-Item $PidFile -Force }

if ($killed.Count -eq 0) {
    Write-Host "未发现运行中的 dev 四服务（5001–5003 / 8000）"
} else {
    Write-Host "共停止 $($killed.Count) 个进程"
}
