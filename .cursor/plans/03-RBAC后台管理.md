# Plan 03：RBAC 后台管理（模块化单体）

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | 迁移 admin 后台 RBAC（角色/权限/部门/菜单），blog-admin 可登录并加载动态菜单 |
| **前置依赖** | [02-认证与用户登录.md](./02-认证与用户登录.md) 验收通过 |
| **周期** | ~1-1.5 周 |
| **架构形态** | 模块化单体（`internal/user/admin/` 包） |
| **里程碑** | M2 用户域 |
| **原方案章节** | 三（3.1 user-service、3.3 表归属）、五（5.4 认证授权）、8.3 第8周 |

## 模块范围

对应未来 **user-service** 的后台管理部分。

### NestJS 对照

| Go 包 | NestJS 路径 |
|-------|-------------|
| `internal/user/admin` | `blog-server/src/modules/features/admin/` + `admin/system/` |

### 负责表

`role`, `privilege`, `dept`, `menu`, `user_roles_role`

## API 路径（保持兼容）

| 路径 | 说明 |
|------|------|
| `/api/v1/admin/*` | 后台管理入口 |
| `/api/v1/admin/system/*` | 角色/权限/部门/菜单 |

## 关键实现要点

### RBAC

- 角色-权限-菜单树 CRUD
- 接口级权限校验中间件（对照 NestJS `@RequirePermission`）
- 数据权限（data-scope）若 NestJS 有实现，本阶段同步迁移

### 权限中间件

- 从 JWT ctx 读取 userID + roles
- 对照 privilege 表校验接口权限码
- 无权限返回与原项目一致的错误码

## 任务清单

- [ ] 角色 CRUD + 权限分配
- [ ] 部门树 CRUD
- [ ] 菜单树 CRUD（动态路由数据）
- [ ] 权限校验中间件接入 admin 路由
- [ ] admin 登录态与 blog-admin 前端联调
- [ ] 用户-角色关联维护

## 验收标准

- [ ] blog-admin 可登录并加载动态菜单
- [ ] RBAC：无权限用户访问 admin 接口返回对应错误码
- [ ] 角色/部门/菜单 CRUD 与 NestJS 响应格式一致
- [ ] newman 契约测试：admin 核心接口通过

### 可脚本化验收

```bash
# 需先获取 admin token（Plan 02 登录接口）
newman run deploy/postman/admin-rbac-smoke.json \
  --env-var baseUrl=http://localhost:5000 \
  --env-var token=$ADMIN_TOKEN
```

## 本计划不做

- auth / captcha / pub — Plan 02 已完成
- 敏感词、notification、operation-log — Plan 04
- 博客内容模块 — Plan 05 起
- 物理拆微服务 — Plan 10
- RAG 模块 — v3 范围外

## 生产切换提示

本计划完成后，可将 admin 路径加入 Nginx 灰度：

```
/api/v1/admin → Go
```

## 风险与注意点

| 风险 | 对策 |
|------|------|
| RBAC 权限码不一致 | 从 NestJS 数据库导出 privilege 表对照 |
| 菜单树结构差异 | 对照 blog-admin 路由配置逐字段核对 |
| data-scope 遗漏 | 检查 NestJS `security/data-scope` 是否需同步 |

## 下一步

完成验收后进入 [04-敏感词与运维骨架.md](./04-敏感词与运维骨架.md)。
