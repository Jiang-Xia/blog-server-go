# 清空 x_my_blog 并从 myblog 全量同步数据，随后 FLUSHDB Redis（与 monolith.yaml 中 redis.db 一致）
# 用法：在 blog-server-go 根目录执行 .\scripts\sync-data-x-my-blog.ps1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "==> sync myblog -> x_my_blog + flush redis"
go run scripts/sync_data_x_my_blog.go

Write-Host "sync-data-x-my-blog done"
