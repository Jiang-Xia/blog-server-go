# Plan 13：RPG 后台补全与社区禁言联动

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 补齐 Admin RPG 写操作 stub；实现 BanGuard；敏感词 hit 审核联动 comment/msgboard/reply 状态 |
| **前置依赖** | [17-微服务跨服务协作补齐.md](./17-微服务跨服务协作补齐.md) + [18-领域事件发布补齐.md](./18-领域事件发布补齐.md) 验收通过（敏感词联动与 HP/禁言依赖跨服务 gRPC 与 Stream） |
| **周期** | ~2 周 |
| **架构形态** | 4 微服务（**rpg-service** admin 写操作；**blog-service** BanGuard；**user-service** 敏感词审核） |
| **里程碑** | M6 Nest 差异补齐 |
| **Nest 对照** | `modules/rpg/` admin、`guards/ban.guard.ts`、`features/sensitive-word/` |

## 背景

Plan 09 交付后 C 端 RPG 可用，但 `services/rpg/internal/app/adapters.go` 中多处 admin 写接口返回「待完善」。Plan 06 注明 **BanGuard 未实现**；Plan 04 注明敏感词 **approve/reject 未同步来源实体状态**。

## 模块范围

### A. Admin RPG 写操作（rpg-service）

对照 Nest `RpgAdminController` + 各 admin service，替换 `adapters.go` 中 `notReady(...)` stub：

| 接口域 | 待补方法（当前 stub） |
|--------|----------------------|
| 成就 | Create / Update / Delete |
| 任务 | Create / Update |
| 奖池 | Create / Update / Delete |
| 用户 | Unban |
| 物品 | Create / Update / UploadAsset / DeleteAsset |
| 活动 | Create |
| 公会 | Delete / RemoveMember |

**已实现可保留**：列表、查询、删任务、调钻、统计等。

实现要求：

- 逻辑迁入 `services/rpg/internal/rpg/admin/`，**删除 adapters 层 stub**，handler 直调 service
- 写操作校验 RBAC + 记录 operation_log（经 user-service 或本地中间件，与现网一致）
- 物品素材上传走现有 `/admin/rpg/items/upload-asset` 静态目录

### B. BanGuard（blog-service + rpg-service）

Nest：`BanGuard` → `PunishmentService.assertNotBanned(uid)`，用于：

- `POST /comment/create`
- `POST /reply/create`
- `POST /msgboard`
- `POST /rpg/sign`（rpg-service 已有路由，补 guard）

Go 方案：

1. **rpg-service** 暴露 gRPC `AssertNotBanned(uid)` 或 HTTP 内部接口（复用现有 ban 判定逻辑）
2. **blog-service** 在 comment/reply/msgboard create handler **JWT 之后**调用判定；禁言返回与 Nest 一致 HTTP 200 + 业务 code（或 403，交付文档须对照 Nest 实测）
3. **rpg-service** sign 等路由加同等校验

> HP 扣减完整惩罚链：本计划至少完成 **禁言拦截**；`PunishmentService` HP 扣减若 Nest 与 hit 创建联动，在敏感词 C 端过滤路径一并补齐（见 C）。

### C. 敏感词审核联动（user-service）

Nest `SensitiveWordService.approve/reject`：

- 更新 `sensitive_word_hit.status`
- 按 `sourceType` + `sourceId` 更新 comment / msgboard / reply 的 `status`（approved/rejected）

Go 现状：`user/internal/user/sensitive/service.go` 的 `reviewHit` 仅改 hit 表。

任务：

- 扩展 `reviewHit`：approve/reject 后调用 **blog-service gRPC** `UpdateContentModerationStatus(sourceType, sourceId, status)`
- blog-service 实现 gRPC handler，更新对应 Ent 实体
- 集成测试：创建 pending 评论 → admin approve hit → 评论 status 变 approved

### D. 敏感词命中惩罚（可选同期，建议本计划内完成）

对齐 Nest filter 结果：

- `action=2` 直接拒绝
- HP 扣减经 rpg `PunishmentService`（Stream 或 gRPC）
- 累计 hit 触发 ban 时间窗

若工期不足，交付文档「已知限制」中明确 HP/ban 自动化未完全对齐，但 **BanGuard + 审核联动** 必须验收。

## API 路径（无新增对外路径，行为补齐）

| 路径 | 变更 |
|------|------|
| `/api/v1/admin/rpg/*` 写操作 | stub → 真实 CRUD |
| `/api/v1/comment/create` 等 | 增加禁言拦截 |
| `/api/v1/sensitive-word/hit/:id/approve\|reject` | 联动来源实体 |

## 任务清单

- [ ] 梳理 `adapters.go` 全部 `notReady`，逐项对照 Nest 实现
- [ ] rpg admin service 补全 + 测试
- [ ] rpg gRPC：`AssertNotBanned`（或等价）
- [ ] blog comment/reply/msgboard 接入 BanGuard
- [ ] rpg sign 等路由接入 BanGuard
- [ ] user sensitive approve/reject → blog gRPC 状态同步
- [ ] （建议）comment/msgboard create 敏感词 HP/ban 与 Nest diff 测试
- [ ] `deploy/postman/rpg-admin-write-smoke.json` 扩展
- [ ] 更新 Swagger / `make swag-all`

## 验收标准

- [ ] blog-admin RPG 后台：创建/编辑成就、任务、奖池、物品、活动可成功并列表可见
- [ ] admin 解封用户后 BanGuard 不再拦截
- [ ] 禁言用户调用 comment/create 被拒绝（文案与 Nest 一致）
- [ ] sensitive-word hit approve 后对应评论从 pending → approved（admin 评论列表可见）
- [ ] hit reject 后评论 status=rejected
- [ ] 无 `notReady("…待完善")` 残留于 admin 写路径（grep 验收）

### 可脚本化验收

```bash
make dev-all
newman run deploy/postman/rpg-admin-write-smoke.json \
  --env-var baseUrl=http://127.0.0.1:8000 \
  --env-var token=$ADMIN_TOKEN

go test ./services/rpg/internal/rpg/admin/... ./services/user/internal/user/sensitive/... -count=1
```

## 本计划不做

- 公开主页 collects/likes — Plan 14
- `ArticleLevelService` Stream 完整消费 — 可记债到 Plan 15 或单独小 patch
- 微信支付

## 风险与注意点

| 风险 | 对策 |
|------|------|
| admin 写操作分散 | 按 Nest service 文件逐个 port，先成就/奖池/物品 |
| 跨服务事务 | hit 审核与 content 状态最终一致即可，无需 2PC |
| msgboard 无 JWT | BanGuard 仅拦截已登录用户；匿名留言走 IP 限流（现有逻辑） |

## 文档交付

[`docs/13-RPG后台补全与社区禁言联动.md`](../../docs/13-RPG后台补全与社区禁言联动.md)

- [ ] 文档已写入 `docs/`

## 下一步

[14-公开主页收藏与点赞列表.md](./14-公开主页收藏与点赞列表.md)
