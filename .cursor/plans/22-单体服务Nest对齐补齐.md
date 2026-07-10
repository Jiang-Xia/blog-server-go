# Plan 22 · 单体服务 Nest 对齐补齐

> **目标**：让 `services/monolith`（`:5000`）具备 Nest 替换能力，最终下线 `blog-server`（NestJS）。四微服务 + gateway 保留作 **微服务架构学习**，不强制与单体功能 parity。
>
> **前置**：Plan 01–21 已在微服务侧验收；单体当前冻结于 Plan 11 基线 + 部分 RPG 19–21 进程内实现。
>
> **对照文档**：[`docs/nest-parity-matrix.md`](../../docs/nest-parity-matrix.md)、[`docs/api-routes.md`](../../docs/api-routes.md)

## 背景与差距结论

### 架构现状（2026-07-10 定稿）

| 部署 | 端口 | 定位 | Nest 对齐度 |
|------|------|------|-------------|
| **monolith** | `:5000` | **Nest 替换主入口**（Plan 22） | ✅ 对等基准（§3.6 除外） |
| gateway + user/blog/rpg | `:8000` 等 | **微服务架构学习** | ⚠️ 可能落后于单体，非 parity 目标 |
| blog-server Nest | `:5000` | 待替换 | 基准 |

### 单体 vs 微服务：缺失功能清单（Plan 22 前 · 历史）

> **2026-07-10**：下表为 Plan 22 启动前差距；G-01～G-05 已在单体补齐，见 [`docs/22-单体服务Nest对齐补齐.md`](../../docs/22-单体服务Nest对齐补齐.md)。**不要求**反向同步至微服务。

| # | 能力 | 微服务落点 | 单体现状（Plan 22 前） | 阻塞切 Nest |
|---|------|-----------|----------|-------------|
| G-01 | **RAG 知识库**（C 端 SSE 问答 + admin 索引/日志） | `services/blog/internal/rag/` | Ent 表有，**无 handler/索引/SSE** | **是**（home 已接 `/rag/*`） |
| G-02 | **定时任务 admin**（`/scheduled-task/*` CRUD/触发/日志/备份） | `services/blog/internal/scheduledtask/` | **无路由** | **是**（admin 运维） |
| G-03 | **8 内置 cron job**（定时发布、DB 备份、权限缓存、统计 token、邮件汇总等） | Plan 12 | 仅 placeholder + 活动通知 | **是**（scheduled_publish） |
| G-04 | **`/pub/stats` 真实统计** | gateway BFF gRPC 聚合 | **mock 硬编码**（128/12/36） | 中 |
| G-05 | **RAG Stream 索引消费者** | `blog/internal/rag/listener` | 无 | 随 G-01 |
| G-06 | **跨进程 WS Redis push** | `rpg/internal/wspush` | 进程内 WS（**单体更优，非缺失**） | 否 |
| G-07 | **Gateway BFF**（article/info、public profile 聚合） | `gateway/internal/aggregator` | 直连 handler（**等价，非缺失**） | 否 |
| G-08 | **gRPC 跨服务协作** | user/blog/rpg proto | 进程内调用（**等价，非缺失**） | 否 |

### 单体 vs Nest：仍存在的共同差距（🚫 不做）

> 微服务与单体均不对齐，**不纳入 Plan 22**；见 `nest-parity-matrix.md` §3.6。

- 微信支付 JSAPI/小程序
- `/pub/ai-stream` Pub SSE 代理
- 公开主页 `articles` 完整分页（简化列表）
- RAG LLM function calling / Tool 接 rpg gRPC
- Gateway 全局限流、K8s、`article.published` blog 统计缓存消费者

### 单体 vs Nest：部分对齐（⚠️ 可选增强）

| 能力 | 说明 | Plan 22 处理 |
|------|------|-------------|
| 文章 statistics/trends | 部分指标简化 | P22-04 验收时对比 Nest 响应，按需补字段 |
| RAG Tool | 规则路由，非 LLM 二次路由 | 与微服务一致即可 |
| 支付宝 | ✅ 已有 | 回归 smoke |

### 单体已有、无需从微服务移植

- Plan 02–11 全部 REST（user/blog/rpg/pay/admin）
- Plan 17 协作逻辑（进程内 `ListActiveUserIDs`、敏感词过滤、dept 数据权限）
- Plan 18 领域事件发布 + rpg Stream 消费
- Plan 19–21 RPG（ArticleLevel、Punishment 链、WS P0 通知）
- Plan 13 BanGuard + 敏感词 hit → 实体 status 同步
- Plan 14 公开主页 collects/likes
- Plan 16 `/resources/baidutongji`

---

## 实施策略

### 推荐：「单体复主 + pkg 抽公共逻辑」

Go `internal` 规则禁止 monolith import `services/*/internal/*`。避免长期双份维护：

1. **抽公共**：将 RAG、scheduledtask 的可复用核心迁至 `pkg/rag`、`pkg/scheduledtask`（或 `internal/shared/` 若仅 monolith 用则放 monolith 并从 blog **复制**后单向同步）。
2. **Pragmatic 首期**：从 `services/blog/internal/{rag,scheduledtask}/` **复制**到 `services/monolith/internal/`，改 import 为 monolith 路径，wire 装配；验收后再决定是否上抽 pkg。
3. **撤销 deprecated**：Plan 22 验收通过后 `:5000` 作为 Nest 替换入口；微服务保留为 **架构学习** 部署，**不要求**反向同步 parity。

### 不推荐

- 以微服务为功能开发主路径再手工 cherry-pick 到单体（已导致历史缺口）
- 未补齐单体直接切 Nest（RAG + 定时发布会断）
- 强制维持单体 ↔ 微服务双向 parity（微服务仅学习用）

---

## 执行阶段

### P22-01 · RAG 模块移植（~2 周）

**来源**：`services/blog/internal/rag/`、`handler/rag_handler.go`、`register_blog.go` RAG 段

**任务**：

- [ ] 复制/迁移 `internal/rag/{indexer,query,tools,listener,module.go}` 至 monolith
- [ ] 注册路由：`GET /rag/quota|status`、`POST /rag/query-stream`；admin `/admin/rag/*`
- [ ] wire 注入：`wire.go` + `providers.go`
- [ ] RBAC：`middleware/permission.go` 公开路径
- [ ] Stream listener 订阅 `article.*` / 互动事件 → 增量索引
- [ ] 单元测试：`rag/tools/*_test.go` 迁或复跑
- [ ] Swagger：`make swag-monolith`（若有）或 `swag-all`

**验收**：

```powershell
make dev   # :5000
newman run deploy/postman/rag-smoke.json --env-var baseUrl=http://127.0.0.1:5000
go test ./services/monolith/internal/rag/... -count=1
```

### P22-02 · 定时任务与 cron（~2 周）

**来源**：`services/blog/internal/scheduledtask/`、`register_blog.go` scheduled-task 段

**任务**：

- [ ] 迁移 scheduledtask 模块（repo/service/runner/admin handler）
- [ ] 注册 `/scheduled-task/*` 全套 admin API
- [ ] 替换 monolith `scheduler/`：注册 8 内置 job（对齐 Plan 12 清单）
- [ ] `scheduled_publish` 与 Plan 18 事件发布联动
- [ ] crossdb 查询（锁定用户过滤、备份路径）— monolith 可直连 Ent，无需 gRPC
- [ ] 种子数据：内置 task 行幂等 INSERT

**8 内置 job 清单**（对齐 Nest ScheduledTaskModule）：

1. `scheduled_publish` — 定时发文
2. `db_backup` — 数据库备份
3. `clear_permissions_cache` — RBAC Redis 刷新
4. `refresh_tongji_token` — 百度统计 token
5. `daily_activity_notify` — 活动通知（单体已有，合并）
6. `email_digest` — 邮件汇总（若 Nest 有）
7. `clean_operation_log` — 操作日志清理
8. `clean_scheduled_task_log` — 任务日志清理

**验收**：

```powershell
newman run deploy/postman/scheduled-task-smoke.json --env-var baseUrl=http://127.0.0.1:5000
# 手动：scheduled_publish 到点发文 + article.published 事件
```

### P22-03 · pub/stats 与质量债（~3–5 天）

**任务**：

- [ ] `internal/pub/service.go`：读 blog Ent 真实 article/category/tag count（去掉 mock）
- [ ] 对比 Nest `/pub/stats` 响应字段
- [ ] 更新 `docs/api-routes.md` §6（修正「路由并集」表述，列明 Plan 22 后对齐）

**验收**：`GET /api/v1/pub/stats` 计数与 DB 一致。

### P22-04 · 全量回归与 Nest 切流（~1–2 周）

**任务**：

- [ ] 扩展 Postman smoke：`baseUrl=http://127.0.0.1:5000` 跑现有 `deploy/postman/*-smoke.json`
- [ ] 四层测试：`.\scripts\test-run.ps1` 增加/确认 monolith 模式（或 `-MonolithOnly` flag）
- [ ] 与 Nest 并行 1–2 周：`blog-home-nuxt` / `blog-admin` 指向 Go 单体 staging
- [ ] PM2 / docker-compose：新增 `blog-server-go-monolith` profile，端口 5000
- [ ] 切流 checklist：Redis `DEL public_api_paths api_permission_mappings`、JWT 密钥一致、静态文件路径、WS `/realtime`
- [ ] 下线 Nest：停 `blog-server`，文档更新

**验收**：

- [ ] admin 定时任务页、RAG 助手、RPG 全流程、支付回调
- [ ] 无 P0 接口 4xx/5xx 回归

---

## 里程碑与时间

| 阶段 | 内容 | 周期 | 累计 |
|------|------|------|------|
| P22-01 | RAG 移植 | ~2 周 | 2 周 |
| P22-02 | 定时任务 | ~2 周 | 4 周 |
| P22-03 | pub/stats + 文档 | ~3–5 天 | ~4.5 周 |
| P22-04 | 回归 + 切 Nest | ~1–2 周 | **~6 周** |

P22-01 与 P22-02 可各由一人并行（不同目录），但 **wire 冲突需串行合并**。

---

## 切 Nest 前后端配置

| 项目 | 变更 |
|------|------|
| `blog-home-nuxt` | `NUXT_PUBLIC_API_BASE` → Go 单体 URL |
| `blog-admin` | API base → `:5000/api/v1` |
| `blog-home-uniapp` | `VITE_API_BASE` 对齐 |
| PM2 / nginx | upstream 5000 指向 blog-server-go monolith |

---

## 交付文档（Plan 22 完成后）

- [ ] `docs/22-单体服务Nest对齐补齐.md`
- [ ] 更新 `docs/README.md` 索引
- [ ] 更新 `docs/nest-parity-matrix.md`：单体列状态、§6 monolith 章节
- [ ] 更新 `services/monolith/README.md`：移除 deprecated 或改为「Nest 替换入口」

---

## 风险

| 风险 | 缓解 |
|------|------|
| 双份代码再次漂移 | 以 monolith 为唯一功能基准；微服务不强制同步；必要时抽 `pkg/` |
| RAG 依赖外部 embedding API | 与微服务共用 config；staging 先关 RAG 开关 |
| 定时任务 DB 备份路径 Windows/Linux | 沿用 Plan 12 配置项，部署文档写明 |
| 切流 JWT 不兼容 | 共用 `configs/` RSA 密钥与 Nest 一致 |

---

## 快速差距自检（当前）

```powershell
# 单体缺路由（应 404）
curl -s -o NUL -w "%{http_code}" http://127.0.0.1:5000/api/v1/rag/status
curl -s -o NUL -w "%{http_code}" http://127.0.0.1:5000/api/v1/scheduled-task/tasks

# 微服务应有（gateway）
curl -s -o NUL -w "%{http_code}" http://127.0.0.1:8000/api/v1/rag/status
```
