# 启动本机 Nacos（学习路径 Kitex 注册中心；推荐 Docker）
# 用法: .\scripts\start-nacos.ps1
# 控制台: http://localhost:8848/nacos （默认账号 nacos/nacos；本脚本关闭鉴权亦可匿名）

$ErrorActionPreference = "Stop"

$listening = netstat -ano | Select-String ":8848\s" | Select-String "LISTENING"
if ($listening) {
    Write-Host "Nacos 已在 8848 监听" -ForegroundColor Green
    Write-Host "控制台: http://localhost:8848/nacos"
    exit 0
}

$docker = Get-Command docker -ErrorAction SilentlyContinue
if (-not $docker) {
    Write-Host "未找到 docker。请安装 Docker Desktop 后重试，或手动启动 Nacos。" -ForegroundColor Red
    Write-Host "  docker run -d --name blog-nacos -e MODE=standalone -e NACOS_AUTH_ENABLE=false -e JVM_XMS=256m -e JVM_XMX=256m -p 8848:8848 -p 9848:9848 nacos/nacos-server:v2.3.2"
    exit 1
}

$existing = docker ps -a --filter "name=^blog-nacos$" --format "{{.Names}}" 2>$null
if ($existing -eq "blog-nacos") {
    Write-Host "启动已有容器 blog-nacos ..." -ForegroundColor Cyan
    docker start blog-nacos | Out-Null
} else {
    Write-Host "创建并启动 Nacos 容器 blog-nacos ..." -ForegroundColor Cyan
    docker run -d --name blog-nacos `
        -e MODE=standalone `
        -e NACOS_AUTH_ENABLE=false `
        -e JVM_XMS=256m `
        -e JVM_XMX=256m `
        -p 8848:8848 `
        -p 9848:9848 `
        nacos/nacos-server:v2.3.2 | Out-Null
}

for ($i = 1; $i -le 60; $i++) {
    try {
        $r = Invoke-WebRequest -Uri "http://127.0.0.1:8848/nacos/" -UseBasicParsing -TimeoutSec 2 -ErrorAction Stop
        if ($r.StatusCode -ge 200) {
            Write-Host "Nacos 已就绪 :8848" -ForegroundColor Green
            Write-Host "控制台: http://localhost:8848/nacos （服务管理可见 blog.user / blog.blog / blog.rpg）"
            exit 0
        }
    } catch {
        Start-Sleep -Seconds 2
    }
}
Write-Host "Nacos 容器已启动，但控制台尚未就绪，请稍后打开 http://localhost:8848/nacos" -ForegroundColor Yellow
