# 停止 dev-all.ps1 启动的四服务（HTTP + Kitex 端口）
$ErrorActionPreference = "SilentlyContinue"
. (Join-Path $PSScriptRoot "ps-console-utf8.ps1")
. (Join-Path $PSScriptRoot "dev-all-common.ps1")
$Root = Split-Path -Parent $PSScriptRoot
Initialize-DevAllCommon -Root $Root

$ports = @()
foreach ($svc in Get-DevAllServices) { $ports += $svc.Port }
$ports += Get-DevAllGrpcPorts
$ports = $ports | Sort-Object -Unique

$killed = @()

foreach ($port in $ports) {
    $procId = Get-DevPortPid $port
    if ($null -ne $procId -and $killed -notcontains $procId) {
        Stop-Process -Id $procId -Force -ErrorAction SilentlyContinue
        $killed += $procId
        Write-Host "已停止 PID $procId (端口 $port)"
    }
}

if (Test-Path $script:DevAllPidFile) { Remove-Item $script:DevAllPidFile -Force }

if ($killed.Count -eq 0) {
    Write-Host "未发现运行中的 dev 四服务（5001–5003 / 8000 / Kitex 50051–50053）"
} else {
    Write-Host "共停止 $($killed.Count) 个进程"
}
