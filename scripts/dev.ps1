# 等价于 make dev（Windows PowerShell，无需安装 make）
# 用法：在 blog-server-go 根目录执行 .\scripts\dev.ps1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$env:CONFIG_PATH = if ($env:CONFIG_PATH) { $env:CONFIG_PATH } else { "configs/monolith.yaml" }
Write-Host "CONFIG_PATH=$env:CONFIG_PATH"
go run ./services/monolith/cmd/main.go
