# 从 *.example.yaml 生成本地真实配置（已存在则跳过，不覆盖）
# 用法：powershell -File scripts/setup-config.ps1
# 对照 Nest：复制 blog-server/.env.example -> .env.development 后填真实值

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

$pairs = @(
    @{ Example = "configs/monolith.example.yaml";            Target = "configs/monolith.yaml" }
    @{ Example = "configs/monolith.production.example.yaml";  Target = "configs/monolith.production.yaml" }
    @{ Example = "configs/user.example.yaml";               Target = "configs/user.yaml" }
    @{ Example = "configs/blog.example.yaml";               Target = "configs/blog.yaml" }
    @{ Example = "configs/rpg.example.yaml";                Target = "configs/rpg.yaml" }
    @{ Example = "configs/gateway.example.yaml";            Target = "configs/gateway.yaml" }
    @{ Example = "configs/docker/user.example.yaml";        Target = "configs/docker/user.yaml" }
    @{ Example = "configs/docker/blog.example.yaml";        Target = "configs/docker/blog.yaml" }
    @{ Example = "configs/docker/rpg.example.yaml";         Target = "configs/docker/rpg.yaml" }
    @{ Example = "configs/docker/gateway.example.yaml";     Target = "configs/docker/gateway.yaml" }
    @{ Example = "configs/docker/monolith.example.yaml";    Target = "configs/docker/monolith.yaml" }
    @{ Example = "deploy/pm2/env.production.example";       Target = "deploy/pm2/env.production" }
)

$created = 0
$skipped = 0

foreach ($p in $pairs) {
    $examplePath = Join-Path $Root $p.Example
    $targetPath = Join-Path $Root $p.Target

    if (-not (Test-Path $examplePath)) {
        Write-Warning "missing example: $($p.Example)"
        continue
    }

    if (Test-Path $targetPath) {
        Write-Host "skip (exists): $($p.Target)"
        $skipped++
        continue
    }

    $targetDir = Split-Path $targetPath -Parent
    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }

    Copy-Item -Path $examplePath -Destination $targetPath
    Write-Host "created: $($p.Target) <- $($p.Example)"
    $created++
}

Write-Host ""
Write-Host "done: created=$created skipped=$skipped"
if ($created -gt 0) {
    Write-Host "fill secrets: dev configs/*.yaml; prod deploy/pm2/env.production (format same as blog-server, maintained in this repo)"
}
