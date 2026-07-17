# monolith（线上主入口 · 对接 uniapp）

> **Plan 22 起**：`services/monolith`（本地 / 生产均为 **`:8000`**）为 **Nest 替换、本地开发、冒烟与线上部署的唯一基准**。新功能只在此落地；**不要求**同步至四微服务。

## 用途

- `make dev` / `.\scripts\dev.ps1` 本地开发与联调（`services/monolith/cmd/main.go`）
- **线上部署**：配合 `blog-home-uniapp`（及 nuxt / admin）大前端，替换 Nest `blog-server`
- 集成测试 / E2E 全栈验收（`scripts/monolith-smoke.ps1`）

## 与微服务关系

| 域 | 微服务目录 | monolith 状态 |
|----|-----------|---------------|
| user | `services/user/` | 进程内 `internal/user/`，**以单体为准** |
| blog | `services/blog/` | 进程内 `internal/blog/` + RAG + 定时任务 |
| rpg + pay | `services/rpg/` | 进程内 `internal/rpg/` + `internal/pay/` |
| gateway | `services/gateway/` | 单体无 gateway；BFF 能力进程内等价 |

**代码不共用**：Go `internal` 规则下 monolith **不能** import `services/*/internal/*`，故与微服务为双份实现。微服务目录（Plan 10–11）**仅本地 WSL 学习**（gRPC、gateway、多进程），**不上生产**，**不强制功能 parity**。

## 远程部署（PM2）

仅部署本服务；见 [`deploy/pm2/`](../../deploy/pm2/)。

```powershell
cp deploy/pm2/deploy.monolith.local.env.example deploy/pm2/deploy.monolith.local.env
# 填写 SSH；env.production 与 Nest 同格式
make deploy-monolith
```

| 项 | 值 |
|----|-----|
| PM2 名 | `BlogGo_Monolith` |
| 端口 | `:8000`（本地 dev / 生产 PM2；Nest 仍 `:5000`，可并行） |
| ecosystem | `ecosystem.monolith.config.js` |
| 健康检查 | `curl http://127.0.0.1:8000/api/v1/health` |

详见 [`deploy/pm2/README.md`](../../deploy/pm2/README.md)。

## 本地启动

```bash
make dev              # 单体 :8000
# API baseUrl: http://127.0.0.1:8000/api/v1
```

四微服务（**可选 · 仅 WSL 学习**，与单体 `:8000` 二选一）：

```bash
make dev-all          # gateway :8000
make dev-all-stop
```
