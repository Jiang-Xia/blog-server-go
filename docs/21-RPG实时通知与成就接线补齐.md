# Plan 21：RPG 实时通知与成就接线补齐

> 对应计划：[`.cursor/plans/21-RPG实时通知与成就接线补齐.md`](../.cursor/plans/21-RPG实时通知与成就接线补齐.md)
>
> **交付日期**：2026-07-06  
> **架构形态**：rpg-service `notify` + 各玩法 service 埋点；微服务经 `wspush` → blog Redis `realtime:push` → WS Hub

## 交付摘要

补齐 Nest `RpgNotifyService` 主导的 RPG WebSocket 事件与成就/任务/文章/社交/背包埋点，恢复 C 端庆祝动画与隐藏成就（如 `lottery_pity`）。P0 事件已全部实现并接线；P2 连接上下文（`weatherBuff` / connect `activityUpdate`）与部分调用点（guild/rank/buff 过期）保留 notify 方法，业务接线见「已知限制」。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `services/rpg/internal/rpg/notify/payloads.go` | WS 事件名 + payload struct（对齐 Nest `ws-events.ts`） |
| `services/rpg/internal/rpg/notify/service.go` | `RpgNotifyService` 全量 push 方法 + expGain 防抖 |
| `services/rpg/internal/rpg/constants/display.go` | 稀有度/来源中文 enrich |
| `level/service.go` | 升级后 `TrackLevelUp` |
| `achievement/service.go` | `completeAchievement` → `achievementComplete` |
| `quest/service.go` | 完成 → `questComplete`；领奖 → `questReward` |
| `lottery/service.go` | 保底 → `lottery_pity`；`AddTickets` → `lotteryTicketChange` |
| `social/interact.go` | 目标用户 `socialReceived` + 附带 `lifeChange` |
| `level/article_level.go` | `articleLevelUp` / `masterpiece` |
| `inventory/service.go` | `currencyChange` / `itemGranted` |
| `pet/service.go` | `petHatched` |
| `services/monolith/internal/rpg/**` | 单体副本同步（monolith 保留 `BroadcastActivityUpdate` 在线广播） |

## WS 事件覆盖（P0 已接线）

| type | 触发点 |
|------|--------|
| `questComplete` | `QuestService.TrackProgress` 首次达标 |
| `questReward` | `QuestService.ClaimReward` |
| `achievementComplete` | `AchievementService.completeAchievement` |
| `articleLevelUp` | `ArticleLevelService.AddArticleExp` 升级 |
| `masterpiece` | 首次神作 |
| `socialReceived` | `InteractService` cheer/egg/flower |
| `lifeChange` | 社交 hpDelta≠0（含 `lifeRecovered`） |
| `currencyChange` | `InventoryService.AdjustCurrency` |
| `itemGranted` | `InventoryService.GrantItem`（跳过 level_up / buff 卷轴） |
| `lotteryTicketChange` | `LotteryService.AddTickets` |
| `petHatched` | `PetService.Summon` |
| `levelUp` / `expGain` / `tipReceived` / `rechargeComplete` | 既有 |
| `banStatus` / `shieldUsed` | Plan 20 |

## 本地验证

```powershell
# 单元测试（notify mock Pusher）
go test ./services/rpg/internal/rpg/notify/... -count=1

# 全量 PR 级
.\scripts\test-run.ps1 -UnitOnly

# WS 冒烟（需 make dev 单体 :5000）
node scripts/ws-smoke.mjs
```

## 与 NestJS 差异

| 项 | Nest | Go |
|----|------|-----|
| 推送路径 | 进程内 `RealtimeGateway` | rpg `wspush.RedisPusher` → blog Hub |
| `activityUpdate` 全服广播 | cron + 在线 uid | **rpg-service 空实现**；monolith 仍广播 |
| `pushConnectContext` | WS 连接推 `weatherBuff` + 活动快照 | **未接线**；`NotifyWeatherBuff` 方法已备 |
| `rankChange` / `guildEvent` / `buffExpired` | leaderboard/guild/buff 调用 | notify 方法已实现；**业务调用点待补** |
| `buffGranted` | Buff 授予时 | notify 方法已实现；buff service 未调用 |

## 已知限制

- rpg 独立进程无 WS Hub：`BroadcastActivityUpdate` 仍为 no-op（文档化，不阻塞 M7）
- 连接时 `weatherBuff` / connect `activityUpdate` 需在 blog WS handshake 或 BFF 层评估（本计划仅文档化）
- `tipReceived` 可选字段 `articleTitle` / `balance` 由 TipService 后续 enrich 可再补

## 验收勾选

- [x] 升级调用 `TrackLevelUp`（等级成就如 level_20）
- [x] 成就完成推送 `achievementComplete`
- [x] 任务完成/领奖推送 `questComplete` / `questReward`
- [x] 文章升级/神作推送 `articleLevelUp` / `masterpiece`
- [x] 抽奖保底 `TrackProgress(..., "lottery_pity")`
- [x] 社交目标用户 `socialReceived`（含 egg/cheer lifeChange）
- [x] notify 单元测试 + `ws-smoke.mjs` RPG 事件段
- [x] 更新 `nest-parity-matrix.md` P-04
- [x] monolith 副本同步
