# Plan 20：RPG 惩罚链与禁言 WS 对齐

> 对应计划：[`.cursor/plans/20-RPG惩罚链与禁言WS对齐.md`](../.cursor/plans/20-RPG惩罚链与禁言WS对齐.md)
>
> **交付日期**：2026-07-06  
> **架构形态**：rpg-service（`punishment` + `buff` + Stream consumer + `wspush`）

## 交付摘要

将 `PunishmentService.onSensitiveWordHit` 对齐 Nest：扣 HP、护盾消耗、`sensitiveHitsCount` 累计禁言（72h）、生命归零临时/正式禁言并重置 HP；Stream consumer 不再内联扣血。`AdminUnban` 成功后经 `RpgNotifyService` 推送 `banStatus` WS（微服务经 Redis `realtime:push` → blog Hub）。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `services/rpg/internal/rpg/punishment/` | 常量、`applySensitiveWordHit`、`OnSensitiveWordHit`、`AdminUnban` WS |
| `services/rpg/internal/rpg/buff/service.go` | `HasShield` + `consumeUse` |
| `services/rpg/internal/rpg/notify/service.go` | `NotifyShieldUsed` / `NotifyLifeChange` / `NotifyBanStatus` |
| `services/rpg/internal/rpg/event/consumer.go` | `onSensitiveWordHit` 委托 `PunishmentService` |
| `services/rpg/internal/rpg/repo/rpg_repo.go` | `DeleteBuffByID` |
| `services/monolith/internal/rpg/**` | 单体副本同步 |

## 惩罚规则（对齐 Nest 常量）

| 规则 | 常量 | 行为 |
|------|------|------|
| 单次扣血 | `LifeDeductPerHit=20` / payload `hpPenalty` | `lifeValue = max(0, life - penalty)` |
| 护盾 | `buff.HasShield` | 免疫扣血，仍 `sensitiveHitsCount++`，WS `shieldUsed` |
| 累计 5 次 | `HitCountBanThreshold` | 禁言 72h |
| 生命归零 | `ZeroLifeTempBanHours=24` / 连续 3 次 30 天 | 归零后 `lifeValue` 重置 100 |
| admin 解封 | — | 清 DB + WS `{ banned: false, banEndTime: null, banReason: null }` |

## 本地启动与验收

```powershell
make dev-all

# 单元测试
go test ./services/rpg/internal/rpg/punishment/... ./services/rpg/internal/rpg/buff/... ./services/rpg/internal/rpg/notify/... -count=1

# 集成（需全栈 + 库中有 hpPenalty>0 的敏感词，否则 Skip）
go test -tags=integration ./test/integration/... -run SensitiveWordPunishment -count=1

# 手动：发含敏感词评论后查 RPG 状态
$TOKEN = go run scripts/dev_login.go --token-only
curl -sf -X POST http://127.0.0.1:8000/api/v1/comment/create `
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" `
  -d '{"articleId":1,"content":"<测试敏感词>"}'
curl -sf http://127.0.0.1:8000/api/v1/rpg/status -H "Authorization: Bearer $TOKEN"
```

## 与 NestJS 差异

| 项 | Nest | Go |
|----|------|-----|
| 惩罚入口 | comment 服务进程内调用 | blog 发 Stream → rpg consumer |
| WS 推送 | `RealtimeGateway` 本进程 | rpg `wspush.RedisPusher` → blog Hub |
| 社交砸蛋 `lifeChange` WS | 有 | **数据已写库**；WS 留给 Plan 21 |
| 其余 RPG WS 事件表 | 全量 | Plan 21 |

## 已知限制与后续

- **Plan 21**：成就/任务 Stream 接线、其余 RPG WS 事件 debounce/enrich
- 集成测试依赖环境敏感词种子；无匹配词时 `t.Skip`
- 测试账号若已禁言，集成用例会 Skip

## 验收勾选

- [x] `BuffService.HasShield` + 单元测试
- [x] `PunishmentService.OnSensitiveWordHit` 全链 + 单元测试
- [x] `consumer.onSensitiveWordHit` 委托 PunishmentService
- [x] `AdminUnban` 推送 `banStatus` WS
- [x] 集成测试脚本（`TestIntegrationSensitiveWordPunishment`）
- [x] monolith 副本同步
- [x] 更新 `nest-parity-matrix.md` P-01、P-03
