# Plan 13：RPG 后台补全与社区禁言联动

> 对应计划：[`.cursor/plans/13-RPG后台补全与社区禁言联动.md`](../.cursor/plans/13-RPG后台补全与社区禁言联动.md)
>
> **交付日期**：2026-07-03  
> **架构形态**：4 微服务（rpg / blog / user / gateway）

## 交付摘要

补齐 RPG 管理端全部写操作 stub；blog-service 在评论/回复创建前经 gRPC 调用 rpg `AssertNotBanned` 实现 BanGuard；user-service 敏感词 hit 审核通过后经 blog gRPC 同步 comment/msgboard/reply 的 `status`。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `services/rpg/internal/rpg/admin/write.go` | 成就/任务/奖池/物品/活动/公会/解封写操作 |
| `services/rpg/internal/rpg/itemasset/` | 物品 icon/bg 磁盘资产上传删除 |
| `services/rpg/internal/rpg/punishment/service.go` | `AdminUnban` + 既有 `AssertNotBanned` |
| `proto/rpg/v1/rpg.proto` | 新增 `AssertNotBanned` gRPC |
| `proto/blog/v1/article.proto` | 新增 `UpdateContentModerationStatus` gRPC |
| `pkg/rpgsvc/` | blog BanGuard gRPC 客户端 |
| `pkg/blogsvc/` | user 敏感词审核联动 gRPC 客户端 |
| `services/blog/internal/blog/service/moderation_service.go` | 审核状态落库 |
| `services/blog/internal/handler/comment_handler.go` | Create 前 BanGuard |
| `services/blog/internal/handler/reply_handler.go` | Create 前 BanGuard |
| `services/user/internal/user/sensitive/service.go` | approve/reject 后调 blog gRPC |
| `deploy/postman/rpg-admin-write-smoke.json` | admin 读+stats 冒烟 |

单体 `services/monolith/` 已同步 admin 写操作与 punishment/guild/repo；comment/reply Create 经 `rpgport.NewLocalBanChecker` 进程内 BanGuard。敏感词审核在单体模式仍走 Ent 直写（无需 gRPC）。

## 配置与环境

| 服务 | 配置项 | 说明 |
|------|--------|------|
| blog-service | `grpc.rpg_addr: "127.0.0.1:50053"` | BanGuard 调用 rpg gRPC |
| user-service | `grpc.blog_addr: "127.0.0.1:50051"` | 审核联动调用 blog gRPC |
| rpg-service | `storage.upload_path` / `public_prefix` | 物品素材静态目录 `{upload_path}/rpgAssets/` |

## 接口一览

**无新增对外 REST 路径**；行为补齐：

| 类型 | 端点 / RPC | 变更 |
|------|------------|------|
| REST | `/api/v1/admin/rpg/*` 写操作 | stub → 真实 CRUD |
| REST | `POST /comment/create`、`POST /reply/create` | JWT 后 BanGuard（HTTP 200 + bizCode 403） |
| REST | `POST /rpg/sign` | 已有 `AssertNotBanned`（rpg-service 内） |
| gRPC | `rpg.v1.RpgService/AssertNotBanned` | 新增 |
| gRPC | `blog.v1.ArticleService/UpdateContentModerationStatus` | 新增 |

## 本地启动与验收

```powershell
# 微服务（gateway :8000）
.\scripts\dev-all.ps1

# 管理员 token
$env:ADMIN_TOKEN = go run scripts/dev_login.go --token-only

# admin 冒烟
newman run deploy/postman/rpg-admin-write-smoke.json `
  --env-var baseUrl=http://127.0.0.1:8000 `
  --env-var token=$env:ADMIN_TOKEN

# 单元测试
go test ./services/rpg/internal/rpg/admin/... ./services/blog/internal/blog/service/... -count=1

# 无 stub 残留
rg 'notReady\(".*待完善' services/
```

## 与 NestJS 差异

| 项 | Nest | Go |
|----|------|-----|
| BanGuard 错误 | HTTP 403 或 TransformInterceptor 包装 | HTTP 200 + `bizCode=403`，文案一致 |
| 物品素材磁盘路径 | `public/rpgAssets/` | `{storage.upload_path}/rpgAssets/`（默认 `public/uploads/rpgAssets/`） |
| msgboard BanGuard | 无（匿名 IP 限流） | 同 Nest，未加 BanGuard |
| 超管解封 WS 推送 | 有 `BAN_STATUS` 事件 | 未推送 WS（仅清 DB 字段） |

## 已知限制与后续

- 敏感词命中 **HP 扣减 / 自动禁言** 完整惩罚链未完全对齐 Nest（Plan 13 建议项，留待后续 patch）。
- 超管解封未推送 WS `BAN_STATUS`。
- `ArticleLevelService` Stream 消费仍留给后续计划。

## 验收勾选

- [x] 计划内任务清单已全部完成
- [x] 单元测试已通过
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引状态已更新
