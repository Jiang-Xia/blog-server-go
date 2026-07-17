# 启动本机 etcd（学习路径 Kitex 注册中心）
# 用法: .\scripts\start-etcd.ps1
$ErrorActionPreference = "Stop"
$etcdExe = "D:\env\etcd\etcd.exe"
if (-not (Test-Path $etcdExe)) {
    Write-Host "未找到 $etcdExe，请安装 etcd 或使用 Docker:" -ForegroundColor Red
    Write-Host "  docker run -d --name etcd -p 2379:2379 quay.io/coreos/etcd:v3.5.16 /usr/local/bin/etcd --advertise-client-urls=http://0.0.0.0:2379 --listen-client-urls=http://0.0.0.0:2379"
    exit 1
}
$listening = netstat -ano | Select-String ":2379\s" | Select-String "LISTENING"
if ($listening) {
    Write-Host "etcd 已在 2379 监听" -ForegroundColor Green
    exit 0
}
$dataDir = Join-Path $env:TEMP "etcd-blog-data"
New-Item -ItemType Directory -Force -Path $dataDir | Out-Null
Start-Process -FilePath $etcdExe -ArgumentList @(
    "--data-dir=$dataDir",
    "--advertise-client-urls=http://127.0.0.1:2379",
    "--listen-client-urls=http://127.0.0.1:2379"
) -WindowStyle Hidden
Start-Sleep -Seconds 2
$etcdctl = "D:\env\etcd\etcdctl.exe"
if (Test-Path $etcdctl) {
    & $etcdctl endpoint health
} else {
    Write-Host "etcd 已启动 :2379" -ForegroundColor Green
}
