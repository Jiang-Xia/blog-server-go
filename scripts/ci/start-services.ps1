# CI / test-run 后台启动四微服务（Windows）
$ErrorActionPreference = "Stop"
$Root = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
Set-Location $Root

$PidFile = Join-Path $Root ".ci-services.pids"
$LogDir = Join-Path $Root ".ci-logs"
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
Set-Content -Path $PidFile -Value "" -Encoding UTF8

function Start-One($Name, $Cfg, $Main, $Port) {
    Write-Host "start $Name :$Port"
    $cmd = "`$env:CONFIG_PATH='$Cfg'; Set-Location '$Root'; go run $Main 1>> '$LogDir\$Name.log' 2>> '$LogDir\$Name.err.log'"
    $p = Start-Process powershell -ArgumentList @("-WindowStyle", "Hidden", "-Command", $cmd) -PassThru
    Add-Content -Path $PidFile -Value "$Name=$($p.Id)" -Encoding UTF8
}

function Wait-Health([int]$Port, [string]$Name) {
    $deadline = (Get-Date).AddSeconds(90)
    while ((Get-Date) -lt $deadline) {
        try {
            $r = Invoke-RestMethod -Uri "http://127.0.0.1:$Port/api/v1/health" -TimeoutSec 3
            if ($r.code -eq 200) { Write-Host "$Name ready"; return }
        } catch {}
        Start-Sleep -Seconds 1
    }
    throw "$Name health timeout, see $LogDir\$Name.err.log"
}

Start-One user configs/user.yaml ./services/user/cmd/main.go 5002
Start-Sleep -Seconds 2
Start-One blog configs/blog.yaml ./services/blog/cmd/main.go 5001
Start-One rpg  configs/rpg.yaml  ./services/rpg/cmd/main.go  5003
Start-Sleep -Seconds 2
Start-One gateway configs/gateway.yaml ./services/gateway/cmd/main.go 8000

Wait-Health 5002 user
Wait-Health 5001 blog
Wait-Health 5003 rpg
Wait-Health 8000 gateway
Write-Host "all services up"
