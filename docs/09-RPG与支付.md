# Plan 09：RPG 与支付

> 对应计划：[`.cursor/plans/09-RPG与支付.md`](../.cursor/plans/09-RPG与支付.md)
>
> **交付日期**：2026-06-30  
> **架构形态**：模块化单体（`internal/rpg/` + `internal/pay/` + `internal/blog/scheduler`）

## 交付摘要

在 Go 单体中落地 RPG 全模块与支付宝充值：`/api/v1/rpg/*`、公开主页、`/admin/rpg/*`、`/pay/*`；Ent 层 19 张 `rpg_*`/`pay_order` 表由业务服务消费；RPG 事件经 Redis Stream `rpg-handlers` 消费组驱动经验/任务/成就；升级/打赏/充值完成经 `RpgNotifyService` → WS 推送；每日 08:00 活动通知 cron 对齐 Nest。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `internal/rpg/core/` | RPG 主记录、完整状态 |
| `internal/rpg/level/` | 等级/经验、签到 |
| `internal/rpg/inventory/` | 背包、装扮槽、钻石 |
| `internal/rpg/quest/`、`lottery/`、`pet/` | 任务、抽奖、宠物 |
| `internal/rpg/guild/`、`activity/`、`leaderboard/` | 公会、赛季活动、排行榜 |
| `internal/rpg/social/` | 打赏、加油/砸蛋/送花、声望 |
| `internal/rpg/buff/` | 用户 Buff、天气 Buff |
| `internal/rpg/achievement/` | 成就追踪 |
| `internal/rpg/recharge/` | RPG 充值意向与履约 |
| `internal/rpg/admin/` | 后台列表/部分 CRUD |
| `internal/rpg/profile/` | 公开主页 RPG 展示 |
| `internal/rpg/notify/` | WS 推送（levelUp、expGain、tip、recharge、activity） |
| `internal/rpg/event/` | Stream 消费：blog 事件 → 经验/任务 |
| `internal/rpg/seeds/` | 启动时幂等种子数据 |
| `internal/pay/service/` | smartwalle/alipay SDK、回调、轮询 |
| `internal/pay/repo/` | pay_order CRUD |
| `internal/handler/rpg_*.go`、`pay_*.go` | HTTP 路由 |
| `internal/app/rpg_adapters.go` | handler 接口适配 |

## 配置与环境

`configs/monolith.yaml` 新增 `pay` 段；支持环境变量覆盖：

| 变量 | 说明 |
|------|------|
| `PAY_ALIPAY_APP_ID` | 支付宝应用 ID |
| `PAY_ALIPAY_PRIVATE_KEY` | 应用私钥 |
| `PAY_ALIPAY_PUBLIC_KEY` | 支付宝公钥 |
| `PAY_ALIPAY_NOTIFY_URL` | 异步通知 URL（默认 `/api/v1/pay/notice`） |

沙箱开发：`pay.sandbox: true`。

## 接口一览

### C 端 RPG `/api/v1/rpg/*`

签到、状态、等级奖励、排行榜、禁言、成就、任务、Buff、抽奖、背包/装扮、宠物、活动、天气 Buff、公会、打赏、社交互动、充值创建/查询等（对齐 Nest `RpgController`）。

### 公开主页

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/user/public/:uid` | 公开主页 |
| GET | `/rpg/public/:uid/status` | 公开 RPG 状态 |
| GET | `/rpg/public/status/batch?uids=` | 批量等级徽章 |

### Admin RPG `/api/v1/admin/rpg/*`（JWT）

成就/任务/奖池/用户/统计/物品/活动/公会/打赏/社交日志等列表与部分写操作。

### 支付

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/pay/trade/create` | 支付宝下单 |
| GET | `/pay/trade/query` | 查询交易 |
| POST | `/pay/notice` | 支付宝异步回调 |
| POST | `/pay/order/*` | 管理端订单 CRUD（JWT） |

### WS 事件（服务端 → 客户端）

`levelUp`、`expGain`（8s 防抖）、`tipReceived`、`rechargeComplete`、`activityUpdate`、`rankChange` 等。

## 本地启动与验收

```powershell
cd blog-server-go
$env:CONFIG_PATH="configs/monolith.yaml"
go run ./services/monolith/cmd/main.go

# 登录 token
go run scripts/dev_login.go --token-only

# 健康检查
curl -sf http://localhost:5000/health

# RPG 签到（需 Bearer token）
curl -sf -X POST http://localhost:5000/api/v1/rpg/sign -H "Authorization: Bearer <token>"

# RPG 状态
curl -sf http://localhost:5000/api/v1/rpg/status -H "Authorization: Bearer <token>"

# 编译
go build ./services/monolith/cmd/...
```

支付宝沙箱：配置 `pay` 段后调用 `/rpg/recharge/create` + `/pay/trade/create` + 沙箱支付 + `/pay/notice` 回调验证发钻与 WS `rechargeComplete`。

## 与 NestJS 差异

| 项 | 说明 |
|----|------|
| Admin 写操作 | 部分 CRUD（成就/奖池/素材上传/解封等）返回「待完善」，列表/查询/删任务/调钻已可用 |
| 公开主页 | 收藏/点赞列表经 blog gRPC 分页返回（Plan 14）；文章列表仍为简化实现 |
| 文章等级 | `ArticleLevelService` 未完整接入 Stream 消费 |
| 敏感词惩罚 | `BanGuard` 仅禁言时间判定，HP 扣减完整逻辑待补 |
| 支付宝 | Go 使用 `smartwalle/alipay/v3`，Nest 使用 `alipay-sdk` npm |

## 已知限制与后续

- `deploy/postman/full-regression.json` 全量契约集待补
- Nest 并行 diff 测试待脚本化
- Plan 10：物理拆 4 微服务、OpenTelemetry、生产 docker-compose
- RAG 模块不在 v3 范围

## 验收勾选

- [x] RPG C 端核心路由可编译并注册
- [x] Admin RPG / Pay 路由注册
- [x] RPG Stream 消费组 `rpg-handlers` 启动
- [x] 活动通知 cron 08:00 注册
- [x] 启动时 RPG 种子 SyncAllPredefined
- [x] 支付宝 SDK 集成（需配置密钥后沙箱联调）
- [x] WS levelUp / tip / recharge 推送基础设施
- [ ] Postman 全量回归（待补 collection）
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引已更新
