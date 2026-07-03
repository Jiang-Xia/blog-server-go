# Plan 15：RAG 知识库模块

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 迁移 Nest `RagModule`：C 端流式问答 + admin 索引/日志 + Redis Stream 增量索引 |
| **前置依赖** | [13-RPG后台补全与社区禁言联动.md](./13-RPG后台补全与社区禁言联动.md) 验收通过（quota 与 RPG 用户体系稳定）；[14-公开主页收藏与点赞列表.md](./14-公开主页收藏与点赞列表.md) 可并行 |
| **周期** | ~3–4 周 |
| **架构形态** | 4 微服务（建议 **blog-service** 承载 RAG HTTP + 索引；表在 blog Ent；gateway 代理 `/rag/*`、`/admin/rag/*`） |
| **里程碑** | M6 Nest 差异补齐 |
| **Nest 对照** | `blog-server/src/modules/rag/` |

## 背景

v3 方案明确 RAG「后续扩展」。home 工具箱 AI 摘要 / 站内 RAG 问答依赖：

- `GET /rag/quota`、`GET /rag/status`
- `POST /rag/query-stream`（AI SDK UI Message SSE）
- `GET/POST /admin/rag/*`（stats、query-logs、index-jobs、chunks、reindex）

monolith Ent 已有 `rag_query_log` 等 schema 片段，但 **无运行时模块**。

## 模块范围

| Go 包（目标） | Nest 对照 |
|---------------|-----------|
| `services/blog/internal/rag/` | `rag/` 根 |
| `.../rag/query/` | `rag-query.service.ts` |
| `.../rag/indexer/` | `indexer/` |
| `.../rag/embedding/` | `embedding/` |
| `.../rag/vector/` | `vector/` hybrid + vector search |
| `.../rag/quota/` | `quota/`（Redis 日配额） |
| `.../rag/admin/` | `admin/` |
| `.../rag/listener/` | `listeners/rag-event.consumer`（Stream 消费文章变更） |

### 负责表（blog-service Ent）

- `knowledge_chunk`（或现网表名，对照 entimport）
- `rag_index_job`
- `rag_query_log`

> 验收前对照 MySQL 现网表名与 Nest TypeORM entity，必要时仅 ent schema 改名不改表。

## API 路径

### C 端 `/api/v1/rag`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/quota` | JWT | 今日问答配额 |
| GET | `/status` | 公开 | enabled、chunkCount、embedding 模式 |
| POST | `/query-stream` | JWT | SSE：`text/event-stream`，AI SDK UI Message 格式 |

SSE 响应头（对齐 Nest）：

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`
- `x-vercel-ai-ui-message-stream: v1`

### Admin `/api/v1/admin/rag`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/stats` | 概览统计 |
| GET | `/query-logs` | 查询日志分页 |
| GET | `/index-jobs` | 索引任务 |
| GET | `/chunks` | 知识块列表 |
| POST | `/reindex` | 全量或单篇 `articleId` 重建 |

### Gateway

新增前缀：`rag`、`admin/rag`（`admin/rag` 优先于 `admin/` user 路由）→ **blog-service**。

## 配置（`configs/*.yaml` 新增 `rag` 段）

对照 Nest `Config.ragConfig` / embedding：

| 配置项 | 说明 |
|--------|------|
| `rag.enabled` | 总开关 |
| `rag.daily_quota` | 每用户日配额 |
| `rag.embedding.mode` | `local` / `remote` |
| `rag.embedding.remote_url` | OpenAI 兼容 embedding API |
| `rag.embedding.api_key` | 密钥（env 注入） |
| `rag.llm.base_url` / `api_key` / `model` | 对话模型 |
| `rag.chunk.size` / `overlap` | 分块参数 |

未配置 embedding/llm 时：`/status` 返回 `configured=false`；`/query-stream` 返回与 Nest 一致的业务错误。

## 关键实现要点

### 索引流水线

1. **ChunkService**：Markdown/HTML  strip → 分块
2. **EmbeddingService**：本地 hash embedding（Nest fallback）或远程 API
3. **VectorSearch + HybridSearch**：MySQL 存向量 JSON 或 float 列（与现网一致）；关键词 + 向量混合
4. **RagIndexerService**：全量/单篇 reindex；写 `rag_index_job`
5. **RagEventConsumer**：消费 `blog.article.published` / 更新事件 → 异步 reindex

### 流式问答

- 对齐 Nest `writeUiMessageSse` 事件序列：`start` → `data-citations` → `text-start` → `text-delta`* → `text-end` → `finish` → `[DONE]`
- 配额：`quota` Redis key `rag:quota:{uid}:{date}`；首包成功前 `consume`
- 每次请求写 `rag_query_log`（question、answer_preview、citations JSON、latency、status）

### Admin

- blog-admin RAG 管理页对接上述 5 个接口
- reindex 异步 job（goroutine + job 表状态），避免 HTTP 超时

### RPG 联动（若有）

Nest `RagModule` imports `RpgModule` 用于 quota 与用户信息；Go 通过 user gRPC `GetUser` 即可。

## 任务清单

- [ ] Ent schema 对齐现网三表 + migrate 如需
- [ ] embedding + chunk + indexer MVP（先支持 article 正文）
- [ ] hybrid search + query prepare
- [ ] `query-stream` SSE handler（Hertz 流式写）
- [ ] quota Redis
- [ ] admin 5 接口
- [ ] Stream consumer 增量索引
- [ ] gateway 路由 + permission fallback 公开 `/rag/status`
- [ ] `deploy/postman/rag-smoke.json`
- [ ] 单元测试：chunk、quota、SSE 帧格式
- [ ] 集成测试：reindex 1 篇文章 → query 返回 citations（mock LLM 可选）

## 验收标准

- [ ] `GET /rag/status` 在 enabled 且 reindex 后有 chunkCount > 0
- [ ] 登录用户 `POST /rag/query-stream` 返回 SSE，前端 AI 组件可渲染
- [ ] 超配额返回 429（与 Nest 一致）
- [ ] admin reindex 触发 job，chunks 列表可见
- [ ] admin query-logs 可查刚才问答记录
- [ ] 发布/更新文章后 Stream 触发增量索引（日志可观测）
- [ ] 未配置 LLM 时错误信息友好，不 panic

### 可脚本化验收

```bash
make dev-all
newman run deploy/postman/rag-smoke.json \
  --env-var baseUrl=http://127.0.0.1:8000 \
  --env-var token=$USER_TOKEN

# SSE 冒烟（需 curl 支持 -N）
curl -N -X POST http://127.0.0.1:8000/api/v1/rag/query-stream \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"博客架构是什么？"}]}'
```

## 本计划不做

- `POST /pub/ai-stream`（旧 Pub SSE，home 若已切 RAG 可不迁）
- 独立 `rag-service` 第五进程（2G 机器优先 blog-service 内聚）
- 向量数据库 Milvus/Pinecone（保持 MySQL 方案与 Nest 一致）

## 风险与注意点

| 风险 | 对策 |
|------|------|
| LLM/embedding 外部依赖 | 集成测试 mock；文档写清 env |
| SSE 与 Hertz | 参考 Nest 超时；禁用反向代理 buffering（nginx 文档） |
| 大文章 reindex CPU | 异步 job + admin 进度 |
| 表结构不一致 | 先 `ent describe` 对照 Nest 生产库 |

## 文档交付

- [`docs/15-RAG知识库模块.md`](../../docs/15-RAG知识库模块.md)
- 根工作区 `feature-doc-sync`：更新 `blog-server/docs/feature-backlog.md` 若 RAG 从待办移除
- home README / `blog-home-nuxt/api/` 若封装变更

- [ ] 文档已写入 `docs/`

## 下一步

[16-百度统计代理.md](./16-百度统计代理.md)
