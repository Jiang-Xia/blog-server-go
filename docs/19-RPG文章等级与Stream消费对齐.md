# Plan 19：RPG 文章等级与 Stream 消费对齐

> 对应计划：[`.cursor/plans/19-RPG文章等级与Stream消费对齐.md`](../.cursor/plans/19-RPG文章等级与Stream消费对齐.md)
>
> **交付日期**：2026-07-06  
> **架构形态**：4 微服务（rpg-service 消费 `blog:events` → blog gRPC 写 `x_article` RPG 列）

## 交付摘要

实现 Nest `ArticleLevelService` 等价能力：`rpg-service` Stream 消费者在发文/评论/点赞/收藏/浏览事件后累加作者文章 `articleExp`、连升 `articleLevel`、判定 `isMasterpiece`，并推进作者成就 `article_level_up` / `masterpiece`。打赏流程补齐 `tipTotal` 原子累加。文章详情 HTTP 接口现返回 RPG 字段供验收。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `proto/blog/v1/article.proto` | 新增 `GetArticleRPGFields` / `UpdateArticleRPGFields` / `AddArticleTipTotal` |
| `pkg/blogsvc/article_rpg.go` | rpg → blog gRPC 客户端 |
| `services/blog/internal/blog/grpcserver/server.go` | blog gRPC 实现（Ent 写 `x_article`） |
| `services/rpg/internal/rpg/level/article_level.go` | `ArticleLevelService` |
| `services/rpg/internal/rpg/event/consumer.go` | 五个 handler 接入文章等级 |
| `services/rpg/internal/rpg/social/tip.go` | 打赏后 `AddTipTotal` |
| `services/monolith/internal/rpg/level/*` | 单体 Ent 读写副本 |
| `services/rpg/internal/rpg/seeds/seeds.go` | 作者成就种子 `first_masterpiece` 等 |
| `services/blog/internal/blog/domain/types.go` | 详情 DTO 暴露 RPG 字段 |

## 配置与环境

无新增环境变量。rpg-service 沿用 `grpc.blog_addr` 连接 blog-service（与敏感词审核、公开主页列表相同）。

## 接口一览

| 类型 | 说明 |
|------|------|
| gRPC `blog.v1.ArticleService/GetArticleRPGFields` | rpg 读文章 RPG 快照 |
| gRPC `blog.v1.ArticleService/UpdateArticleRPGFields` | rpg 写 `articleExp` / `articleLevel` / `isMasterpiece` / `reputationGained` |
| gRPC `blog.v1.ArticleService/AddArticleTipTotal` | 打赏总额 increment |
| REST `GET /api/v1/article/info` | 响应 `info.articleExp` / `articleLevel` / `isMasterpiece` / `tipTotal` |

无新增对外 REST；Stream 消费为内部副作用。

## 本地启动与验收

```powershell
# 微服务全栈
make dev-all

# 单元测试（升级阈值、神作边界）
go test ./services/rpg/internal/rpg/level/... -count=1

# 集成（需 MySQL/Redis + 全栈运行）
go test -tags=integration ./test/integration/... -run ArticleLevel -count=1

# 手动：发文后查详情
$TOKEN = go run scripts/dev_login.go --token-only
curl -sf "http://127.0.0.1:8000/api/v1/article/info?id=<id>" -H "Authorization: Bearer $TOKEN"
# info.articleExp 应 > 0（声望加成后约 10）
```

## 与 NestJS 差异

| 项 | Nest | Go |
|----|------|-----|
| 写库路径 | 同进程 TypeORM | rpg gRPC → blog Ent（微服务）/ Ent 直连（单体） |
| 文章等级阈值公式 | `level*(level-1)*20` | 一致 |
| 发文声望 | `reputation=null` 不加作者声望 | 一致（`reputationSkip=true`） |
| 浏览 dedup | `viewerUid \|\| 'anon'` | 匿名 `viewerUid=0`（Plan 18 未传 visitor） |
| WS `articleLevelUp` / `masterpiece` | 有 | **留给 Plan 21** |
| 事件幂等 | Redis eventId | 沿用现有 consumer 语义，未新增 eventId 幂等 |

## 已知限制与后续

- **Plan 20**：`PunishmentService` 全链、admin 解封 WS
- **Plan 21**：`articleLevelUp` / `masterpiece` WebSocket 推送
- 公开主页 `articles` 分页 — 仍明确不做（`nest-parity-matrix.md` §3.6）
- `blog.article.viewed` 未传 `viewerUid` 时匿名 dedup 粒度较粗

## 验收勾选

- [x] 计划内任务清单已全部完成
- [x] 单元测试已通过；集成测试脚本已添加（`TestIntegrationArticleLevelComment`）
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引状态已更新
