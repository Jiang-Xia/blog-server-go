# Docker 部署（仅本地/CI 全栈）

本地或测试机一键拉起 **4 微服务 + MySQL + Redis + Jaeger**：

```bash
docker compose -f deploy/docker/docker-compose.yml up -d --build
# 或 make up
```

Jaeger UI：`http://localhost:16686`

**2G 生产机**请用 **PM2 + 二进制**，见 [`deploy/pm2/README.md`](../pm2/README.md)（复用宿主机 MySQL/Redis，内存更省）。

可选：部署前在服务器执行 `deploy/docker/setup-swap.sh` 增加 1G swap。
