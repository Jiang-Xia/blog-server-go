# Plan 04：实时通信与 RPG/支付

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | WebSocket 替代 Socket.IO、Redis Stream 事件驱动、RPG + 支付宝全模块、单体全量回归 |
| **前置依赖** | [03-博客内容域.md](./03-博客内容域.md) 验收通过 |
| **周期** | ~3 周 |
| **架构形态** | 模块化单体（`internal/blog/ws/` + `internal/rpg/`） |
| **原方案章节** | 四（WebSocket 完整方案）、五（5.7/5.8）、8.3 第10-12周 |

## 模块范围

| 域 | 对应未来服务 | NestJS 路径 |
|----|-------------|-------------|
| WebSocket Hub | blog-service | `core/realtime/` |
| 事件驱动 | 跨域 | `core/events/` |
| RPG | rpg-service | `modules/rpg/` |
| 支付 | rpg-service | `modules/pay/` |

### 负责表

- blog 域：`notification`（推送逻辑）
- rpg 域：`rpg_*` 全部、`pay_order`

> **不在范围**：RAG 模块（`modules/rag/`）— 后续扩展

## API / 协议

| 路径/协议 | 说明 |
|-----------|------|
| `WS /realtime` | 原生 WebSocket 升级入口 |
| `/api/v1/rpg/*` | RPG 全部接口 |
| `/api/v1/admin/rpg/*` | RPG 后台管理 |
| `/api/v1/pay/*` | 支付宝充值 |

## WebSocket 方案要点（总方案第四章）

### 12 项坑与对策

| # | 坑 | 对策 |
|---|-----|------|
| 1 | 无自动重连 | 前端指数退避重连 |
| 2 | 无心跳 | 控制帧 ping + 应用层 ping/pong 双管齐下 |
| 3 | 无消息确认 | seq 序号 + 重连后 `GET /notification/since?seq=` 补漏 |
| 4 | 无 room | topic 订阅映射 |
| 5 | 并发写 panic | 单一 writeLoop + channel |
| 6-8 | 超时/泄漏 | ReadDeadline/WriteDeadline + CloseHandler |
| 9 | 大消息 | SetReadLimit(4096) |
| 10 | send 满 | select default 踢慢客户端 |
| 11 | 优雅关闭 | hub.Close() |
| 12 | 跨实例 | Redis pub/sub 广播 |

### 核心组件

- `internal/blog/ws/hub.go` — 连接管理、topic 路由、Redis pub/sub 订阅
- `internal/blog/ws/client.go` — ReadPump / WritePump
- `internal/blog/handler/ws.go` — Hertz 升级入口，JWT 验身份

### 跨模块推送（单体版 → Plan 05 拆服务后仍用 Redis）

```go
// rpg 模块内：升级推送
s.rds.Publish(ctx, "realtime:push", payload)
// blog WS Hub 订阅 realtime:push 并转发给客户端
```

## 前端改造（Nuxt）

总方案 4.4：`composables/useWs.ts` ~100 行

| Socket.IO | 原生 WS |
|-----------|---------|
| `io.connect('/realtime')` | `new WebSocket(url?token=...)` |
| 自动重连 | 指数退避 + jitter |
| `socket.on('event')` | `{type, data}` JSON 路由 |
| `socket.join('room')` | `{"type":"subscribe","data":{"topic":"..."}}` |
| 断线补漏 | 重连后拉 `/api/v1/notification/since?seq=` |

需同步改造：`blog-home-nuxt`、`blog-admin`（若有 WS 消费）

## Redis Stream 事件驱动

总方案 5.7：替代 NestJS EventEmitter

| 事件 | 发布方 | 消费方 | 动作 |
|------|--------|--------|------|
| rpg.tip | rpg | blog | 加声望、发通知 |
| rpg.levelup | rpg | blog | WS 推送升级 |
| article.publish | blog | — | 更新统计缓存 |

## 任务清单

### Week 1：WebSocket

- [ ] 实现 Hub + Client（对照总方案 4.2 代码）
- [ ] Hertz WS 升级入口 `/realtime`（总方案 4.3）
- [ ] JWT 身份验证（query token 或 Authorization header）
- [ ] topic 订阅/取消订阅
- [ ] Redis pub/sub 跨模块推送
- [ ] notification since 补漏接口
- [ ] Nuxt `useWs` composable + 联调

### Week 2：RPG 模块

- [ ] 等级/经验系统
- [ ] 签到
- [ ] 任务系统
- [ ] 背包/宠物
- [ ] 抽奖
- [ ] 公会/赛季
- [ ] 打赏
- [ ] admin RPG 管理接口
- [ ] RPG 事件 → Redis Stream → notification/WS

### Week 3：支付 + 全量回归

- [ ] 支付宝 SDK（smartwalle/alipay）
- [ ] 充值下单/回调/订单查询
- [ ] pay_order 表 CRUD
- [ ] cron 完善：签到重置、赛季结算等
- [ ] 全量回归：C 端 + admin + WS + RPG + 支付
- [ ] Postman/newman 契约测试全量跑通
- [ ] 与 NestJS 并行对比测试（关键接口响应 diff）

## 验收标准

- [ ] WS 连接、心跳、重连、topic 订阅正常
- [ ] RPG 升级/打赏可收到 WS 推送
- [ ] 断线重连后 notification 补漏正确
- [ ] 支付宝沙箱充值链路通过
- [ ] 单体全功能回归通过（对照 NestJS 功能清单）
- [ ] 前端 blog-home-nuxt 无 Socket.IO 依赖，改用原生 WS

## 生产切换提示

本计划完成后，**整个 Go 单体可替代 NestJS**：

- Nginx 全量切流到 Go 单体（`:5000` 或自定义端口）
- NestJS 只读观察 1 周，确认无问题后下线
- Plan 05 是在 Go 单体内部拆 4 服务，不再涉及 NestJS

## 风险与注意点

| 风险 | 对策 |
|------|------|
| WS 与 HTTP 端口 | 单体阶段同进程；Plan 05 拆后 WS 留 blog-service |
| 支付宝回调 URL | 沙箱/生产 config 分离 |
| RPG 逻辑复杂 | 逐模块对照 NestJS service 迁移，写集成测试 |
| 前端 WS 改造 | 保留 Socket.IO 开关，灰度切换 composable |

## 下一步

完成验收后进入 [05-微服务拆分与生产上线.md](./05-微服务拆分与生产上线.md)。
