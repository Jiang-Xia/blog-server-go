# blog-server-go 测试

Go 原生测试体系（`testing` + build tags），由 [blog-server](../blog-server) 用例迁移。默认 API 入口 gateway `http://127.0.0.1:8000/api/v1`。

## CI / 行业规范流水线

| 阶段 | 触发 | 内容 |
|------|------|------|
| **PR** | GitHub Actions `ci.yml` | `go vet`、单元覆盖率门禁、编译四服务 |
| **main push** | 同上 + `integration-e2e` job | Ent migrate + seed + 启四服务 → 冒烟/集成/E2E |
| **手动** | Actions `workflow_dispatch` | 全量 `scripts/test-run.sh` |

### 本地一键（推荐）

```powershell
# Windows：Docker 测试库(3307/6380) + 四层测试 + 自动清理
.\scripts\test-run.ps1

# 仅 PR 级单元门禁
.\scripts\test-run.ps1 -UnitOnly

# 使用本机已有 MySQL/Redis（3306/6379）
.\scripts\test-run.ps1 -SkipDocker
```

```bash
# Linux / macOS / WSL
bash scripts/test-run.sh
make test-run
```

### 分层命令

```powershell
make test-ci          # 单元 + 覆盖率（≈ PR 门禁）
make test-coverage
make test-smoke
make test-integration
make test-e2e
```

CI 脚本目录：`scripts/ci/`（`prepare_config`、`migrate_schemas`、`seed_test_data`、启停服务）。

测试基础设施：`deploy/docker/docker-compose.test.yml`（MySQL 3307、Redis 6380）。

## 分层

| 层级 | 位置 | 命令 | 说明 |
|------|------|------|------|
| **单元** | `pkg/**/*_test.go` | `go test ./pkg/... -count=1` | 纯函数/包逻辑，无需启动服务 |
| **冒烟** | `test/smoke/` | `go test -tags=smoke ./test/smoke/... -v` | health、登录、WS ping/pong、dev 推送 |
| **集成** | `test/integration/` | `go test -tags=integration ./test/integration/... -v` | 分模块 HTTP 接口 |
| **E2E** | `test/e2e/` | `go test -tags=e2e ./test/e2e/... -v` | 跨接口业务链路（签到、评论） |

共享辅助：`test/testutil`（HTTP 客户端、`body.code` 断言、`devlogin` 登录）  
登录实现：`internal/devlogin`（`scripts/dev_login.go` 与测试共用）

## 前置

```powershell
.\scripts\dev-all.ps1          # 启动四微服务
go test -tags=smoke ./test/smoke/... -v
.\scripts\dev-all-stop.ps1
```

| 项 | 说明 |
|----|------|
| Go 1.26+ | 必须 |
| MySQL + Redis | 集成/E2E/冒烟需 bootstrap + sync-data |
| `configs/*.yaml` | `jwt.secret` 须填写（SignToken 集成测试） |
| 测试账号 | `18888888888` / `super` |

环境变量（可选）：

| 变量 | 默认 | 说明 |
|------|------|------|
| `TEST_BASE` / `DEV_LOGIN_BASE` | `http://127.0.0.1:8000` | gateway 地址 |
| `TEST_USERNAME` / `TEST_PASSWORD` | 超级管理员 | 登录账号 |
| `CONFIG_PATH` | `configs/gateway.yaml` 等 | 与 dev-all 一致 |

## 一键命令（Makefile）

```powershell
go test ./pkg/... -count=1                              # 单元
go test -tags=smoke ./test/smoke/... -count=1 -v        # 冒烟
go test -tags=integration ./test/integration/... -v   # 集成
go test -tags=e2e ./test/e2e/... -v                     # E2E

# Linux/macOS
make test-unit test-smoke test-integration test-e2e
make test-all
```

服务未启动时，集成/E2E/冒烟会 `t.Skip`（不会误报失败）。

## 与 Nest blog-server 差异

1. **响应格式**：HTTP 恒 200，断言用 `body.code`（见 `testutil.IsOK` / `IsUnauthorized`）。
2. **默认端口**：gateway `:8000`（Nest 为 `:5000`）。
3. **无 Node/newman 依赖**：Postman 集合 `deploy/postman/` 仍可用于手工验收，自动化测试已改为 Go。

## 对照 blog-server

| blog-server | blog-server-go |
|-------------|----------------|
| `npm test` / `*.spec.ts` | `go test ./pkg/...` |
| `scripts/test/test-rpg-apis.js` | `test/integration` RPG 用例 |
| `scripts/test/test-rpg-e2e.js` | `test/e2e` 签到/互动链路 |
| `scripts/ws-smoke.mjs` | `test/smoke` WS 用例 |
| `scripts/dev_login.go` | `internal/devlogin` + 脚本薄封装 |

## 目录

```
test/
├── testutil/                # HTTP 客户端、断言、登录辅助
├── smoke/smoke_test.go
├── integration/
│   ├── integration_test.go
│   ├── integration_extended_test.go
│   └── integration_rpg_write_test.go
└── e2e/
    ├── e2e_test.go
    ├── e2e_rpg_extended_test.go
    └── e2e_blog_extended_test.go
internal/devlogin/login.go
pkg/**/*_test.go
```

### 集成覆盖要点

- 博客：BFF `article/info`、相关文章、浏览量、点赞/收藏/评论/回复读接口
- RPG：多周期排行榜、公会、宠物、充值状态、admin/rpg 读列表、未授权写拦截
- 支付/用户：`pay/order/list`、privilege、user/list
- 写操作（条件执行）：穿戴/卸装、Buff 激活/停用

### E2E 覆盖要点

- 签到完整链路、评论（`/comment/create`）
- 点赞/收藏切换、BFF 文章详情
- RPG：任务/成就、宠物改名/兑换、社交加油/送花、公会加入/离开、抽奖
