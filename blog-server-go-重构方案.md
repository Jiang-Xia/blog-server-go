# Blog-Server Go 重构架构与技术落地方案（v3 · 4 服务学习版）

> 源项目：[Jiang-Xia/blog-server](https://github.com/Jiang-Xia/blog-server)（NestJS 11 + TypeORM + MySQL 8 + Redis 7，闭源）
> 目标：用 **Go + Hertz** 重构为 4 服务微服务架构（gateway / blog / user / rpg），**共享 MySQL 单库**，原生 WebSocket 替代 Socket.IO，保持 API 兼容与前端无感切换。
> 定位：重构 + 微服务学习实践，生产可跑在 2G 服务器。
> 版本：v3（2026-06-30 定稿——4 服务、共享库、单体优先渐进拆分）

---

## 目录

- [一、架构总览](#一架构总览)
- [二、技术选型](#二技术选型)
- [三、服务拆分设计](#三服务拆分设计)
- [四、原生 WebSocket 完整方案](#四原生-websocket-完整方案)
- [五、关键技术落地](#五关键技术落地)
- [六、目录结构](#六目录结构)
- [七、2G 服务器部署优化](#七2g-服务器部署优化)
- [八、迁移策略与路线图](#八迁移策略与路线图)
- [九、测试与可观测性](#九测试与可观测性)
- [十、风险与决策点](#十风险与决策点)

---

## 一、架构总览

### 1.1 架构图

```
                          ┌──────────────┐
                          │   Nginx/CDN   │
                          └──────┬───────┘
                                 │
                          ┌──────▼───────┐
                          │  api-gateway  │  Hertz :8000
                          │  鉴权/限流/路由/BFF 聚合
                          └──┬───┬───┬───┘
                             │   │   │
                    ┌────────┘   │   └────────┐
                    │            │            │
             ┌──────▼─────┐ ┌────▼─────┐ ┌────▼──────┐
             │blog-service│ │user-svc  │ │rpg-service│
             │ :5001      │ │ :5002    │ │ :5003     │
             │ gRPC       │ │ gRPC     │ │ gRPC      │
             │            │ │          │ │           │
             │ 文章/分类   │ │ 用户/角色 │ │ 签到/任务  │
             │ 评论/点赞   │ │ 认证/JWT │ │ 背包/抽奖  │
             │ 收藏/留言   │ │ 验证码   │ │ 公会/赛季  │
             │ 文件/友链   │ │ 部门/菜单│ │ 打赏/充值  │
             └──────┬─────┘ └────┬─────┘ └─────┬─────┘
                    │            │             │
                    │  WebSocket Hub 集成在     │
                    │  blog-service 或 gateway  │
                    │            │             │
             ┌──────▼────────────▼─────────────▼─────┐
             │        共享基础设施                      │
             │  ┌──────────┐  ┌──────────┐            │
             │  │ MySQL 8  │  │ Redis 7  │            │
             │  │ 单库 blog │  │ 缓存/事件 │            │
             │  └──────────┘  └──────────┘            │
             └────────────────────────────────────────┘
```

### 1.2 核心设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 服务数量 | **4 个**（gateway + blog + user + rpg） | 学习微服务 + 2G 机器可承载 |
| 数据库 | **共享 MySQL 单库** | 学习重点在服务拆分/通信，分库增加复杂度但收益不大 |
| 事件驱动 | **Redis Stream** | 已有 Redis，不引入 NATS，减少中间件 |
| WebSocket | 集成在 **blog-service** | WS 消息多与博客/通知相关，避免单独一个服务 |
| 内部通信 | **gRPC** | 学习微服务核心技能，强类型契约 |
| 对外接口 | **REST**（gateway 暴露） | 前端无感切换，保持 `/api/v1/*` |
| 部署 | **docker-compose** | 2G 机器不上 K8s，4 服务 + MySQL + Redis |

### 1.3 内存预估（2G 机器）

| 组件 | 内存 |
|------|------|
| gateway | 30~50MB |
| blog-service（含 WS） | 60~90MB |
| user-service | 40~60MB |
| rpg-service | 60~90MB |
| MySQL 8（调优后） | 300~450MB |
| Redis 7 | 30~60MB |
| Docker daemon | 100~150MB |
| 系统 | 100~150MB |
| **合计** | **~720MB ~ 1.1GB** |
| **剩余余量** | **~900MB ~ 1.28GB** |

**能跑，做好第七章的优化即可。**

---

## 二、技术选型

### 2.1 核心技术栈

| 层次 | 选型 | 理由 |
|------|------|------|
| 语言 | **Go 1.23+** | 泛型成熟，稳定 |
| Web 框架（对外 REST） | **Hertz** ✅ | 高性能、sonic JSON、非反射路由 |
| RPC（内部通信） | **gRPC** + protobuf | 强类型契约、高性能、微服务标配 |
| IDL | **protobuf** + buf | 管理 proto 定义与代码生成 |
| ORM | **Ent** | 代码生成、强类型、Schema 即文档 |
| 数据库 | MySQL 8.0 | 共享单库，保持与现网一致 |
| 迁移工具 | **golang-migrate** | 版本化迁移 |
| Redis 客户端 | **rueidis** | 高性能、pipeline |
| 事件驱动 | **Redis Stream** | 复用已有 Redis，不引入新中间件 |
| 配置 | **Viper** | 多源配置 |
| 日志 | **zap** | 结构化、零分配 |
| JWT | **golang-jwt/v5** | 社区标准 |
| 校验 | **go-playground/validator** | struct tag |
| WebSocket | **gorilla/websocket** | 事实标准，资料最全 |
| 邮件 | **gomail** | 替代 Nodemailer |
| 支付宝 | **smartwalle/alipay** | 成熟 SDK |
| 定时任务 | **robfig/cron/v3** | 替代 @nestjs/schedule |
| 依赖注入 | **wire** | 编译期生成，零反射 |
| 链路追踪 | **OpenTelemetry** | 导出 Jaeger |
| 指标 | **Prometheus** client | /metrics |
| 测试 | **testify** + **testcontainers-go** | 单元 + 集成 |
| 容器 | Docker multi-stage + distroless | 镜像最小化 |

### 2.2 与原项目依赖对照

| 原依赖（NestJS/TS） | Go 替代 |
|---------------------|---------|
| @nestjs/core + Express | Hertz |
| @nestjs/typeorm / typeorm | Ent |
| ioredis | rueidis |
| @nestjs/jwt + passport-jwt | golang-jwt/v5 + 中间件 |
| passport-local / passport-github2 | 自写 + golang.org/x/oauth2 |
| crypto-js + node-jsencrypt | bcrypt（兼容旧哈希）+ crypto/rsa |
| svg-captcha | mojocn/base64Captcha |
| @nestjs/swagger | swaggo/swag |
| class-validator | go-playground/validator |
| @nestjs/schedule | robfig/cron/v3 |
| socket.io + redis-adapter | gorilla/websocket + Redis pub/sub |
| nodemailer | gomail |
| alipay-sdk | smartwalle/alipay |
| multer | Hertz multipart |
| ua-parser-js | mileusna/useragent |
| marked | yuin/goldmark |

---

## 三、服务拆分设计

### 3.1 服务职责

| 服务 | 端口 | 职责 | 原模块对应 |
|------|------|------|-----------|
| **api-gateway** | 8000 | 统一入口、JWT 验签、限流、路由转发、BFF 响应聚合 | 无（新增） |
| **user-service** | 5002 | 用户、角色、权限、部门、菜单、认证（JWT/GitHub OAuth/验证码）、敏感词 | security/auth, security/captcha, features/user, admin/system, features/sensitive-word |
| **blog-service** | 5001 | 文章、分类、标签、评论、回复、点赞、收藏、留言板、友链、资源、文件、邮件、通知、操作日志、**WebSocket Hub**、定时任务 | features/article, category, tag, comment, reply, like, collect, msgboard, link, resources, file, email, notification, operation-log, scheduled-task, core/realtime |
| **rpg-service** | 5003 | 等级/经验、签到、任务、背包、宠物、抽奖、公会、赛季、打赏、支付宝充值 | pay, rpg |

> WebSocket 放 blog-service 的原因：WS 推送内容主要是文章通知、评论提醒、RPG 升级等，blog-service 是消息最密集的服务，减少跨服务推送。rpg-service 需要推送时通过 Redis Stream 发事件给 blog-service 的 WS Hub 转发。

### 3.2 服务间通信

| 场景 | 方式 | 示例 |
|------|------|------|
| 客户端 → 服务 | REST | `GET /api/v1/article/1` 到 gateway |
| gateway → 内部服务 | **gRPC** | gateway 调 blog-service.GetArticle |
| 同步跨服务调用 | **gRPC** | blog 显示作者信息 → 调 user-service.GetUser |
| 异步事件 | **Redis Stream** | rpg 打赏 → 发事件 → blog 加声望、notify 发通知 |
| 实时推送 | **Redis pub/sub** | 任意服务 → blog-service WS Hub → 推客户端 |

### 3.3 共享单库的表归属

虽然共用一个 MySQL 库，但**每个服务只访问自己负责的表**，通过代码约束（Ent schema 按服务分组），不跨服务查别人的表：

| 服务 | 负责的表 |
|------|---------|
| user-service | user, role, privilege, dept, menu, user_roles_role |
| blog-service | article, category, tag, article_tags_tag, comment, reply, like, collect, msgboard, link, resources, file, notification, operation_log |
| rpg-service | rpg_* 全部, pay_order |

> 学习要点：微服务的"数据库独占"原则是为了**松耦合**，共享库时通过代码纪律模拟这个边界——每个服务的 Ent schema 只定义自己的表，repo 只查自己的表。跨表需求走服务间调用，不走 JOIN。

### 3.4 认证流程（微服务版）

```
客户端 --[Bearer token]--> gateway
  gateway 验 JWT（调 user-service 的 VerifyToken 或本地验签）
  提取 userID/roles，写入 gRPC metadata
  转发到内部服务
内部服务从 metadata 取 userID，不重复验签（信任内网）
```

JWT 密钥共享给 gateway 做本地验签（性能优先），或调 user-service 验签（学习远程调用，但每次登录态校验多一次 RPC，可加 Redis 缓存）。

---

## 四、原生 WebSocket 完整方案

### 4.1 坑全景与对策

| # | 坑 | 对策 |
|---|----|------|
| 1 | 无自动重连 | 前端自实现指数退避重连 |
| 2 | 无心跳保活 | 应用层 ping/pong + 控制帧 ping 双管齐下 |
| 3 | 无消息确认 | 业务幂等 + 消息序号 seq + 重连拉取补漏 |
| 4 | 无房间/命名空间 | 服务端维护 userID→Client 和 topic→[]userID 映射 |
| 5 | 并发写 panic | 所有写走单一 writeLoop goroutine + channel |
| 6 | 读超时空连接 | SetReadDeadline + pong 续期 |
| 7 | 写超时阻塞 | SetWriteDeadline + 超时关闭 |
| 8 | 连接泄漏 | 心跳超时清理 + SetCloseHandler |
| 9 | 大消息阻塞 | SetReadLimit(4096) |
| 10 | send channel 满死锁 | select default 踢人 |
| 11 | 优雅关闭 | hub.Close() 发 close 帧等排空 |
| 12 | 跨实例广播 | Redis pub/sub 广播 |

### 4.2 Hub + Client 完整实现

#### Hub（连接管理与消息路由）

```go
// services/blog/internal/ws/hub.go
package ws

import (
    "context"
    "sync"
    "time"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10 // 54s
    maxMessageSize = 4096
    sendBufSize    = 256
)

type Hub struct {
    mu sync.RWMutex

    // userID -> 该用户所有连接（支持多端登录）
    clients map[uint64]map[*Client]struct{}

    // topic -> 订阅的 userID 集合（替代 Socket.IO room）
    topics map[string]map[uint64]struct{}

    register   chan *Client
    unregister chan *Client
    broadcast  chan broadcastMsg

    rds RedisSubscriber  // 订阅 Redis pub/sub，接收其他服务发来的推送
}

type broadcastMsg struct {
    userIDFilter uint64 // 0 = 广播所有人
    topicFilter  string // 空 = 不限 topic
    payload      []byte
}

func NewHub(rds RedisSubscriber) *Hub {
    return &Hub{
        clients:   make(map[uint64]map[*Client]struct{}),
        topics:    make(map[string]map[uint64]struct{}),
        register:  make(chan *Client, 64),
        unregister: make(chan *Client, 64),
        broadcast: make(chan broadcastMsg, 256),
        rds:       rds,
    }
}

func (h *Hub) Run(ctx context.Context) {
    // 订阅 Redis pub/sub，接收 rpg-service 等其他服务发来的推送
    rdsCh := h.rds.Subscribe(ctx, "realtime:push")

    for {
        select {
        case c := <-h.register:
            h.addClient(c)
        case c := <-h.unregister:
            h.removeClient(c)
        case msg := <-h.broadcast:
            h.dispatch(msg)
        case rdsMsg := <-rdsCh:
            // 其他服务通过 Redis pub/sub 发来的推送
            h.dispatch(parseRedisMsg(rdsMsg))
        case <-ctx.Done():
            h.closeAll()
            return
        }
    }
}

func (h *Hub) Register(c *Client)  { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }

func (h *Hub) addClient(c *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.clients[c.userID] == nil {
        h.clients[c.userID] = make(map[*Client]struct{})
    }
    h.clients[c.userID][c] = struct{}{}
}

func (h *Hub) removeClient(c *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if conns, ok := h.clients[c.userID]; ok {
        delete(conns, c)
        if len(conns) == 0 {
            delete(h.clients, c.userID)
            // 清理该用户在所有 topic 的订阅
            for topic, users := range h.topics {
                delete(users, c.userID)
                if len(users) == 0 {
                    delete(h.topics, topic)
                }
            }
        }
    }
}

// dispatch 根据 filter 分发消息到目标 client
func (h *Hub) dispatch(msg broadcastMsg) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    var targets map[uint64]struct{}
    if msg.topicFilter != "" {
        targets = h.topics[msg.topicFilter]
    }

    for uid, conns := range h.clients {
        if msg.userIDFilter != 0 && uid != msg.userIDFilter {
            continue
        }
        if msg.topicFilter != "" {
            if _, ok := targets[uid]; !ok {
                continue
            }
        }
        for c := range conns {
            select {
            case c.send <- msg.payload:
            default:
                // 缓冲区满，踢掉慢客户端
                go func(cl *Client) { h.unregister <- cl }(c)
            }
        }
    }
}

// PublishToUser 给指定用户推送（本服务内调用）
func (h *Hub) PublishToUser(userID uint64, payload []byte) {
    h.broadcast <- broadcastMsg{userIDFilter: userID, payload: payload}
}

// PublishToTopic 给 topic 订阅者推送
func (h *Hub) PublishToTopic(topic string, payload []byte) {
    h.broadcast <- broadcastMsg{topicFilter: topic, payload: payload}
}

// SubscribeTopic 用户订阅 topic（替代 socket.join）
func (h *Hub) SubscribeTopic(userID uint64, topic string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if h.topics[topic] == nil {
        h.topics[topic] = make(map[uint64]struct{})
    }
    h.topics[topic][userID] = struct{}{}
}

func (h *Hub) UnsubscribeTopic(userID uint64, topic string) {
    h.mu.Lock()
    defer h.mu.Unlock()
    if users, ok := h.topics[topic]; ok {
        delete(users, userID)
        if len(users) == 0 {
            delete(h.topics, topic)
        }
    }
}

func (h *Hub) closeAll() {
    h.mu.Lock()
    defer h.mu.Unlock()
    for _, conns := range h.clients {
        for c := range conns {
            close(c.send)
        }
    }
}
```

#### Client（单连接读写循环）

```go
// services/blog/internal/ws/client.go
package ws

import (
    "encoding/json"
    "time"

    "github.com/gorilla/websocket"
)

type Client struct {
    userID uint64
    conn   *websocket.Conn
    send   chan []byte
    hub    *Hub
}

// 业务消息格式（替代 Socket.IO 的 event）
type Message struct {
    Type string          `json:"type"`            // ping/subscribe/notification/...
    Seq  uint64          `json:"seq,omitempty"`   // 消息序号，前端断线补漏用
    Data json.RawMessage `json:"data,omitempty"`
}

func NewClient(userID uint64, conn *websocket.Conn, hub *Hub) *Client {
    return &Client{
        userID: userID,
        conn:   conn,
        send:   make(chan []byte, sendBufSize),
        hub:    hub,
    }
}

func (c *Client) ReadPump() {
    defer func() {
        c.hub.Unregister(c)
        c.conn.Close()
    }()
    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    // 控制帧 pong：浏览器自动回，续期读超时
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })

    for {
        _, raw, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
        var msg Message
        if err := json.Unmarshal(raw, &msg); err != nil {
            continue
        }
        switch msg.Type {
        case "ping":
            // 应用层心跳，回 pong
            c.send <- []byte(`{"type":"pong"}`)
        case "subscribe":
            var d struct{ Topic string `json:"topic"` }
            json.Unmarshal(msg.Data, &d)
            c.hub.SubscribeTopic(c.userID, d.Topic)
        case "unsubscribe":
            var d struct{ Topic string `json:"topic"` }
            json.Unmarshal(msg.Data, &d)
            c.hub.UnsubscribeTopic(c.userID, d.Topic)
        // 其他业务消息路由...
        }
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    for {
        select {
        case msg, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
                return
            }
        case <-ticker.C:
            // 控制帧 ping：探活（浏览器底层自动回 pong）
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

> **两套心跳并存**：服务端发控制帧 `PingMessage`（浏览器自动回 pong，触发 PongHandler 续期读超时）——连接探活；客户端发应用层 `{"type":"ping"}`，服务端回 `{"type":"pong"}`——业务心跳。两套独立。

#### 跨服务推送（rpg-service → blog-service WS Hub）

rpg-service 需要推送给客户端时，通过 Redis pub/sub 发消息，blog-service 的 Hub 订阅后转发：

```go
// rpg-service 内：升级后推送通知
func (s *LevelService) OnLevelUp(ctx context.Context, userID uint64, newLevel int) error {
    payload, _ := json.Marshal(map[string]any{
        "type": "rpg.levelup",
        "seq":  s.nextSeq(),
        "data": map[string]any{"level": newLevel},
    })
    // 发到 Redis pub/sub，blog-service 的 Hub 会收到
    return s.rds.Publish(ctx, "realtime:push", payload)
}
```

```go
// blog-service 内：Redis 订阅器
type RedisSubscriber interface {
    Subscribe(ctx context.Context, channel string) <-chan []byte
}

type rueidisSubscriber struct{ client *rueidis.Client }

func (s *rueidisSubscriber) Subscribe(ctx context.Context, channel string) <-chan []byte {
    ch := make(chan []byte, 256)
    s.client.Subscribe(ctx, channel, func(msg rueidis.PubMessage) {
        select {
        case ch <- msg.Payload:
        default:
        }
    })
    return ch
}
```

### 4.3 WebSocket 升级入口

```go
// services/blog/internal/handler/ws.go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  4096,
    WriteBufferSize: 4096,
    CheckOrigin: func(r *http.Request) bool {
        return true // 生产环境校验 Origin 白名单
    },
}

func RegisterWS(h *server.Hertz, hub *ws.Hub, jwtSecret string) {
    h.GET("/realtime", func(ctx context.Context, c *app.RequestContext) {
        // 从 query 或 header 取 token 验证身份
        token := string(c.Query("token"))
        if token == "" {
            token = strings.TrimPrefix(string(c.GetHeader("Authorization")), "Bearer ")
        }
        claims := &CustomClaims{}
        tk, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        if err != nil || !tk.Valid {
            c.AbortWithStatus(consts.StatusUnauthorized)
            return
        }
        conn, err := upgrader.Upgrade(ctx, c.GetWriter(), c.GetHeader())
        if err != nil {
            return
        }
        client := ws.NewClient(claims.UID, conn, hub)
        hub.Register(client)
        go client.ReadPump()
        go client.WritePump()
    })
}
```

### 4.4 前端改造（Nuxt composable）

原生 WebSocket 是浏览器内置 API，不需要引库，自己封装一个 composable ~100 行：

```ts
// composables/useWs.ts（blog-home-nuxt / blog-admin 共用）
export const useWs = () => {
  const config = useRuntimeConfig()
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  let reconnectTimer: number
  let reconnectAttempts = 0
  let heartbeatTimer: number
  let lastSeq = 0  // 断线补漏用

  const connect = () => {
    const token = useCookie('token').value
    ws.value = new WebSocket(`${config.public.wsUrl}/realtime?token=${token}`)

    ws.value.onopen = () => {
      connected.value = true
      reconnectAttempts = 0
      startHeartbeat()
      // 断线重连后补漏：拉离线期间的消息
      if (lastSeq > 0) {
        fetch(`/api/v1/notification/since?seq=${lastSeq}`)
          .then(r => r.json())
          .then(msgs => msgs.forEach(handleMessage))
      }
    }

    ws.value.onmessage = (e) => {
      const msg = JSON.parse(e.data)
      if (msg.type === 'pong') return
      if (msg.seq) lastSeq = msg.seq
      handleMessage(msg)
    }

    ws.value.onclose = () => {
      connected.value = false
      stopHeartbeat()
      scheduleReconnect()
    }

    ws.value.onerror = () => ws.value?.close()
  }

  const startHeartbeat = () => {
    heartbeatTimer = window.setInterval(() => {
      if (connected.value) ws.value?.send(JSON.stringify({ type: 'ping' }))
    }, 20000)
  }

  const stopHeartbeat = () => clearInterval(heartbeatTimer)

  const scheduleReconnect = () => {
    const delay = Math.min(1000 * 2 ** reconnectAttempts, 30000)
    const jitter = Math.random() * 1000
    reconnectTimer = window.setTimeout(() => {
      reconnectAttempts++
      connect()
    }, delay + jitter)
  }

  const send = (type: string, data?: unknown) => {
    ws.value?.send(JSON.stringify({ type, data }))
  }

  onMounted(connect)
  onUnmounted(() => {
    clearTimeout(reconnectTimer)
    stopHeartbeat()
    ws.value?.close()
  })

  return { connected, send }
}
```

| Socket.IO 用法 | 原生 WS 等价 |
|----------------|-------------|
| `io.connect('/realtime')` | `new WebSocket(url)` |
| 自动重连 | 指数退避 + jitter |
| `socket.on('event', cb)` | `{type, data}` 消息路由 |
| `socket.join('room')` | 发 `{"type":"subscribe","data":{"topic":"..."}}` |
| 断线补漏 | 重连后 `GET /api/v1/notification/since?seq=xxx` |

---

## 五、关键技术落地

### 5.1 配置管理（Viper）

```go
// pkg/config/config.go
type Config struct {
    App   AppConfig   `mapstructure:"app"`
    HTTP  HTTPConfig  `mapstructure:"http"`
    GRPC  GRPCConfig  `mapstructure:"grpc"`
    MySQL MySQLConfig `mapstructure:"mysql"`
    Redis RedisConfig `mapstructure:"redis"`
    JWT   JWTConfig   `mapstructure:"jwt"`
    OAuth OAuthConfig `mapstructure:"oauth"`
    Mail  MailConfig  `mapstructure:"mail"`
    Pay   PayConfig   `mapstructure:"pay"`
}

type AppConfig struct {
    Name      string `mapstructure:"name"`
    Env       string `mapstructure:"env"`
    APIPrefix string `mapstructure:"api_prefix"` // /api/v1
}

type JWTConfig struct {
    Secret     string        `mapstructure:"secret"`
    AccessTTL  time.Duration `mapstructure:"access_ttl"`
    RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}
```

### 5.2 统一响应与错误码（保持前端兼容）

```go
// pkg/response/response.go
type Body struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c context.Context, data interface{}) {
    hresp.JSON(c, consts.StatusOK, Body{Code: 0, Message: "success", Data: data})
}

func Error(c context.Context, ec errcode.ErrCode, args ...any) {
    hresp.JSON(c, consts.StatusOK, Body{Code: ec.Code(), Message: ec.Message(args...)})
}
```

错误码与原项目一一对应。

### 5.3 gRPC 服务定义与调用

```protobuf
// proto/user/v1/user.proto
syntax = "proto3";
package user.v1;
option go_package = "github.com/Jiang-Xia/blog-server-go/proto/user/v1;userv1";

service UserService {
  rpc GetUser(GetUserReq) returns (User);
  rpc GetUserBatch(GetUserBatchReq) returns (UserBatch);
  rpc VerifyToken(VerifyTokenReq) returns (VerifyTokenResp);
}

message User {
  uint64 id = 1;
  string nickname = 2;
  string avatar = 3;
  string email = 4;
  repeated string roles = 5;
}

message GetUserReq { uint64 id = 1; }
message GetUserBatchReq { repeated uint64 ids = 1; }
message UserBatch { repeated User users = 1; }
```

```protobuf
// proto/blog/v1/article.proto
service ArticleService {
  rpc CreateArticle(CreateArticleReq) returns (Article);
  rpc ListArticles(ListArticlesReq) returns (ArticleList);
  rpc GetArticle(GetArticleReq) returns (Article);
  rpc DeleteArticle(DeleteArticleReq) returns (Empty);
}
```

gateway 调内部服务：

```go
// services/gateway/internal/aggregator/article.go
func (a *ArticleAggregator) ArticleDetail(ctx context.Context, id uint64) (*ArticleDetail, error) {
    // 1. 调 blog-service 拿文章
    article, err := a.blogClient.GetArticle(ctx, &blogv1.GetArticleReq{Id: id})
    if err != nil {
        return nil, err
    }
    // 2. 调 user-service 拿作者信息
    author, err := a.userClient.GetUser(ctx, &userv1.GetUserReq{Id: article.AuthorId})
    if err != nil {
        return nil, err
    }
    return &ArticleDetail{Article: article, Author: author}, nil
}
```

### 5.4 认证授权

```go
// services/gateway/internal/middleware/jwt.go
// gateway 本地验签，提取 userID 写入 gRPC metadata
func JWT(secret string) app.HandlerFunc {
    return func(ctx context.Context, c *app.RequestContext) {
        token := strings.TrimPrefix(string(c.GetHeader("Authorization")), "Bearer ")
        claims := &CustomClaims{}
        tk, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        if err != nil || !tk.Valid {
            response.Error(ctx, errcode.TokenExpired)
            c.Abort()
            return
        }
        ctx = context.WithValue(ctx, ctxKeyUserID{}, claims.UID)
        ctx = context.WithValue(ctx, ctxKeyRoles{}, claims.Roles)
        c.Next(ctx)
    }
}

// gRPC 拦截器：gateway 调内部服务时透传 userID
func ForwardAuth(ctx context.Context) context.Context {
    uid := ctxutil.UserID(ctx)
    return metadata.AppendToOutgoingContext(ctx, "x-user-id", strconv.FormatUint(uid, 10))
}

// 内部服务拦截器：从 metadata 取 userID
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    md, _ := metadata.FromIncomingContext(ctx)
    if v := md.Get("x-user-id"); len(v) > 0 {
        uid, _ := strconv.ParseUint(v[0], 10, 64)
        ctx = context.WithValue(ctx, ctxKeyUserID{}, uid)
    }
    return handler(ctx, req)
}
```

### 5.5 密码与加密（兼容旧数据）

```go
// pkg/crypto/password.go
func Hash(plain string) (string, error) {
    b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
    return string(b), err
}

func Verify(hashed, plain string) bool {
    if strings.HasPrefix(hashed, "$2") {
        return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
    }
    // 兼容老 crypto-js 哈希
    return legacyCryptoJSVerify(hashed, plain)
}
```

存量用户首次登录成功后静默升级 bcrypt。RSA 解密用标准库 `crypto/rsa`。

### 5.6 Ent ORM

```go
// services/blog/ent/schema/article.go
type Article struct{ ent.Schema }

func (Article) Fields() []ent.Field {
    return []ent.Field{
        field.Uint64("id"),
        field.String("title").NotEmpty().MaxLen(200),
        field.Text("content"),
        field.Text("content_html").Optional(),
        field.String("cover").Optional(),
        field.Uint8("status").Default(0), // 0草稿 1发布 2定时
        field.Uint64("likes").Default(0),
        field.Uint64("views").Default(0),
        field.Bool("topping").Default(false),
        field.Bool("is_delete").Default(false),
        field.Uint64("uid"),
        field.Time("created_at").Default(time.Now),
        field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
    }
}

func (Article) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("category", Category.Type).Ref("articles").Unique().Field("category_id"),
        edge.From("tags", Tag.Type).Ref("articles"),
        edge.To("comments", Comment.Type),
    }
}
```

### 5.7 事件驱动（Redis Stream）

```go
// pkg/event/bus.go
type Bus interface {
    Publish(ctx context.Context, evt Event) error
    Subscribe(ctx context.Context, group string, handler func(Event) error)
}

type Event struct {
    Type    string          // article.published, tip.given, ...
    Payload json.RawMessage
}

// redisStreamBus 实现：XADD 发布，XREADGROUP 消费，含重试与死信
```

事件示例：
- `article.published` → blog 发，notify 消费发邮件
- `tip.given` → rpg 发，blog 消费加作者声望
- `rpg.levelup` → rpg 发，blog WS Hub 推送客户端

### 5.8 定时任务

```go
// pkg/scheduler/scheduler.go
type Scheduler struct{ c *cron.Cron }

func (s *Scheduler) Register(job NamedJob) {
    s.c.AddJob(job.Spec(), cron.NewChain(cron.Recover(logger.Stdout)).Then(job))
}

type NamedJob interface {
    cron.Job
    Spec() string
    Name() string
}
```

文章定时发布、签到重置、赛季结算等迁移为 NamedJob。

### 5.9 依赖注入（wire）

每个服务独立 wire：

```go
// services/blog/internal/app/wire.go
//go:build wireinject

func InitializeService(cfg *config.Config) (*Service, func(), error) {
    wire.Build(
        db.NewEntClient,
        cache.NewRueidis,
        event.NewBus,
        ws.NewHub,
        repo.NewArticleRepo,
        repo.NewCategoryRepo,
        service.NewArticleService,
        handler.NewArticleHandler,
        handler.RegisterWS,
        server.NewGRPCServer,
        server.NewHTTPServer,
        wire.Struct(new(Service), "*"),
    )
    return nil, nil, nil
}
```

---

## 六、目录结构

```
blog-server-go/
├── proto/                        # protobuf 定义（4 服务共享）
│   ├── user/v1/user.proto
│   ├── blog/v1/article.proto
│   ├── blog/v1/comment.proto
│   └── rpg/v1/rpg.proto
├── pkg/                          # 跨服务共享库
│   ├── response/                 # 统一响应体
│   ├── errcode/                  # 错误码
│   ├── crypto/                   # 密码哈希/RSA
│   ├── logger/                   # zap 封装
│   ├── otel/                     # OpenTelemetry
│   ├── config/                   # 配置加载
│   └── timeutil/
├── services/                     # 4 个微服务
│   ├── gateway/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/          # BFF REST handler
│   │   │   ├── aggregator/       # 多服务响应聚合
│   │   │   ├── middleware/       # JWT/限流/CORS
│   │   │   └── config/
│   │   └── Dockerfile
│   ├── blog/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── domain/           # 领域模型
│   │   │   ├── service/          # 用例
│   │   │   ├── repo/             # Ent 仓储
│   │   │   ├── handler/          # gRPC + WS handler
│   │   │   ├── ws/               # WebSocket Hub（核心）
│   │   │   ├── event/            # 事件发布/消费
│   │   │   └── config/
│   │   ├── ent/schema/           # Ent schema（blog 相关表）
│   │   └── Dockerfile
│   ├── user/
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── domain/
│   │   │   ├── service/
│   │   │   ├── repo/
│   │   │   ├── handler/          # gRPC handler
│   │   │   ├── auth/             # JWT/OAuth/验证码
│   │   │   └── config/
│   │   ├── ent/schema/           # Ent schema（user 相关表）
│   │   └── Dockerfile
│   └── rpg/
│       ├── cmd/main.go
│       ├── internal/
│       │   ├── domain/
│       │   ├── service/
│       │   ├── repo/
│       │   ├── handler/
│       │   ├── pay/               # 支付宝适配
│       │   └── config/
│       ├── ent/schema/            # Ent schema（rpg 相关表）
│       └── Dockerfile
├── migrations/                    # golang-migrate 脚本
├── deploy/
│   ├── docker/
│   │   ├── docker-compose.yml
│   │   └── mysql.cnf             # MySQL 调优配置
│   └── nginx/
├── configs/                       # 各服务配置
│   ├── gateway.yaml
│   ├── blog.yaml
│   ├── user.yaml
│   └── rpg.yaml
├── scripts/
├── Makefile
├── buf.yaml                       # protobuf 管理
├── buf.gen.yaml
├── go.mod                         # monorepo 单 go.mod
└── go.sum
```

---

## 七、2G 服务器部署优化

### 7.1 必做优化

#### MySQL 调小内存

```ini
# deploy/docker/mysql.cnf
[mysqld]
innodb_buffer_pool_size = 128M
max_connections = 100
performance_schema = OFF
tmp_table_size = 16M
max_heap_table_size = 16M
table_open_cache = 200
```

#### Go 服务限制连接池

```go
// 每个服务的 Ent client 初始化
db.SetMaxOpenConns(15)    // 4 服务 × 15 = 60 连接 < MySQL 100
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

#### Go 内存限制

```dockerfile
# 每个服务 Dockerfile 加
ENV GOMEMLIMIT=80MiB    # 硬限制 80MB，到上限更积极 GC
ENV GOGC=50             # 让 GC 更勤快，换内存
```

### 7.2 建议加：swap 兜底

```bash
fallocate -l 1G /swapfile
chmod 600 /swapfile
mkswap /swapfile
swapon /swapfile
# 写入 /etc/fstab 持久化
echo '/swapfile none swap sw 0 0' >> /etc/fstab
```

### 7.3 docker-compose

```yaml
version: "3.9"
services:
  gateway:
    build: ./services/gateway
    ports: ["8000:8000"]
    environment:
      - GOMEMLIMIT=60MiB
      - GOGC=50
    depends_on: [blog, user, rpg]
    restart: always

  blog:
    build: ./services/blog
    ports: ["5001:5001"]
    environment:
      - GOMEMLIMIT=90MiB
      - GOGC=50
    depends_on: [mysql, redis]
    restart: always

  user:
    build: ./services/user
    ports: ["5002:5002"]
    environment:
      - GOMEMLIMIT=70MiB
      - GOGC=50
    depends_on: [mysql, redis]
    restart: always

  rpg:
    build: ./services/rpg
    ports: ["5003:5003"]
    environment:
      - GOMEMLIMIT=90MiB
      - GOGC=50
    depends_on: [mysql, redis]
    restart: always

  mysql:
    image: mysql:8.0
    command: --default-authentication-plugin=mysql_native_password
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: blog
    volumes:
      - mysql_data:/var/lib/mysql
      - ./deploy/docker/mysql.cnf:/etc/mysql/conf.d/mysql.cnf
    ports: ["3306:3306"]
    restart: always

  redis:
    image: redis:7-alpine
    command: redis-server --maxmemory 64mb --maxmemory-policy allkeys-lru
    volumes: ["redis_data:/data"]
    ports: ["6379:6379"]
    restart: always

volumes:
  mysql_data:
  redis_data:
```

### 7.4 Dockerfile（每服务，多阶段构建）

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY . .
RUN go mod download && \
    go generate ./... && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" \
      -o /out/service ./services/blog/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/service /app/service
COPY configs/blog.yaml /app/configs/config.yaml
USER nonroot:nonroot
EXPOSE 5001
ENTRYPOINT ["/app/service", "-c", "/app/configs/config.yaml"]
```

### 7.5 Makefile

```makefile
.PHONY: proto ent build test lint docker up down

proto:
	@buf generate

ent:
	@go generate ./services/...

build:
	@for svc in gateway blog user rpg; do \
	  CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$$svc ./services/$$svc/cmd; \
	done

up:
	@docker-compose up -d --build

down:
	@docker-compose down

test:
	@go test -race -cover ./...

lint:
	@golangci-lint run

wire:
	@for svc in gateway blog user rpg; do \
	  (cd services/$$svc/internal/app && wire); \
	done
```

---

## 八、迁移策略与路线图

### 8.1 核心原则

**单体先行 → 验证边界 → 拆 4 服务**。不要一上来就 4 服务，先用模块化单体把全部功能跑通（3 个月），再物理拆分（2 周）。原因：
- 单体阶段专注业务逻辑，不被基础设施分心
- 验证模块边界是否合理
- 拆分时是"物理移动 + 换通信方式"，不是重写

### 8.2 路线图

| 阶段 | 周期 | 架构 | 内容 | 验收 |
|------|------|------|------|------|
| **阶段0：脚手架** | 第1周 | 单体 | 仓库初始化、Hertz+Ent+rueidis+wire、proto 定义、配置/日志/响应/错误码/中间件 | /health 200 |
| **阶段1：模块化单体** | 第2~12周 | 单体 | 全部功能跑通，模块边界清晰，跨模块走接口 | 全功能回归通过 |
| **阶段2：拆 4 服务** | 第13~14周 | 4 服务 | 物理拆分 gateway/blog/user/rpg，gRPC 通信，共享库 | 4 服务独立运行 |
| **阶段3：可观测+上线** | 第15~16周 | 4 服务 | OTel 链路、Prometheus 指标、docker-compose 部署、监控告警 | 生产稳定 1 周 |

### 8.3 阶段1详细周计划

| 周 | 模块 |
|----|------|
| 1 | 脚手架：仓库、Ent schema 全量建模、配置/日志/响应/中间件 |
| 2 | auth + captcha + pub（统计） |
| 3-4 | article + category + tag |
| 5-6 | comment + reply + like + collect |
| 7 | msgboard + link + file + email |
| 8 | admin（角色/权限/部门/菜单） |
| 9 | notification + operation-log + sensitive + scheduler |
| 10 | realtime（WS Hub 集成） |
| 11 | pay + rpg |
| 12 | 全量回归 + 契约测试 |

### 8.4 拆服务（阶段2）操作步骤

以拆 user-service 为例：

| 步骤 | 工作 | 耗时 |
|------|------|------|
| 1 | `services/user/` 目录独立，搬 user 相关模块 | 1h |
| 2 | 写 user.proto，buf generate 生成 gRPC 代码 | 2h |
| 3 | 其他模块的 UserService 调用从本地接口换成 gRPC client | 3h |
| 4 | 拆出 user 的 main.go / config / Dockerfile | 1h |
| 5 | 联调测试 | 4h |
| **合计** | | **~1.5 天** |

4 个服务约 1 周拆完。

### 8.5 绞杀者模式（生产切换）

阶段1单体跑稳后，Nginx 按路径灰度切流到 Go 单体，NestJS 逐步下线。阶段2拆服务是 Go 单体内部拆分，不再涉及 NestJS。

---

## 九、测试与可观测性

### 9.1 测试策略

| 层级 | 工具 | 范围 |
|------|------|------|
| 单元测试 | testify | service/repo，mock 接口 |
| 集成测试 | testcontainers-go | 真 DB/Redis |
| 契约测试 | newman + postman | 对齐旧接口响应 |
| gRPC 契约 | buf breaking | proto 兼容性 |
| 压测 | k6 / wrk | 上线前对比 |

### 9.2 可观测性

| 维度 | 工具 |
|------|------|
| 日志 | zap → Loki 或文件 |
| 指标 | Prometheus + Grafana（/metrics） |
| 链路 | OpenTelemetry，跨服务 trace 透传 |
| profiling | pprof :6060 |

跨服务链路：gateway 收到请求生成 trace → gRPC metadata 透传 traceId → 内部服务继承 span → 全链路可视化。

```go
// gRPC server/client 注册 otel 拦截器，自动透传 trace
grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor())
```

---

## 十、风险与决策点

| # | 决策点 | 选择 | 状态 |
|---|--------|------|------|
| 1 | Web 框架 | Hertz | ✅ 已确认 |
| 2 | 服务数量 | 4（gateway/blog/user/rpg） | ✅ 已确认 |
| 3 | 数据库 | 共享 MySQL 单库 | ✅ 已确认 |
| 4 | ORM | Ent | 推荐 |
| 5 | WebSocket | gorilla/websocket + 原生前端 | ✅ 已确认 |
| 6 | 内部通信 | gRPC | ✅ |
| 7 | 事件驱动 | Redis Stream | ✅ |
| 8 | 密码哈希 | 兼容 crypto-js + 静默升级 bcrypt | 推荐 |
| 9 | 部署 | docker-compose（2G 机器） | ✅ |
| 10 | 演进策略 | 单体先行 → 拆 4 服务 | ✅ |

---

## 附录 A：原 API 路径与服务映射

| 原路径 | 服务 | gRPC service |
|--------|------|--------------|
| `/api/v1/auth` | user-service | AuthService |
| `/api/v1/user` | user-service | UserService |
| `/api/v1/captcha` | user-service | CaptchaService |
| `/api/v1/admin` `/api/v1/admin/system` | user-service | AdminService |
| `/api/v1/article` | blog-service | ArticleService |
| `/api/v1/category` | blog-service | CategoryService |
| `/api/v1/tag` | blog-service | TagService |
| `/api/v1/comment` `/api/v1/reply` | blog-service | CommentService |
| `/api/v1/like` `/api/v1/collect` `/api/v1/msgboard` | blog-service | InteractionService |
| `/api/v1/file` | blog-service | FileService |
| `/api/v1/notification` | blog-service | NotifyService |
| `/api/v1/pub` | gateway 聚合 | — |
| `/api/v1/pay` | rpg-service | PayService |
| `/api/v1/rpg` `/api/v1/admin/rpg` | rpg-service | RpgService |
| `/realtime` (WS) | blog-service | — |

## 附录 B：第一周动作清单

1. **Day 1-2**：monorepo 初始化、`go mod init`、引入 Hertz/Ent/rueidis/zap/Viper/wire/buf
2. **Day 2-3**：用 `entimport` 从现有 MySQL 反向生成 Ent schema，人工校对
3. **Day 3-4**：实现 `pkg/` 下 config/logger/response/errcode/crypto + 中间件骨架
4. **Day 4-5**：wire 装配，单体能启动、`/health` 200、连上 MySQL/Redis
5. **Day 6-7**：迁移 auth + captcha + pub，跑通登录链路，写契约测试

## 附录 C：一句话总结

> **Hertz + Ent + rueidis + gRPC + wire，单体先行 3 个月验证边界，再拆为 gateway/blog/user/rpg 4 服务，共享 MySQL 单库，Redis Stream 事件驱动，gorilla/websocket 原生 WS 替代 Socket.IO（两套心跳+断线补漏+跨服务 Redis pub/sub 广播），docker-compose 部署在 2G 机器，单服务 distroless 镜像 < 25MB。**
