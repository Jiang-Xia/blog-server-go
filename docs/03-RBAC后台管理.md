# Plan 03：RBAC 后台管理

> 对应计划：[`.cursor/plans/03-RBAC后台管理.md`](../.cursor/plans/03-RBAC后台管理.md)
>
> **交付日期**：2026-06-30  
> **架构形态**：模块化单体（`services/monolith/internal/user/admin/`）

## 交付摘要

迁移 Nest admin RBAC 至 Go 单体：角色/权限/部门/菜单 CRUD、角色数据权限、动态菜单树；路径与 blog-admin 现有 API 对齐（`/role`、`/dept`、`/privilege`、`/admin/menu`）。Plan 02 的全局 Permission 中间件自动覆盖新路由。

## 目录与模块

| 路径 | 职责 |
|------|------|
| `services/monolith/internal/user/admin/service.go` | RBAC 业务逻辑（数据权限、菜单 meta 包装） |
| `services/monolith/internal/user/repo/admin_repo*.go` | 角色/部门/菜单/权限/数据权限 CRUD（直查 MySQL） |
| `services/monolith/internal/user/repo/pagination.go` | Nest 分页结构 `page/pages` |
| `services/monolith/internal/handler/admin_handler.go` | RBAC HTTP 端点 |
| `services/monolith/internal/handler/register.go` | 注册 `/role` `/dept` `/privilege` `/admin/menu` |
| `deploy/postman/admin-rbac-smoke.json` | 契约冒烟（需 admin 账号 + RSA 密码） |

## 接口一览

### `/api/v1/role/*`（Nest RoleController）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/role/menu-privilege-tree` | 菜单+权限配置树 |
| POST | `/role` | 创建角色（privileges/menus ID 数组） |
| GET | `/role` | 分页列表（含 dataScopes/articleDataScope） |
| GET | `/role/:id/data-scope` | 查询数据权限 |
| PUT | `/role/:id/data-scope` | 更新数据权限 |
| GET | `/role/:id` | 详情（privileges/menus 为 ID 数组） |
| PATCH | `/role/:id` | 更新权限/菜单绑定 |
| DELETE | `/role/:id` | 删除角色 |

### `/api/v1/dept/*`（Nest DeptController）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/dept` | 创建部门 |
| GET | `/dept` | 分页列表（article 数据权限过滤） |
| GET | `/dept/tree` | 部门树（数据权限 + 筛选） |
| GET | `/dept/:id` | 详情 |
| PATCH | `/dept/:id` | 更新 |
| DELETE | `/dept/:id` | 删除（有子部门拒绝） |

### `/api/v1/privilege/*`（Nest PrivilegeController）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/privilege` | 创建 |
| GET | `/privilege` | 分页（含 privilegePageName） |
| GET | `/privilege/:id` | 详情 |
| PATCH | `/privilege/:id` | 更新 |
| DELETE | `/privilege/:id` | 删除 |

### `/api/v1/admin/menu/*`（Nest MenuController）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/admin/menu` | **当前用户**角色过滤后的动态菜单树（meta 包装） |
| POST | `/admin/menu` | 创建菜单 |
| PATCH | `/admin/menu` | 更新菜单 |
| GET | `/admin/menu/detail?id=` | 菜单详情 |
| DELETE | `/admin/menu?id=` | 删除菜单（需 JWT） |

### 用户-角色（Plan 02 已有）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/user/admin/create` | 超管创建用户 + roleIds/deptId |
| POST | `/user/admin/update/:id` | 超管更新用户 + roleIds |

## 本地启动与验收

```powershell
cd d:\study\myGithub\blog-server-go
go run ./services/monolith/cmd/main.go

# 构建
go build ./services/monolith/...

# 未登录访问 RBAC 接口应 401
curl.exe -s http://localhost:5000/api/v1/role

# Postman/newman（需先登录获取 token，密码须 RSA 加密同 Plan 02）
newman run deploy/postman/admin-rbac-smoke.json `
  --env-var baseUrl=http://localhost:5000 `
  --env-var token=$ADMIN_TOKEN
```

前置：MySQL `x_my_blog`、Redis DB `1`、privilege 缓存与 Nest 共用；变更权限后 `redis-cli -n 1 DEL api_permission_mappings public_api_paths role_permissions:*`。

## 与 NestJS 差异

| 项 | 说明 |
|----|------|
| 分页字段 | RBAC 列表用 Nest `page/pages`；用户列表仍为 Plan 02 的 `currentPage/totalPages` |
| 时间格式 | RBAC 实体 `createTime/updateTime` 格式化为 `YYYY-MM-DD HH:mm:ss` |
| requireOwnership | 中间件已加载该字段，**尚未**做资源所有权校验 |
| 角色 update | 与 Nest 一致，PATCH 仅刷新 privileges/menus 关联，不更新 roleName/roleDesc |

## 已知限制与后续

- 敏感词、operation-log → Plan 04
- 文章模块 data-scope 消费 → Plan 05
- newman 冒烟需本机安装 `newman` 且配置 admin 账号

## 验收勾选

- [x] 计划内任务清单已全部完成
- [x] 可脚本化验收已通过（`go build ./services/monolith/...` + 401 权限探测）
- [x] 本文档已写入 `docs/`
- [x] [`docs/README.md`](./README.md) 索引状态已更新
