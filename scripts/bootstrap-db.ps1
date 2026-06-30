# 等价于 make bootstrap-db（Windows PowerShell，无需安装 make）
# 用法：在 blog-server-go 根目录执行 .\scripts\bootstrap-db.ps1
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "==> bootstrap x_my_blog from myblog"
go run scripts/bootstrap_x_my_blog.go

Write-Host "==> generate ent schema from MySQL"
go run scripts/gen_ent_schema.go

Write-Host "==> ent codegen"
Push-Location services/monolith/ent
try {
    go generate ./...
} finally {
    Pop-Location
}

Write-Host "bootstrap-db done"
