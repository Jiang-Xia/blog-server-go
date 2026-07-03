# Plan 15：RAG 知识库模块

> 对应计划：[`.cursor/plans/15-RAG知识库模块.md`](../.cursor/plans/15-RAG知识库模块.md)
>
> **交付日期**：2026-07-03  
> **架构形态**：4 微服务（RAG HTTP + 索引在 **blog-service**；gateway 代理 `/rag/*`、`/admin/rag/*`）

## 交付摘要

在 blog-service 实现 Nest `RagModule` 对等能力：C 端配额/状态/流式问答（AI SDK UI Message SSE）、admin 五接口、Redis Stream 增量索引、本地 hash Embedding 回退。**Plan 15.1** 补充静态页索引与 blog 内 Tool 规则路由（RPG 实时榜留 Plan 17）。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `services/blog/ent/schema/knowledge_chunk.go` 等 | RAG 三表 Ent schema |
| `services/blog/internal/rag/` | chunk、embedding、hybrid、quota、query、indexer、admin |
| `services/blog/internal/rag/static_page.go` | 静态页 registry + embed Markdown |
| `services/blog/internal/rag/content/pages/*.md` | 特性/RPG 攻略/工具箱说明（与 Nest 同步） |
| `services/blog/internal/rag/tools/` | P1/P2 只读 Tool + 规则路由 orchestrator |
| `services/blog/internal/crossdb/rag_tools.go` | Tool 用跨库 SQL（文章排行/作者/分类/标签/归档） |
| `services/blog/internal/rag/listener/` | Stream 消费 `article.*` / `user.locked` → 增量索引 |
| `services/blog/internal/handler/rag_handler.go` | HTTP 端点 |
| `pkg/config/config.go` | `rag` 配置段 |
| `services/gateway/internal/proxy/router.go` | `rag`、`admin/rag` → blog-service |
| `services/blog/internal/middleware/permission.go` | 开发环境公开 `GET /rag/status` |
| `deploy/postman/rag-smoke.json` | Newman 冒烟 |

## 配置与环境

`configs/blog.yaml` / `blog.example.yaml` 新增 `rag` 段（密钥建议 env 注入）：

```yaml
rag:
  enabled: false
  daily_quota: 20
  top_k: 6
  allow_local_fallback: true
  embedding:
    mode: local
    remote_url: ""
    api_key: ""
    model: "BAAI/bge-large-zh-v1.5"
  llm:
    base_url: "https://api.deepseek.com/v1"
    api_key: ""
    model: "deepseek-chat"
  chunk:
    size: 600
    overlap: 120
```

- `rag.enabled=true` 且配置 LLM Key 后 `/query-stream` 可完整流式回答。
- 未配置远程 Embedding 时自动使用 local hash 向量（与 Nest fallback 一致）。

## 接口一览

### C 端（gateway → blog-service）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/v1/rag/quota` | JWT | 今日配额 used/limit/remaining |
| GET | `/api/v1/rag/status` | 公开 | enabled、chunkCount、embedding 模式 |
| POST | `/api/v1/rag/query-stream` | JWT | SSE：AI SDK UI Message（start → citations → text-delta → finish → [DONE]） |

### Admin

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/admin/rag/stats` | 概览统计 |
| GET | `/api/v1/admin/rag/query-logs` | 查询日志分页 |
| GET | `/api/v1/admin/rag/index-jobs` | 索引任务 |
| GET | `/api/v1/admin/rag/chunks` | 知识块列表 |
| POST | `/api/v1/admin/rag/reindex` | 全量或单篇 `articleId` 重建 |

## 本地启动与验收

```powershell
# 微服务全栈
powershell -ExecutionPolicy Bypass -File scripts/dev-all.ps1

# 单元
go test ./services/blog/internal/rag/... -count=1

# Postman（需 token）
go run scripts/dev_login.go --token-only
newman run deploy/postman/rag-smoke.json --env-var baseUrl=http://127.0.0.1:8000 --env-var token=$TOKEN

# SSE 冒烟（需 rag.enabled + LLM Key）
curl -N -X POST http://127.0.0.1:8000/api/v1/rag/query-stream `
  -H "Authorization: Bearer $TOKEN" `
  -H "Content-Type: application/json" `
  -d '{"messages":[{"role":"user","content":"博客架构是什么？"}]}'
```

验收步骤：admin reindex → `GET /rag/status` chunkCount > 0 → query-stream 返回 citations SSE。

## 与 NestJS 差异

- **Tool 路由**：已实现 P1/P2（文章排行、作者/分类/标签、友链留言、站点导航、归档、我的发文统计）**规则匹配**；未实现 LLM function calling 二次路由。
- **RPG 实时 Tool**（`get_rpg_leaderboard`、`get_my_rpg_status`）：返回友好提示，真实数据待 **Plan 17** rpg gRPC。
- **静态页**：已索引 3 页（特性/RPG 攻略/工具箱）；改 Markdown 后须 admin **全量 reindex**（无 Stream 增量）。
- **MySQL FULLTEXT**：混合检索关键词分当前为应用层 fallback（Nest 优先 FULLTEXT，失败再 fallback）。
- 超配额 HTTP 状态码返回 **429**（bizCode 同步 429）；RAG 未开启/未配置返回 **503**。

## 已知限制与后续

- 生产须配置独立 Embedding API（如 SiliconFlow）+ LLM Key。
- **Plan 17**：RPG 榜 / 我的 RPG 状态 Tool 接 rpg-service gRPC；可选补 LLM function calling。
- 静态页 Markdown 与 Nuxt 页面须人工同步（改 `content/pages/*.md` 后 reindex）。
- `POST /pub/ai-stream` 旧 Pub SSE 仍不在 Go 范围。

## 验收勾选

- [x] 计划内任务清单已全部完成
- [x] 单元测试已通过（chunk / quota / SSE）
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引状态已更新
