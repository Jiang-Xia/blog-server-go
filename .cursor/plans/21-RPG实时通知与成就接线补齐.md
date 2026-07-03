# Plan 21：RPG 实时通知与成就接线补齐

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 补齐 Nest `RpgNotifyService` 主导的 WebSocket 事件与成就/任务接线，恢复 C 端庆祝动画与隐藏成就 |
| **前置依赖** | [19-RPG文章等级与Stream消费对齐.md](./19-RPG文章等级与Stream消费对齐.md)、[20-RPG惩罚链与禁言WS对齐.md](./20-RPG惩罚链与禁言WS对齐.md)（文章升级/惩罚 WS 依赖业务结果） |
| **周期** | ~1–1.5 周 |
| **架构形态** | **rpg-service** `notify` + 各玩法 service 埋点；推送经 **blog-service** WS Hub |
| **里程碑** | M7 Nest RPG 玩法债补齐 |
| **Nest 对照** | `modules/rpg/core/rpg-notify.service.ts`、`modules/core/realtime/constants/ws-events.ts` |
| **对等矩阵** | **P-04** |

## 背景

Go `internal/rpg/notify/service.go` 仅实现：

- `levelUp`、`expGain`（8s 防抖）、`tipReceived`、`rechargeComplete`
- `BroadcastActivityUpdate` 在 rpg 独立进程下 **空实现**

Nest 另有十余种 RPG WS 事件（见 `RPG-GAMEPLAY.md` §WebSocket）。同时：

- `LevelService.AddExp` 升级时 **未调用** `achievement.TrackLevelUp`
- 抽奖保底 **未** `TrackProgress(uid, "lottery_pity")`
- 成就完成 **无** `achievementComplete` WS
- 任务完成/领奖 **无** `questComplete` / `questReward` WS

## 模块范围

### A. 扩展 RpgNotifyService

在 `notify/service.go` 增加方法（payload 对齐 Nest `ws-events.ts`）：

| WS type | 触发场景 | 优先级 |
|---------|----------|--------|
| `questComplete` | 任务进度达 target | P0 |
| `questReward` | 领奖成功 | P0 |
| `achievementComplete` | `completeAchievement` | P0 |
| `articleLevelUp` | Plan 19 `AddArticleExp` 升级 | P0 |
| `masterpiece` | Plan 19 神作首次 | P0 |
| `lifeChange` | Plan 20 惩罚扣血 / 社交砸蛋 | P0 |
| `banStatus` | Plan 20 禁言/解封 | P0（解封 Plan 20 可能已做） |
| `shieldUsed` | Plan 20 护盾抵消 | P1 |
| `socialReceived` | cheer/egg/flower 目标用户 | P1 |
| `currencyChange` | 货币变动（打赏/任务/成就发币） | P1 |
| `itemGranted` | 背包新增物品 | P1 |
| `lotteryTicketChange` | 抽奖券变化 | P2 |
| `petHatched` | 宠物孵化 | P2 |
| `buffGranted` / `buffExpired` | Buff 获得/过期 | P2 |
| `rankChange` | 排行榜 TopN 变动（1h dedupe） | P2 |
| `guildEvent` | 公会事件 | P2 |
| `weatherBuff` | WS 连接上下文推送 | P2 |

实现要求：

- 事件名与 Nest **字符串一致**（home-uniapp / home-nuxt 已订阅）
- `expGain` 防抖逻辑保持；新增事件按 Nest 是否防抖对齐
- `activityUpdate`：评估经 Redis pub 广播或 gateway BFF（文档记录限制）

### B. 业务埋点接线

| 包 | 改动 |
|----|------|
| `level/service.go` | 升级后 `achievement.TrackLevelUp` + 已有 `NotifyLevelUp` |
| `achievement/service.go` | `completeAchievement` 后 `NotifyAchievementComplete` |
| `quest/service.go` | 进度完成 / `Claim` 后 notify |
| `lottery/service.go` | 保底触发 `TrackProgress(..., "lottery_pity")` |
| `social/interact.go` | 目标用户 `NotifySocialReceived` + 可选 `lifeChange` |
| `inventory/service.go` | 发物品/货币时 notify（与 Nest 调用点对照） |
| `level/article_level.go` | 升级/神作 notify（Plan 19 交付后接入） |

### C. WS 连接上下文（可选）

Nest `RpgRealtimeConnectListener`：连接时推 `weatherBuff`、`activityUpdate`。  
Go 评估在 blog WS handshake 或 rpg 侧补推；本计划 **至少文档化** 差异。

### D. 前端契约

对照 `blog-home-nuxt` / `blog-home-uniapp` RPG WS 监听表，列出 **必须** 与 **可选** 事件；验收以 H5 冒烟为准（不强制微信开发者工具全量）。

## 任务清单

- [ ] 盘点 Nest `rpg-notify.service.ts` 全部 `notify*` 方法 → Go 缺口表
- [ ] 扩展 `notify/service.go` + payload struct（`internal/rpg/notify/payloads.go`）
- [ ] `level` / `achievement` / `quest` / `lottery` / `social` 埋点
- [ ] Plan 19/20 产生的业务点接入 notify
- [ ] 单元测试：notify 方法 mock `wspush.Pusher` 断言 type/payload
- [ ] `scripts/ws-smoke.mjs` 或扩展现有脚本覆盖 2–3 个新事件
- [ ] 更新 `nest-parity-matrix.md` §3.4 RPG WS 行

## 验收标准

- [ ] 升级除 `levelUp` 外，成就 `level_20` 等可推进（`TrackLevelUp` 已调用）
- [ ] 完成成就时前端可收到 `achievementComplete`（H5 或 ws 冒烟）
- [ ] 任务可领取时收到 `questComplete` 或领奖 `questReward`
- [ ] 文章升级（Plan 19 后）作者收到 `articleLevelUp`
- [ ] 抽奖保底触发 `lottery_pity` 成就进度
- [ ] 社交互动目标用户收到 `socialReceived`（至少 egg/cheer 之一）

## 本计划不做

- ArticleLevel 业务本身 — Plan 19
- Punishment 规则本身 — Plan 20
- Socket.IO 兼容层
- `activityUpdate` 全服广播（若架构不支持，写入已知限制）

## 风险与注意点

| 风险 | 对策 |
|------|------|
| rpg 进程无 WS Hub | 统一 `wspush` → blog Redis channel |
| 事件风暴 | 复用 Nest debounce（expGain/rankChange/weather） |
| 前端未监听某事件 | 以 Nuxt RPG 音频/庆祝组件为准做 P0 列表 |

## 文档交付

[`docs/21-RPG实时通知与成就接线补齐.md`](../../docs/21-RPG实时通知与成就接线补齐.md)

- [ ] 文档已写入 `docs/`

## 下一步

M7 收尾：生产切流 Nest → Go gateway（见 Plan 10 运维 checklist）。§3.6 所列项**不执行**。
