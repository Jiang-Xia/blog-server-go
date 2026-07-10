# 22 · 单体服务 Nest 对齐补齐

> **交付日期**：2026-07-10  
> **计划**：[`.cursor/plans/22-单体服务Nest对齐补齐.md`](../.cursor/plans/22-单体服务Nest对齐补齐.md)

## 目标

将 `services/monolith`（`:5000`）补齐为 **Nest 替换主入口**（RAG、定时任务、pub/stats 等 Plan 22 范围）。Plan 22 初期自微服务 blog 侧移植模块；**验收后以单体为唯一功能基准**，四微服务保留作架构学习，**不要求反向同步 parity**。

## 部署策略（2026-07-10 定稿）

| 形态 | 端口 | 定位 |
|------|------|------|
| **monolith** | `:5000` | Nest 替换、生产、新功能开发 |
| gateway + 4 服务 | `:8000` 等 | 微服务架构学习（gRPC BFF、多进程） |

## 交付内容

### P22-01 RAG 知识库

- 从 `services/blog/internal/rag/` 移植至 `services/monolith/internal/rag/`
- 路由：`/api/v1/rag/*`、`/api/v1/admin/rag/*`
- Stream 消费者 `ConsumerGroupRAG`（`article.*` / `user.locked` 增量索引）

### P22-02 定时任务

- 移植 `scheduledtask/`、`crossdb/`
- 路由：`/api/v1/scheduled-task/*`
- 8 内置 cron + RPG 活动通知（8:00）独立 job
- Ent schema 对齐 Nest：`scheduled_task` 用 `TimestampMixin`，`scheduled_task_log` 无 mixin
- **5 段 Nest cron 兼容**：`normalizeCronExpr` 自动补秒字段（如 `* * * * *` → `0 * * * * *`）

### P22-03 pub/stats

- `internal/pub/service.go` 读 Ent 真实计数（非 mock）

### P22-04 冒烟

- 脚本：`scripts/monolith-smoke.ps1`（含公开主页 articles、article/statistics 等）
- Go smoke：`TEST_BASE=http://127.0.0.1:5000 go test -tags=smoke ./test/smoke/...`

### P22-05 远程部署（PM2）

- 与四微服务同一套 `deploy/pm2/` 流程（`releases/` + `current` + `pm2 reload`）
- 单体专用：`ecosystem.monolith.config.js`、`BlogGo_Monolith`、`:5000`
- 命令：`make deploy-monolith`（见 [`deploy/pm2/README.md`](../deploy/pm2/README.md)）
- 生产 env 仍用 `deploy/pm2/env.production`，同步生成 `configs/monolith.yaml`

### P22-06 单体后续增强（Plan 22 后）

- 公开主页 `GET /user/public/:uid/articles` 分页/字段对齐 Nest
- `GET /article/statistics` 后台统计大屏全量指标
- RAG：LLM function calling 兜底 + Tool 接 RPG/文章检索；静态页启动自动 reindex

## 验证记录（本机 2026-07-10）

| 接口 | 结果 |
|------|------|
| `GET /health` | ✅ |
| `GET /api/v1/pub/stats` | ✅ 真实计数 |
| `GET /api/v1/rag/status` | ✅ chunkCount=1498 |
| `GET /api/v1/scheduled-task/tasks` | ✅ 8 任务 running |
| `go test -tags=smoke TestSmokeHealthAndPublic` | ✅ |

## 已知限制

| 项 | 说明 |
|----|------|
| RPG seed `isDelete` 列 | 部分 `x_rpg_*` 表与 Ent `TimeMixin` 不一致，种子同步 WARN 不影响主路径 |
| Nest 切流 | 待 staging 并行与前端 `baseUrl` 切换（P22-04 后半） |
| §3.6 明确不做项 | 微信支付、`/pub/ai-stream` 等仍不对齐 |

## 相关文件

- `services/monolith/internal/rag/`
- `services/monolith/internal/scheduledtask/`
- `services/monolith/internal/crossdb/`
- `services/monolith/ent/schema/scheduled_task*.go`
- `scripts/monolith-smoke.ps1`
