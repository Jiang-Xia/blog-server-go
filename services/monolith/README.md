# monolith（Nest 替换主入口）

> **Plan 22 起**：`services/monolith`（本地 dev **`:8000`**）为 **Nest 替换、本地开发、冒烟与生产部署的唯一基准**。新功能只在此落地；**不要求**同步至四微服务。

## 用途

- `make dev` / `.\scripts\dev.ps1` 本地开发与联调（`services/monolith/cmd/main.go`）
- **替换 `blog-server`（NestJS）** 的生产部署目标
- 集成测试 / E2E 全栈验收（`scripts/monolith-smoke.ps1`）

## 与微服务关系

| 域 | 微服务目录 | monolith 状态 |
|----|-----------|---------------|
| user | `services/user/` | 进程内 `internal/user/`，**以单体为准** |
| blog | `services/blog/` | 进程内 `internal/blog/` + RAG + 定时任务 |
| rpg + pay | `services/rpg/` | 进程内 `internal/rpg/` + `internal/pay/` |
| gateway | `services/gateway/` | 单体无 gateway；BFF 能力进程内等价 |

**Go `internal` 规则**：monolith **不能** import `services/*/internal/*`，故与微服务存在代码双份。微服务目录（Plan 10–11 产物）保留用于 **学习微服务架构**（gRPC、gateway 代理、多进程部署），**不强制功能 parity**，亦不作为 Nest 对等验收依据。

## 远程部署（PM2）

与四微服务共用 [`deploy/pm2/`](../../deploy/pm2/) 流程（`env.production` → `configs/monolith.yaml` → 交叉编译 → tar → SSH → `pm2 reload`）。

```powershell
cp deploy/pm2/deploy.monolith.local.env.example deploy/pm2/deploy.monolith.local.env
# 填写 SSH；env.production 与 Nest 同格式
make deploy-monolith
```

| 项 | 值 |
|----|-----|
| PM2 名 | `BlogGo_Monolith` |
| 端口 | `:8000`（本地 dev / 生产 PM2，与 Nest `:5000` 分离） |
| ecosystem | `ecosystem.monolith.config.js` |
| 健康检查 | `curl http://127.0.0.1:8000/api/v1/health` |

详见 [`deploy/pm2/README.md`](../../deploy/pm2/README.md)。

## 本地启动

```bash
make dev              # 单体 :8000
# API baseUrl: http://127.0.0.1:8000/api/v1
```

四微服务联调（**可选 · 架构学习**）：

```bash
make dev-all          # gateway :8000
make dev-all-stop
```
