# HTTP / gRPC 路由全表

> **更新日期**：2026-07-10  
> **对外主入口（Nest 替换）**：`http://127.0.0.1:8000`（`services/monolith`，`make dev` / `.\scripts\dev.ps1`）  
> **数据来源**：`services/monolith/internal/handler/`（单体全量）；§1–5 另列 gateway / 各微服务（**架构学习对照**，可能落后于单体）  
> **Nest blog-server**：`http://127.0.0.1:5000`（待替换，可与 Go 并行）  
> **微服务 gateway**：`http://127.0.0.1:8000`（与单体二选一）；开发直连见各服务端口表。

## 约定

| 项 | 说明 |
|----|------|
| API 前缀 | 默认 `/api/v1`（`configs/*.yaml` → `app.api_prefix`） |
| 响应格式 | `{ code, message, data }`（gateway BFF 与各微服务 HTTP 一致） |
| 鉴权列 | **公开** = handler 未强制 JWT；**JWT** = 需 `Authorization: Bearer`；实际还可能经 RBAC `permission` 中间件（Redis/DB 权限表） |
| Gateway 列 | **BFF-gRPC** = gateway 本地聚合/调内部 gRPC；**代理** = HTTP 反向代理；**本地** = gateway 自身处理 |

### 服务端口

| 服务 | HTTP | gRPC | 说明 |
|------|------|------|------|
| **monolith** | **`:8000`** | — | **Nest 替换主入口**（全路由） |
| gateway | `:8000` | — | 微服务统一 REST 入口（与单体二选一） |
| blog-service | `:5001` | `:50051` | 文章/互动/WS/通知/定时任务 |
| user-service | `:5002` | `:50052` | 用户/RBAC/敏感词 |
| rpg-service | `:5003` | `:50053` | RPG/支付/公开主页 |

### Gateway 路由决策（`services/gateway/internal/proxy/router.go`）

```
/api/v1/* 请求
  ├─ pub/*              → BFF（仅 /pub/stats 已注册；其余无 upstream → 502）
  ├─ article/info       → BFF-gRPC（blog.GetArticleDetail）
  ├─ user/public/:uid   → BFF-gRPC（rpg.GetPublicProfile，精确 3 段路径）
  ├─ user/* captcha role dept privilege admin/* sensitive-word operation-log
  │                     → 代理 → user-service
  ├─ rpg/* admin/rpg/* pay/* user/public/* rpg/public/*
  │                     → 代理 → rpg-service
  └─ 其余               → 代理 → blog-service

/realtime               → 代理 → blog-service（WebSocket）
/health                 → gateway 本地
/metrics                → gateway 本地（可观测性开启时）
```

---

## 1. Gateway 本地路由

| 方法 | 路径 | Gateway | 说明 | 源码 |
|------|------|---------|------|------|
| GET | `/health` | 本地 | 健康检查 | `gateway/internal/app/app.go` |
| GET | `/api/v1/health` | 本地 | 健康检查 | 同上 |
| GET | `/metrics` | 本地 | Prometheus（`enable_metrics`） | 同上 |
| GET | `/api/v1/pub/stats` | **BFF-gRPC** | blog.GetPubStats + user.CountUsers | `gateway/internal/aggregator/stats.go` |
| GET | `/api/v1/article/info` | **BFF-gRPC** | blog.GetArticleDetail | `gateway/internal/aggregator/article.go` |
| GET | `/api/v1/user/public/:uid` | **BFF-gRPC** | rpg.GetPublicProfile | `gateway/internal/aggregator/profile.go` |
| ANY | `/realtime` | 代理→blog | WebSocket 升级 | `gateway/internal/proxy/router.go` |
| ANY | `/realtime/*path` | 代理→blog | WS 子路径 | 同上 |
| ANY | `/api/v1/*`（未上表） | 代理 | 按前缀转发 user/blog/rpg | 同上 |

---

## 2. user-service（`:5002`）

> 注册：`services/user/internal/handler/register_user.go`  
> Gateway：路径前缀 `user`、`captcha`、`role`、`dept`、`privilege`、`admin/`（不含 `admin/rpg`）、`sensitive-word`、`operation-log` → 代理到此服务。

### 2.1 健康与静态

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/health` | 公开 | 健康检查 |
| GET | `/api/v1/health` | 公开 | 健康检查 |
| GET | `/metrics` | 公开 | Prometheus |
| GET | `/static/*` | 公开 | 上传静态文件（配置了 `storage.upload_path` 时） |

### 2.2 验证码

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/captcha` | 公开 |
| POST | `/api/v1/captcha/verify` | 公开 |

### 2.3 用户 `/api/v1/user`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/user/authCode` | 公开 |
| POST | `/api/v1/user/register` | 公开 |
| POST | `/api/v1/user/login` | 公开 |
| GET | `/api/v1/user/refresh` | 公开 |
| GET | `/api/v1/user/info` | JWT |
| POST | `/api/v1/user/list` | 公开 |
| PATCH | `/api/v1/user/status` | JWT |
| PATCH | `/api/v1/user/edit` | JWT |
| PATCH | `/api/v1/user/password` | JWT |
| POST | `/api/v1/user/resetPassword` | 公开 |
| DELETE | `/api/v1/user` | JWT |
| POST | `/api/v1/user/email/sendCode` | 公开 |
| POST | `/api/v1/user/email/register` | 公开 |
| POST | `/api/v1/user/email/login` | 公开 |
| GET | `/api/v1/user/auth/github` | 公开 |
| GET | `/api/v1/user/auth/github/callback` | 公开 |
| POST | `/api/v1/user/auth/ticket/exchange` | 公开 |
| POST | `/api/v1/user/auth/wechat/miniprogram` | 公开 |
| POST | `/api/v1/user/admin/create` | JWT |
| POST | `/api/v1/user/admin/update/:id` | JWT |

### 2.4 角色 `/api/v1/role`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/role/menu-privilege-tree` | RBAC |
| POST | `/api/v1/role` | RBAC |
| GET | `/api/v1/role` | RBAC |
| GET | `/api/v1/role/:id/data-scope` | RBAC |
| PUT | `/api/v1/role/:id/data-scope` | RBAC |
| GET | `/api/v1/role/:id` | RBAC |
| PATCH | `/api/v1/role/:id` | RBAC |
| DELETE | `/api/v1/role/:id` | RBAC |

### 2.5 部门 `/api/v1/dept`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/dept` | RBAC |
| GET | `/api/v1/dept/tree` | RBAC |
| GET | `/api/v1/dept` | RBAC |
| GET | `/api/v1/dept/:id` | RBAC |
| PATCH | `/api/v1/dept/:id` | RBAC |
| DELETE | `/api/v1/dept/:id` | RBAC |

### 2.6 权限 `/api/v1/privilege`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/privilege` | RBAC |
| GET | `/api/v1/privilege` | RBAC |
| GET | `/api/v1/privilege/:id` | RBAC |
| PATCH | `/api/v1/privilege/:id` | RBAC |
| DELETE | `/api/v1/privilege/:id` | RBAC |

### 2.7 菜单 `/api/v1/admin`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/admin/menu` | RBAC |
| POST | `/api/v1/admin/menu` | RBAC |
| PATCH | `/api/v1/admin/menu` | RBAC |
| GET | `/api/v1/admin/menu/detail` | RBAC |
| DELETE | `/api/v1/admin/menu` | JWT |

### 2.8 敏感词 `/api/v1/sensitive-word`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/sensitive-word` | RBAC |
| POST | `/api/v1/sensitive-word` | RBAC |
| POST | `/api/v1/sensitive-word/batch` | RBAC |
| GET | `/api/v1/sensitive-word/hit` | RBAC |
| POST | `/api/v1/sensitive-word/hit/:id/approve` | RBAC |
| POST | `/api/v1/sensitive-word/hit/:id/reject` | RBAC |
| PATCH | `/api/v1/sensitive-word/:id` | RBAC |
| DELETE | `/api/v1/sensitive-word/:id` | RBAC |

### 2.9 操作日志 `/api/v1/operation-log`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/operation-log` | RBAC |
| DELETE | `/api/v1/operation-log/clean` | RBAC |

---

## 3. blog-service（`:5001`）

> 注册：`services/blog/internal/handler/register_blog.go`、`ws_handler.go`  
> Gateway：除 user/rpg 前缀外的 `/api/v1/*` 默认代理到此服务。

### 3.1 健康、静态、WebSocket

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/health` | 公开 | 健康检查 |
| GET | `/api/v1/health` | 公开 | 健康检查 |
| GET | `/metrics` | 公开 | Prometheus |
| GET | `/static/*` | 公开 | 上传静态文件 |
| GET | `/realtime` | JWT（WS） | WebSocket Hub |

### 3.2 站内通知 `/api/v1/notification`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/notification/list` | JWT |
| GET | `/api/v1/notification/unread-count` | JWT |
| GET | `/api/v1/notification/since` | JWT |
| PATCH | `/api/v1/notification/read` | JWT |

### 3.3 开发调试 `/api/v1/dev`（非 production 慎用）

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/dev/ws-push` | JWT |
| POST | `/api/v1/dev/ws-push-redis` | JWT |
| POST | `/api/v1/dev/event-publish` | JWT |

### 3.4 文章 `/api/v1/article`

| 方法 | 路径 | 鉴权 | Gateway 备注 |
|------|------|------|--------------|
| POST | `/api/v1/article/list` | 公开 | 代理 |
| GET | `/api/v1/article/info` | 公开 | **BFF-gRPC**（gateway 拦截，不经 blog HTTP） |
| POST | `/api/v1/article/create` | JWT | 代理 |
| POST | `/api/v1/article/edit` | JWT | 代理 |
| DELETE | `/api/v1/article/delete` | JWT | 代理 |
| POST | `/api/v1/article/views` | 公开 | 代理 |
| POST | `/api/v1/article/likes` | 公开 | 代理 |
| PATCH | `/api/v1/article/disabled` | 公开 | 代理 |
| PATCH | `/api/v1/article/topping` | 公开 | 代理 |
| GET | `/api/v1/article/my-list` | JWT | 代理 |
| GET | `/api/v1/article/archives` | 公开 | 代理 |
| GET | `/api/v1/article/related` | 公开 | 代理 |
| GET | `/api/v1/article/author-stats` | JWT | 代理 |
| GET | `/api/v1/article/statistics` | 公开 | 代理 |

### 3.5 分类 `/api/v1/category`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/category` | JWT |
| GET | `/api/v1/category` | 公开 |
| GET | `/api/v1/category/:id` | 公开 |
| PATCH | `/api/v1/category/:id` | JWT |
| DELETE | `/api/v1/category/:id` | JWT |

### 3.6 标签 `/api/v1/tag`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/tag` | JWT |
| GET | `/api/v1/tag` | 公开 |
| GET | `/api/v1/tag/:id/article` | 公开 |
| GET | `/api/v1/tag/:id` | 公开 |
| PATCH | `/api/v1/tag/:id` | JWT |
| DELETE | `/api/v1/tag/:id` | JWT |

### 3.7 评论 `/api/v1/comment`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/comment/create` | JWT |
| DELETE | `/api/v1/comment/delete` | JWT |
| GET | `/api/v1/comment/findAll` | 公开 |
| GET | `/api/v1/comment/admin` | RBAC |
| GET | `/api/v1/comment/my-list` | JWT |
| GET | `/api/v1/comment/on-my-articles` | JWT |

### 3.8 回复 `/api/v1/reply`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/reply/create` | JWT |
| DELETE | `/api/v1/reply/delete` | JWT |
| GET | `/api/v1/reply/findAll` | 公开 |
| GET | `/api/v1/reply/my-list` | JWT |

### 3.9 点赞 `/api/v1/like`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/like` | 公开 |
| GET | `/api/v1/like/check` | JWT |
| GET | `/api/v1/like/my-ids` | JWT |

### 3.10 收藏 `/api/v1/collect`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/collect` | JWT |
| DELETE | `/api/v1/collect/:id` | JWT |
| GET | `/api/v1/collect/list` | JWT |
| GET | `/api/v1/collect/check` | JWT |
| GET | `/api/v1/collect/count` | 公开 |

### 3.11 留言板 `/api/v1/msgboard`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/msgboard` | 公开 |
| GET | `/api/v1/msgboard` | 公开 |
| POST | `/api/v1/msgboard/delete` | JWT |

### 3.12 友链 `/api/v1/link`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/link` | 公开 |
| GET | `/api/v1/link` | 公开 |
| GET | `/api/v1/link/:id` | 公开 |
| PATCH | `/api/v1/link/:id` | 公开 |
| DELETE | `/api/v1/link` | JWT |

### 3.13 大文件 `/api/v1/file`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/file/uploadBigFile` | JWT |
| POST | `/api/v1/file/uploadBigFile/merge` | JWT |
| GET | `/api/v1/file/uploadBigFile/checkFile` | JWT |

### 3.14 资源 `/api/v1/resources`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/resources/daily-img` | 公开 |
| GET | `/api/v1/resources/baidutongji` | JWT+RBAC | 百度统计 OpenAPI 代理（Plan 16） |
| GET | `/api/v1/resources/weather` | 公开 |
| POST | `/api/v1/resources/uploadFile` | JWT |
| POST | `/api/v1/resources/upload-media` | JWT |
| POST | `/api/v1/resources/upload-media/register-avatar` | 公开 |
| GET | `/api/v1/resources/files` | 公开 |
| GET | `/api/v1/resources/register-avatars` | 公开 |
| GET | `/api/v1/resources/file/:id` | 公开 |
| DELETE | `/api/v1/resources/file` | JWT |
| POST | `/api/v1/resources/folder` | 公开 |
| PATCH | `/api/v1/resources/file` | 公开 |

### 3.15 定时任务 `/api/v1/scheduled-task`（Plan 12）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/v1/scheduled-task/tasks` | JWT+RBAC | 已注册任务列表（含 running） |
| GET | `/api/v1/scheduled-task/tasks/all` | JWT+RBAC | 分页任务定义 |
| GET | `/api/v1/scheduled-task/tasks/:id` | JWT+RBAC | 单个任务 |
| POST | `/api/v1/scheduled-task/tasks` | JWT+RBAC | 创建任务 |
| PATCH | `/api/v1/scheduled-task/tasks/:id` | JWT+RBAC | 更新 cron/启用等 |
| DELETE | `/api/v1/scheduled-task/tasks/:id` | JWT+RBAC | 删除任务 |
| GET | `/api/v1/scheduled-task/status/:taskName` | JWT+RBAC | 运行状态 |
| POST | `/api/v1/scheduled-task/trigger/:taskName` | JWT+RBAC | 手动触发 |
| POST | `/api/v1/scheduled-task/stop/:taskName` | JWT+RBAC | 停止 cron |
| POST | `/api/v1/scheduled-task/start/:taskName` | JWT+RBAC | 启动 cron |
| PATCH | `/api/v1/scheduled-task/log-recording/:taskName` | JWT+RBAC | 切换执行日志 |
| POST | `/api/v1/scheduled-task/cache/clear-permissions` | JWT **超管** | 清 RBAC Redis |
| POST | `/api/v1/scheduled-task/cache/refresh-tongji-token` | JWT **超管** | 刷新百度 token（Plan 16） |
| GET | `/api/v1/scheduled-task` | JWT+RBAC | 分页执行日志 |
| GET | `/api/v1/scheduled-task/backups` | JWT **超管** | 备份文件列表 |
| GET | `/api/v1/scheduled-task/backups/download` | JWT **超管** | 下载最新备份 |
| GET | `/api/v1/scheduled-task/backups/:fileName/download` | JWT **超管** | 下载指定备份 |

> Gateway：`scheduled-task/*` 显式代理 → blog-service。

### 3.16 RAG 知识库 `/api/v1/rag`（Plan 15）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/v1/rag/quota` | JWT | 今日问答配额 |
| GET | `/api/v1/rag/status` | 公开 | enabled、chunkCount、embedding 模式 |
| POST | `/api/v1/rag/query-stream` | JWT | SSE（AI SDK UI Message Stream） |

### 3.17 RAG 管理 `/api/v1/admin/rag`（Plan 15）

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/api/v1/admin/rag/stats` | JWT+RBAC | 概览统计 |
| GET | `/api/v1/admin/rag/query-logs` | JWT+RBAC | 查询日志分页 |
| GET | `/api/v1/admin/rag/index-jobs` | JWT+RBAC | 索引任务 |
| GET | `/api/v1/admin/rag/chunks` | JWT+RBAC | 知识块列表 |
| POST | `/api/v1/admin/rag/reindex` | JWT+RBAC | 全量或单篇重建 |

> Gateway：`rag/*`、`admin/rag/*` 显式代理 → blog-service（`admin/rag` 优先于 user-service 的 `admin/` 路由）。

---

## 4. rpg-service（`:5003`）

> 注册：`services/rpg/internal/handler/register.go`  
> Gateway：`rpg/*`、`admin/rpg/*`、`pay/*`、`user/public/*`、`rpg/public/*` → 代理到此服务（**例外**：`GET /user/public/:uid` 由 gateway BFF-gRPC 处理）。

### 4.1 健康与静态

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/health` | 公开 |
| GET | `/api/v1/health` | 公开 |
| GET | `/metrics` | 公开 |
| GET | `/static/*` | 公开 |

### 4.2 C 端 RPG `/api/v1/rpg`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/rpg/sign` | JWT |
| GET | `/api/v1/rpg/sign-info` | JWT |
| GET | `/api/v1/rpg/status` | JWT |
| GET | `/api/v1/rpg/hit-records` | JWT |
| GET | `/api/v1/rpg/level-rewards` | 公开 |
| GET | `/api/v1/rpg/leaderboard` | 公开 |
| GET | `/api/v1/rpg/ban-status` | JWT |
| GET | `/api/v1/rpg/my-achievements` | JWT |
| GET | `/api/v1/rpg/quests` | 公开 |
| GET | `/api/v1/rpg/my-quests` | JWT |
| POST | `/api/v1/rpg/quest/claim` | JWT |
| GET | `/api/v1/rpg/my-buffs` | JWT |
| POST | `/api/v1/rpg/buff/:id/activate` | JWT |
| POST | `/api/v1/rpg/buff/:id/deactivate` | JWT |
| GET | `/api/v1/rpg/lottery/pool` | 公开 |
| POST | `/api/v1/rpg/lottery/draw` | JWT |
| GET | `/api/v1/rpg/lottery/history` | JWT |
| GET | `/api/v1/rpg/lottery/tickets` | JWT |
| GET | `/api/v1/rpg/inventory` | JWT |
| GET | `/api/v1/rpg/loadout` | JWT |
| POST | `/api/v1/rpg/loadout/equip` | JWT |
| POST | `/api/v1/rpg/loadout/unequip` | JWT |
| GET | `/api/v1/rpg/pets` | JWT |
| GET | `/api/v1/rpg/pets/catalog` | 公开 |
| POST | `/api/v1/rpg/pets/summon` | JWT |
| POST | `/api/v1/rpg/pets/exchange` | JWT |
| PATCH | `/api/v1/rpg/pets/:id/rename` | JWT |
| GET | `/api/v1/rpg/activities/current` | 公开 |
| POST | `/api/v1/rpg/activities/share-poster` | JWT |
| GET | `/api/v1/rpg/weather-buff` | 公开 |
| GET | `/api/v1/rpg/guilds` | 公开 |
| GET | `/api/v1/rpg/guild/my` | JWT |
| GET | `/api/v1/rpg/guild/:id` | 公开 |
| POST | `/api/v1/rpg/guild/create` | JWT |
| POST | `/api/v1/rpg/guild/join` | JWT |
| POST | `/api/v1/rpg/guild/leave` | JWT |
| POST | `/api/v1/rpg/article/tip` | JWT |
| POST | `/api/v1/rpg/social/cheer` | JWT |
| POST | `/api/v1/rpg/social/egg` | JWT |
| POST | `/api/v1/rpg/social/flower` | JWT |
| POST | `/api/v1/rpg/recharge/create` | JWT |
| GET | `/api/v1/rpg/recharge/status` | JWT |

### 4.3 RPG 后台 `/api/v1/admin/rpg`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| GET | `/api/v1/admin/rpg/achievements` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/achievements` | JWT+RBAC |
| PATCH | `/api/v1/admin/rpg/achievements/:id` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/achievements/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/quests` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/quests` | JWT+RBAC |
| PATCH | `/api/v1/admin/rpg/quests/:id` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/quests/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/lottery/pool` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/lottery/pool` | JWT+RBAC |
| PATCH | `/api/v1/admin/rpg/lottery/pool/:id` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/lottery/pool/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/lottery/records` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/users` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/users/:uid/currency` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/users/:uid/currency/deduct` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/users/:uid/unban` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/users/:uid` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/stats` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/items` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/items` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/items/upload-asset` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/items/asset` | JWT+RBAC |
| PATCH | `/api/v1/admin/rpg/items/:id` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/items/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/activities` | JWT+RBAC |
| POST | `/api/v1/admin/rpg/activities` | JWT+RBAC |
| PATCH | `/api/v1/admin/rpg/activities/:id` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/activities/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/guilds` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/guilds/:id` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/guilds/:id/members` | JWT+RBAC |
| DELETE | `/api/v1/admin/rpg/guilds/:id/members/:uid` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/tips` | JWT+RBAC |
| GET | `/api/v1/admin/rpg/social-logs` | JWT+RBAC |

### 4.4 公开主页 `/api/v1/user/public` & `/api/v1/rpg/public`

| 方法 | 路径 | 鉴权 | Gateway 备注 |
|------|------|------|--------------|
| GET | `/api/v1/user/public/:uid` | 公开 | **BFF-gRPC** |
| GET | `/api/v1/user/public/:uid/articles` | 公开 | 代理→rpg |
| GET | `/api/v1/user/public/:uid/collects` | 公开 | 代理→rpg |
| GET | `/api/v1/user/public/:uid/likes` | 公开 | 代理→rpg |
| GET | `/api/v1/rpg/public/status/batch` | 公开 | 代理→rpg |
| GET | `/api/v1/rpg/public/:uid/status` | 公开 | 代理→rpg |

### 4.5 支付 `/api/v1/pay`

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/api/v1/pay/trade/create` | 公开 | 下单 |
| GET | `/api/v1/pay/trade/query` | 公开 | 查询 |
| POST | `/api/v1/pay/trade/refund` | 公开 | 退款 |
| POST | `/api/v1/pay/trade/close` | 公开 | 关单 |
| POST | `/api/v1/pay/openid` | 公开 | 微信 openid |
| POST | `/api/v1/pay/h5-open-mini` | 公开 | H5 唤起小程序 |
| POST | `/api/v1/pay/notice` | 公开 | 支付宝/微信异步通知 |

### 4.6 支付订单 `/api/v1/pay/order`

| 方法 | 路径 | 鉴权 |
|------|------|------|
| POST | `/api/v1/pay/order/create` | JWT |
| GET | `/api/v1/pay/order/list` | JWT |
| POST | `/api/v1/pay/order/refund` | JWT |
| POST | `/api/v1/pay/order/close` | JWT |
| GET | `/api/v1/pay/order/query` | JWT |
| POST | `/api/v1/pay/order/delete` | JWT |
| POST | `/api/v1/pay/order/mark-recharge-fulfilled` | JWT |

---

## 5. 内部 gRPC（不对外暴露）

> 仅供 gateway BFF 或服务间调用；默认 insecure，带 `grpcmeta` 鉴权拦截器。

### 5.1 user.v1.UserService（`:50052`）

| RPC | 调用方 | 说明 |
|-----|--------|------|
| GetUser | blog/rpg gateway | 按 ID 查用户摘要（含 dept_id） |
| GetUserBatch | blog | 批量用户 |
| VerifyToken | 内部 | JWT 校验 |
| CountUsers | gateway BFF | 用户总数（pub/stats） |
| SendSystemEmail | blog 定时任务 | 系统 HTML 邮件（SMTP 同源） |
| EvaluateContent | blog-service | 敏感词分级检测（Plan 17） |
| CreateHitRecord | blog-service | 写入敏感词命中记录（Plan 17） |
| ListActiveUserIDs | blog-service | C 端过滤锁定/禁用作者（Plan 17） |
| GetDept | blog-service | 文章列表/详情 deptName（Plan 17） |
| ResolveAccessibleDeptIDs | blog-service | admin 文章数据权限（Plan 17） |
| AssertDeptAccess | blog-service | 文章编辑/删除机构校验（Plan 17） |
| ListSensitiveWordHits | rpg-service | C 端命中记录分页（Plan 17） |

### 5.2 blog.v1.ArticleService（`:50051`）

| RPC | 调用方 | 说明 |
|-----|--------|------|
| GetArticle | — | 文章摘要（占位） |
| ListArticles | — | 列表（占位） |
| GetArticleDetail | gateway BFF | 文章详情 JSON |
| GetPubStats | gateway BFF | 文章/分类/标签计数 |
| UpdateContentModerationStatus | user 敏感词 | 审核后同步 comment/msgboard/reply |
| ListPublicCollectArticles | rpg-service | 公开主页收藏文章分页 |
| ListPublicLikeArticles | rpg-service | 公开主页点赞文章分页 |

### 5.3 rpg.v1.RpgService（`:50053`）

| RPC | 调用方 | 说明 |
|-----|--------|------|
| GetProfile | — | 等级/经验摘要 |
| GetPublicProfile | gateway BFF | 公开主页 JSON |

---

## 6. monolith 主入口（`:8000`，Nest 替换）

`make dev` / `.\scripts\dev.ps1` 在 `services/monolith` 注册**全部 HTTP 路由**（user / blog / rpg / pay / RAG / 定时任务 / WS），为 **Nest 对等验收与生产部署基准**：

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/pub/stats` | 真实 article/category/tag 计数 |
| GET | `/api/v1/rag/*` | RAG 知识库（含 LLM FC 兜底、RPG Tool） |
| GET/POST | `/api/v1/scheduled-task/*` | 定时任务 admin + 8 内置 cron |
| GET | `/api/v1/user/public/:uid/articles` | 公开主页文章分页（对齐 Nest） |
| GET | `/api/v1/article/statistics` | 后台统计大屏全量指标 |

> **§1–5** 描述 gateway + 四微服务拆分形态，供学习 gRPC BFF / 多进程部署；功能可能落后于单体，见 [`nest-parity-matrix.md`](./nest-parity-matrix.md)。详见 [`services/monolith/README.md`](../services/monolith/README.md)。

---

## 7. 维护说明

1. **改路由后须同步本文**：修改 `register_*.go` 或 gateway `app.go` / `proxy/router.go` 时更新对应章节，并执行 `make swag-all` 刷新 OpenAPI（见 [12-swagger-api-doc.md](./12-swagger-api-doc.md)）。
2. **RBAC 公开路径**：除 handler 层 JWT 外，是否在未登录时可访问还取决于 DB `x_privilege` / Redis 缓存；fallback 列表见各服务 `middleware/permission.go`。
3. **Postman 冒烟**：单体 `baseUrl=http://127.0.0.1:8000`（`monolith-smoke.ps1`）；四微服务 gateway 与单体同端口二选一。
4. **相关文档**：[22-单体服务Nest对齐补齐](./22-单体服务Nest对齐补齐.md)、[10-微服务拆分与生产上线](./10-微服务拆分与生产上线.md)、[11-微服务代码物理拆分](./11-微服务代码物理拆分.md)。
