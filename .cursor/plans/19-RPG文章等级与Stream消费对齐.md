# Plan 19：RPG 文章等级与 Stream 消费对齐

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 实现 `ArticleLevelService` 等价能力；深化 `rpg-service` 对 `blog:events` 的消费，使文章 `articleExp` / `articleLevel` / `isMasterpiece` 与 Nest 一致增长 |
| **前置依赖** | [18-领域事件发布补齐.md](./18-领域事件发布补齐.md) 验收通过（blog 侧已 publish 全量互动事件） |
| **周期** | ~1–1.5 周 |
| **架构形态** | **rpg-service** 消费 + **blog gRPC** 更新文章表（或 rpg 经 blog 写 article RPG 字段） |
| **里程碑** | M7 Nest RPG 玩法债补齐 |
| **Nest 对照** | `modules/rpg/level/article-level.service.ts` + `listeners/rpg-event.consumer.ts`（`onArticlePublished` / `onCommentCreated` 等） |
| **对等矩阵** | [`docs/nest-parity-matrix.md`](../../docs/nest-parity-matrix.md) **P-02** |

## 背景

Plan 09/18 后：用户侧经验/任务/成就 Stream 链路已通，但 **作者向文章等级体系整块缺失**。

Go `consumer.go` 现状：

- `onArticlePublished`：计算 `initialExp` 后 `_ = initialExp // stub`
- `onCommentCreated` / `onLikeCreated` / `onCollectCreated`：未解析 `authorUid` / `articleId`，不给作者加文章经验
- `onArticleViewed`：仅 Redis dedup，无 `addArticleExp`

Nest 在以上事件均调用 `ArticleLevelService.addArticleExp`，并联动声望、神作判定、作者成就（`article_level_up` / `masterpiece`）。

## 模块范围

### A. ArticleLevelService（rpg-service）

新建 `services/rpg/internal/rpg/level/article_level.go`（或 `articlelevel/` 包），对齐 Nest：

| 方法 | 职责 |
|------|------|
| `AddArticleExp(ctx, articleID, amount, authorUID, reputation?)` | 累加 `articleExp`、连升 `articleLevel`、可选 `reputationGained` + `AddReputation` |
| `AddTipTotal(ctx, articleID, amount)` | 打赏总额 increment（`TipService` 已有部分逻辑可复用/迁移） |
| `checkMasterpiece(article)` | `articleLevel >= MASTERPIECE_LEVEL` 或 `articleExp >= MASTERPIECE_EXP` → `isMasterpiece=1` |

**写库路径（二选一，验收前定案）：**

1. **推荐**：扩展 `proto/blog/v1/article.proto` → `UpdateArticleRPGFields(articleId, articleExp, articleLevel, isMasterpiece, reputationGained, tipTotal)`，由 blog-service Ent 更新 `article` 表（rpg 不直连 blog 库表）
2. 备选：rpg CrossDB 只读/写 article RPG 列（与 Plan 11 边界冲突，非首选）

阈值与常量复用 `internal/rpg/constants/economy.go`，与 Nest `ECONOMY.*` 对齐。

### B. Stream 消费者改造（`internal/rpg/event/consumer.go`）

| 事件 | 改造要点 |
|------|----------|
| `blog.article.published` | 声望加成后 `AddArticleExp(articleId, initialExp, uid, nil)` |
| `blog.comment.created` | 解析 `authorUid`/`articleId`；评论者经验保留；**作者** `ARTICLE_COMMENT_EXP` + 声望 |
| `blog.like.created` | 同上 `ARTICLE_LIKE_EXP` |
| `blog.collect.created` | 同上 `ARTICLE_COLLECT_EXP` |
| `blog.article.viewed` | dedup 后 `AddArticleExp(..., ARTICLE_VIEW_EXP, authorUid, viewReputation)` |

payload 解析须用 Plan 18 `authorUid` / `articleId` 字段（comment/like/collect 已发布）。

### C. 成就触发（本计划最小集）

在 `AddArticleExp` 内或 consumer 回调：

- 文章升级 → `achievement.TrackProgress(authorUID, "article_level_up")`（增量 1）
- 首次神作 → `TrackProgress(authorUID, "masterpiece")`

> WS `articleLevelUp` / `masterpiece` 推送留给 [21-RPG实时通知与成就接线补齐.md](./21-RPG实时通知与成就接线补齐.md)。

### D. monolith 副本

若 `services/monolith/internal/rpg/event/consumer.go` 仍保留，同步改造或标记 deprecated。

## 任务清单

- [ ] 设计并落地 article RPG 字段更新 gRPC（或选定写库方案）
- [ ] 实现 `ArticleLevelService`：`AddArticleExp` / `AddTipTotal` / 神作判定
- [ ] 改造 `onArticlePublished` / comment / like / collect / viewed 五个 handler
- [ ] `onArticleTipped` 确认 `tipTotal` 与 Nest 一致（如需迁到 ArticleLevelService）
- [ ] 单元测试：升级阈值、神作边界、声望可选参数
- [ ] 集成测试：发文 → 查 article `articleExp`；评论 → 作者文章 exp 增加
- [ ] 更新 `nest-parity-matrix.md` P-02 状态

## 验收标准

- [ ] 发布文章后 `article.articleExp` / `articleLevel` 按声望加成写入（非 0/默认）
- [ ] 他人评论/点赞/收藏后 **作者** 文章 exp 增加（非评论者）
- [ ] 阅读文章（同日同访客 dedup 后）作者文章 exp +1 路径与 Nest 一致
- [ ] 达标后 `isMasterpiece=1`；作者成就 `article_level_up` / `masterpiece` 可推进
- [ ] 用户侧经验/任务（Plan 18）行为不退化

### 可脚本化验收

```powershell
make dev-all
$TOKEN = go run scripts/dev_login.go --token-only

# 发文（或编辑为 publish）
curl -sf -X POST http://127.0.0.1:8000/api/v1/article/create `
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" `
  -d '{"title":"plan19-test","content":"x","status":"publish","categoryId":"..."}'

# 查文章详情 articleExp / articleLevel（C 端或 admin）
curl -sf "http://127.0.0.1:8000/api/v1/article/info?id=<id>"

# 另一用户评论后复查作者文章 exp
go test -tags=integration ./test/integration/... -run ArticleLevel -count=1
```

## 本计划不做

- `PunishmentService` 全链 — [20-RPG惩罚链与禁言WS对齐.md](./20-RPG惩罚链与禁言WS对齐.md)
- RPG WebSocket `articleLevelUp` / `masterpiece` — Plan 21
- 公开主页 `articles` 分页 — **明确不做**（见 `nest-parity-matrix.md` §3.6）
- 微信支付

## 风险与注意点

| 风险 | 对策 |
|------|------|
| rpg 写 article 表越界 | 优先 blog gRPC 更新 RPG 列 |
| 跨服务最终一致 | 与 Nest 相同：Stream 异步，允许秒级延迟 |
| 重复事件导致 exp 重复 | 与 Nest 语义一致；必要时按 eventId 幂等（可选） |

## 文档交付

[`docs/19-RPG文章等级与Stream消费对齐.md`](../../docs/19-RPG文章等级与Stream消费对齐.md)

- [ ] 文档已写入 `docs/`

## 下一步

[20-RPG惩罚链与禁言WS对齐.md](./20-RPG惩罚链与禁言WS对齐.md)
