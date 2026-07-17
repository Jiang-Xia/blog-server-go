# Kitex protobuf 代码生成（需 kitex CLI + protoc）
# 安装：go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3
# protoc：https://github.com/protocolbuffers/protobuf/releases （include 目录需含 google/protobuf）
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

$kitex = Get-Command kitex -ErrorAction SilentlyContinue
if (-not $kitex) {
    $kitexPath = Join-Path (go env GOPATH) "bin\kitex.exe"
    if (Test-Path $kitexPath) {
        $env:Path = "$(Split-Path $kitexPath);$env:Path"
    } else {
        Write-Error "kitex not found. Run: go install github.com/cloudwego/kitex/tool/cmd/kitex@v0.16.3"
    }
}

$protocInclude = $env:PROTOC_INCLUDE
if (-not $protocInclude) {
    foreach ($candidate in @("D:\env\protoc\include", "$env:LOCALAPPDATA\protoc\include", "C:\protoc\include")) {
        if (Test-Path (Join-Path $candidate "google\protobuf\empty.proto")) {
            $protocInclude = $candidate
            break
        }
    }
}
if (-not $protocInclude) {
    Write-Error "PROTOC_INCLUDE not set and google/protobuf/empty.proto not found"
}

$module = "github.com/Jiang-Xia/blog-server-go"
$jobs = @(
    @{ Service = "blog.user"; Idl = "proto/user/v1/user.proto" },
    @{ Service = "blog.blog"; Idl = "proto/blog/v1/article.proto" },
    @{ Service = "blog.rpg";  Idl = "proto/rpg/v1/rpg.proto" }
)

New-Item -ItemType Directory -Force -Path "proto\kitex" | Out-Null

foreach ($j in $jobs) {
    Write-Host "kitex generate $($j.Service) from $($j.Idl)"
    & kitex -module $module -I proto -I $protocInclude -type protobuf -service $j.Service -gen-path proto/kitex $j.Idl
    if ($LASTEXITCODE -ne 0) {
        Write-Error "kitex failed for $($j.Service)"
    }
}

# kitex -service 会在仓库根生成样板，删除以免污染模块
foreach ($junk in @("main.go", "handler.go", "build.sh")) {
    if (Test-Path $junk) {
        Remove-Item -Force $junk
        Write-Host "removed scaffold $junk"
    }
}

Write-Host "kitex generate done -> proto/kitex/"
