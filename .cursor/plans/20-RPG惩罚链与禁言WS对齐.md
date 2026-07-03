# Plan 20：RPG 惩罚链与禁言 WS 对齐

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 将 `PunishmentService` 对齐 Nest：敏感词命中扣 HP、护盾、累计/归零自动禁言；admin 解封推送 WS |
| **前置依赖** | [18-领域事件发布补齐.md](./18-领域事件发布补齐.md)（`blog.sensitive-word.hit` 已发布）；[13-RPG后台补全与社区禁言联动.md](./13-RPG后台补全与社区禁言联动.md)（BanGuard + `AssertNotBanned` 已有） |
| **周期** | ~1 周 |
| **架构形态** | **rpg-service**（`punishment` + `buff` + Stream consumer + notify） |
| **里程碑** | M7 Nest RPG 玩法债补齐 |
| **Nest 对照** | `modules/rpg/core/punishment.service.ts`、`buff.service.ts` `hasShield` |
| **对等矩阵** | **P-01**、**P-03** |

## 背景

Plan 13 交付了 **BanGuard 拦截**（已禁言用户不能评论/签到），但 Plan 13 文档已知限制写明：

- 敏感词 **HP 扣减 / 自动禁言** 完整惩罚链未对齐 Nest
- admin 解封 **无 WS `banStatus`**

Go 现状：

- `event/consumer.go` `onSensitiveWordHit`：仅 `lifeValue -= penalty`、`sensitiveHitsCount++`
- `punishment/service.go`：仅 `GetBanStatus` / `AssertNotBanned` / `AdminUnban`（清 DB，无 WS）
- **无** `buff.hasShield`

Nest `onSensitiveWordHit` → `punishmentService.onSensitiveWordHit(uid, hpPenalty)` 完整链见 `punishment.service.ts`。

## 模块范围

### A. PunishmentService 扩展（`internal/rpg/punishment/service.go`）

实现 `OnSensitiveWordHit(ctx, uid, hpPenalty)`，对齐 Nest 常量：

| 规则 | Nest 常量 | 行为 |
|------|-----------|------|
| 单次扣血 | `LIFE_DEDUCT_PER_HIT` / payload `hpPenalty` | `lifeValue = max(0, life - penalty)` |
| 护盾 | `buffService.hasShield` | 免疫扣血，`sensitiveHitsCount++`，通知 `shieldUsed` |
| 累计 5 次命中 | `HIT_COUNT_BAN_THRESHOLD` | 禁言 72h |
| 生命归零 | `ZERO_LIFE_*` | 临时 24h 或连续 3 次正式 30 天；归零后 `lifeValue` 重置 100 |
| 禁言字段 | `banStartTime` / `banEndTime` | 写入 `x_rpg` |

Stream consumer **改为调用** `PunishmentService.OnSensitiveWordHit`，删除 consumer 内联扣血逻辑。

### B. Buff 护盾（`internal/rpg/buff/service.go`）

- 新增 `HasShield(ctx, uid) (bool, error)`，对齐 Nest 激活中护盾 Buff 判定
- 护盾抵消时：`RpgNotifyService.NotifyShieldUsed`（WS 可 Plan 21 一并接 notify 方法）

### C. Admin 解封 WS

`AdminUnban` 成功后：

- 推送 `banStatus`：`{ banned: false, banEndTime: null, banReason: null }`
- 经 `wspush.Pusher` → blog-service `/realtime`（现有 rpg→blog WS 推送路径）

### D. 社交互动 HP（可选本计划内）

Nest `RpgNotifyService.notifySocialReceived` 在砸蛋等场景推 `lifeChange`。  
`social/interact.go` 已改 `target.LifeValue`；本计划至少保证 **数据一致**；WS `socialReceived`/`lifeChange` 可拆 Plan 21。

## 任务清单

- [ ] `BuffService.HasShield` 实现 + 单元测试
- [ ] `PunishmentService.OnSensitiveWordHit` 全链 port + 单元测试（护盾/累计禁言/归零禁言）
- [ ] `consumer.onSensitiveWordHit` 委托 PunishmentService
- [ ] `AdminUnban` 推送 `banStatus` WS
- [ ] 集成测试：发敏感词评论 → `lifeValue` 下降；累计命中 → `banEndTime` 有值 → BanGuard 拦截
- [ ] monolith 副本同步（若保留）
- [ ] 更新 `nest-parity-matrix.md` P-01、P-03

## 验收标准

- [ ] 登录用户评论命中敏感词（`hpPenalty>0`）→ `lifeValue` 按配置下降
- [ ] 有护盾 Buff 时扣血为 0，`sensitiveHitsCount` 仍累加
- [ ] 累计命中达阈值 → 自动禁言 → `POST /comment/create` 被 BanGuard 拒绝（文案与 Nest 一致）
- [ ] admin `POST /admin/rpg/users/:uid/unban` 后用户可再评论；在线用户收到 `banStatus`（若 WS 冒烟覆盖）
- [ ] 不再在 `consumer.go` 内联惩罚逻辑

### 可脚本化验收

```powershell
make dev-all
$TOKEN = go run scripts/dev_login.go --token-only

# 需库中有会扣 HP 的敏感词种子；发评论触发 hit 事件
curl -sf -X POST http://127.0.0.1:8000/api/v1/comment/create `
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" `
  -d '{"articleId":1,"content":"<测试敏感词>"}'

curl -sf http://127.0.0.1:8000/api/v1/rpg/status -H "Authorization: Bearer $TOKEN"
# 期望 lifeValue 下降

go test ./services/rpg/internal/rpg/punishment/... ./services/rpg/internal/rpg/buff/... -count=1
```

## 本计划不做

- 文章等级 — Plan 19
- 全量 RPG WS 事件表 — Plan 21
- 敏感词审核联动（Plan 13 已完成）
- 微信支付

## 风险与注意点

| 风险 | 对策 |
|------|------|
| 微服务 rpg 推 WS 需经 blog | 复用现有 `wspush` / Redis pub 路径 |
| 与 Nest 禁言时长边界 | 单测固定 clock 或对比 Nest 常量表 |
| 测试环境无敏感词种子 | 文档注明 seed SQL 或 dev 配置 |

## 文档交付

[`docs/20-RPG惩罚链与禁言WS对齐.md`](../../docs/20-RPG惩罚链与禁言WS对齐.md)

- [ ] 文档已写入 `docs/`

## 下一步

[21-RPG实时通知与成就接线补齐.md](./21-RPG实时通知与成就接线补齐.md)
