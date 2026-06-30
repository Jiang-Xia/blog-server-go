# Plan 08：WebSocket 与事件驱动

> 对应计划：[`.cursor/plans/08-WebSocket与事件驱动.md`](../.cursor/plans/08-WebSocket与事件驱动.md)
>
> **交付日期**：2026-06-30  
> **架构形态**：模块化单体（`internal/blog/ws/` + `internal/event/`）

## 交付摘要

实现原生 WebSocket Hub 替代 Nest Socket.IO：`GET /realtime` 升级、双重心跳、topic 订阅、Redis pub/sub 跨模块推送、站内通知 WS 推送与 HTTP since 断线补漏；Redis Stream `blog:events` 发布/消费框架；blog-home-nuxt 默认原生 WS composable（可 `VITE_NUXT_USE_SOCKET_IO=true` 回退 Nest）。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `internal/blog/ws/hub.go` | 连接管理、topic 路由、消息分发 |
| `internal/blog/ws/client.go` | ReadPump / WritePump、双重心跳 |
| `internal/blog/ws/pusher.go` | 本进程推送 + Redis pub/sub |
| `internal/handler/ws_handler.go` | JWT 鉴权、gorilla WS 升级 |
| `internal/event/publisher.go` | Stream XADD 发布 |
| `internal/event/consumer.go` | XREADGROUP 消费、幂等、XACK |
| `internal/event/blog_handlers.go` | `article.published` 消费者骨架 |
| `internal/blog/notification/service.go` | Create 后 WS 推送 siteNotification |
| `scripts/ws-smoke.mjs` | HTTP + WS 可脚本化验收 |

## 配置与环境

无新增 YAML 项；沿用 `configs/monolith.yaml` 的 Redis/JWT/HTTP。

前端（blog-home-nuxt）：

| 变量 | 说明 |
|------|------|
| `VITE_NUXT_USE_SOCKET_IO=true` | 回退 Socket.IO（连 Nest）；默认不设置即用原生 WS |

## 接口一览

### WebSocket

| 协议 | 路径 | 说明 |
|------|------|------|
| WS | `GET /realtime?token=` | JWT 鉴权（query token 或 Authorization Bearer） |

**客户端 → 服务端**

| type | data | 说明 |
|------|------|------|
| `ping` | — | 应用层心跳 |
| `subscribe` | `{ "topic": "..." }` | 订阅 topic |
| `unsubscribe` | `{ "topic": "..." }` | 取消订阅 |

**服务端 → 客户端**

```json
{ "type": "siteNotification", "seq": 123, "data": { "notification": {...}, "unreadCount": 1 } }
```

- `seq`：站内通知用 `site_notification.id`，供断线补漏
- 跨模块推送：任意模块 `Publish(realtime:push, RedisPushMsg)`

### HTTP

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/notification/since?seq=` | 断线补漏（id > seq） |

### 开发冒烟（`app.env=development`）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/dev/ws-push` | Hub 直推当前用户 |
| POST | `/api/v1/dev/ws-push-redis` | Redis pub/sub 推送 |
| POST | `/api/v1/dev/event-publish` | 向 `blog:events` 发布测试事件 |

## Redis 约定

| Key/Channel | 类型 | 说明 |
|-------------|------|------|
| `realtime:push` | pub/sub | 跨模块 WS 推送 |
| `blog:events` | Stream | 领域事件（与 Nest 一致） |
| `blog-handlers` | Consumer Group | blog 域消费者 |
| `blog:event:done:{id}` | String NX | 幂等标记，TTL 7 天 |

## 本地启动与验收

```bash
# 启动
set CONFIG_PATH=configs/monolith.yaml
go run ./services/monolith/cmd/main.go

# 获取 token
go run scripts/dev_login.go --token-only

# HTTP 补漏
curl -sf "http://localhost:5000/api/v1/notification/since?seq=0" -H "Authorization: Bearer $TOKEN"

# 全量 WS 冒烟
node scripts/ws-smoke.mjs
```

## 与 NestJS 差异

| 项 | 说明 |
|----|------|
| 协议 | Socket.IO namespace → 原生 WS JSON `{type,data,seq}` |
| 路径 | 同为 `/realtime`；Go 为 HTTP 升级，Nest 为 Socket.IO |
| RPG 推送 | Plan 09 接入；当前仅 siteNotification + dev 测试推送 |
| Stream 消费 | blog 域仅 `article.published` 骨架；RPG 消费组 Plan 09 |

## 已知限制与后续

- RPG 业务 WS 事件（levelUp 等）在 Plan 09 接入
- `article.published` 消费者暂只打 debug 日志，统计缓存刷新待补
- blog-admin 无 WS 消费，无需改造
- 生产 Origin 校验：`ws_handler` CheckOrigin 当前为 `true`，上线需收紧

## 验收勾选

- [x] WS 连接、心跳、重连、topic 订阅
- [x] 断线重连 notification since 补漏（HTTP + 前端 replay）
- [x] Redis Stream 框架可发布/消费测试事件
- [x] blog-home-nuxt 默认原生 WS（Socket.IO 可选回退）
- [x] pub/sub 跨模块推送（`/dev/ws-push-redis`）
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引已更新
