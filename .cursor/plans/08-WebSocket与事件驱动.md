# Plan 08：WebSocket 与事件驱动（模块化单体）

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | WebSocket Hub 替代 Socket.IO、Redis Stream 事件总线、Nuxt 前端 WS 改造 |
| **前置依赖** | [07-博客周边.md](./07-博客周边.md) 验收通过 |
| **周期** | ~1.5-2 周 |
| **架构形态** | 模块化单体（`internal/blog/ws/` + `internal/event/`） |
| **里程碑** | M4 实时RPG |
| **原方案章节** | 四（WebSocket 完整方案）、五（5.7 Redis Stream）、8.3 第10周 |

## 模块范围

| 域 | 对应未来服务 | NestJS 路径 |
|----|-------------|-------------|
| WebSocket Hub | blog-service | `core/realtime/` |
| 事件驱动 | 跨域 | `core/events/` |

### 负责能力

- blog 域：`notification` 推送逻辑（CRUD 已在 Plan 04）
- 跨模块：Redis Stream / pub-sub 事件总线

> **不在范围**：RAG 模块 — 后续扩展

## API / 协议

| 路径/协议 | 说明 |
|-----------|------|
| `WS /realtime` | 原生 WebSocket 升级入口 |
| `GET /api/v1/notification/since?seq=` | 断线补漏 |

## WebSocket 方案要点（总方案第四章）

### 12 项坑与对策

| # | 坑 | 对策 |
|---|-----|------|
| 1 | 无自动重连 | 前端指数退避重连 |
| 2 | 无心跳 | 控制帧 ping + 应用层 ping/pong 双管齐下 |
| 3 | 无消息确认 | seq 序号 + 重连后 since 补漏 |
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

### 跨模块推送（单体版 → Plan 10 拆服务后仍用 Redis）

```go
// 任意模块内推送
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
| rpg.tip | rpg（Plan 09） | blog | 加声望、发通知 |
| rpg.levelup | rpg（Plan 09） | blog | WS 推送升级 |
| article.publish | blog | — | 更新统计缓存 |

本阶段实现事件总线框架 + 消费者骨架；RPG 业务事件在 Plan 09 接入。

## 任务清单

### Week 1：WebSocket

- [ ] 实现 Hub + Client（对照总方案 4.2 代码）
- [ ] Hertz WS 升级入口 `/realtime`（总方案 4.3）
- [ ] JWT 身份验证（query token 或 Authorization header）
- [ ] topic 订阅/取消订阅
- [ ] Redis pub/sub 跨模块推送
- [ ] notification since 补漏接口完善

### Week 2：事件 + 前端

- [ ] Redis Stream 事件总线框架
- [ ] 消费者注册机制（article.publish 等 blog 域事件）
- [ ] Nuxt `useWs` composable + 联调
- [ ] blog-admin WS 消费改造（若有）
- [ ] WS 冒烟：连接、心跳、重连、topic、补漏

## 验收标准

- [ ] WS 连接、心跳、重连、topic 订阅正常
- [ ] 断线重连后 notification 补漏正确
- [ ] Redis Stream 框架可发布/消费测试事件
- [ ] 前端 blog-home-nuxt 无 Socket.IO 依赖，改用原生 WS
- [ ] pub/sub 跨模块推送可用（可用测试 handler 验证）

### 可脚本化验收

```bash
# HTTP 补漏接口
curl -sf "http://localhost:5000/api/v1/notification/since?seq=0" -H "Authorization: Bearer $TOKEN"

# WS 需专用脚本或 Playwright；可先验 HTTP + 手工 WS 冒烟清单
# scripts/ws-smoke.mjs（可选，本计划可新增）
```

## 本计划不做

- RPG 业务逻辑（等级/签到/抽奖等）— Plan 09
- 支付宝支付 — Plan 09
- 单体全量回归 — Plan 09 收尾
- 物理拆微服务 — Plan 10
- RAG 模块 — v3 范围外

## 生产切换提示

WS 路径可单独灰度；完整 NestJS 下线在 Plan 09 全量回归后。

## 风险与注意点

| 风险 | 对策 |
|------|------|
| WS 与 HTTP 端口 | 单体阶段同进程；Plan 10 拆后 WS 留 blog-service |
| 前端 WS 改造 | 保留 Socket.IO 开关，灰度切换 composable |
| 双心跳实现 | 严格对照总方案 4.2，写集成测试 |

## 下一步

完成验收后进入 [09-RPG与支付.md](./09-RPG与支付.md)。
