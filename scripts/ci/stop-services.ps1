$ErrorActionPreference = "Stop"
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$PidFile = Join-Path $Root ".ci-services.pids"

if (Test-Path $PidFile) {
    Get-Content $PidFile | ForEach-Object {
        if ($_ -match '^(.+)=(\d+)$') {
            $name = $Matches[1]; $pid = [int]$Matches[2]
            Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
            Write-Host "stopped $name ($pid)"
        }
    }
    Remove-Item $PidFile -Force -ErrorAction SilentlyContinue
}

foreach ($port in 8000, 5001, 5002, 5003) {
    $lines = netstat -ano | Select-String ":$port\s" | Select-String "LISTENING"
    foreach ($line in $lines) {
        $procId = ($line -split '\s+')[-1]
        if ($procId -match '^\d+$') {
            Stop-Process -Id ([int]$procId) -Force -ErrorAction SilentlyContinue
        }
    }
}
