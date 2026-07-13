# blog-server-go

[NestJS blog-server](https://github.com/Jiang-Xia/blog-server) 的 Go 重构实现：**Hertz + Ent + gRPC**，对外保持 `/api/v1/*` 与 `{code, message, data}` 响应格式，前端可无感切换。

**生产与 Nest 替换以单体为准**：`services/monolith`（`:5000`）为 **主入口**——本地开发、冒烟、切流、新功能均在此落地。

**四微服务 + gateway**（`:8000`）保留用于 **学习微服务架构**（gRPC BFF、进程拆分、gateway 代理等），**不强制与单体功能 parity**，亦不作为 Nest 替换验收基准。共享 MySQL/Redis 单库。Plan 01–22 见 [`.cursor/plans/README.md`](.cursor/plans/README.md)。

## 技术栈

| 类别 | 选型 |
|------|------|
| HTTP | [CloudWeGo Hertz](https://github.com/cloudwego/hertz) |
| ORM | [Ent](https://entgo.io/) |
| 缓存 / 事件 | Redis（[rueidis](https://github.com/redis/rueidis)） |
| 内部 RPC | gRPC + protobuf（[buf](https://buf.build/)） |
| 依赖注入 | [wire](https://github.com/google/wire) |
| 配置 | Viper（`configs/*.yaml`） |
| 日志 | zap |

## 架构概览

### 主路径：单体（Nest 替换）

```
  blog-home-nuxt / blog-admin / blog-home-uniapp
                        │
                        ▼
              ┌─────────────────┐
              │    monolith     │  :5000  全路由 + WS + RAG + 定时任务
              └────────┬────────┘
                       │
                MySQL + Redis（单库 x_my_blog）
```

| 服务 | 端口 | 职责 |
|------|------|------|
| **monolith** | **5000** | **Nest 替换主入口**；user / blog / rpg 进程内模块，与 Nest 端口二选一 |

### 可选：四微服务（架构学习）

```
                    ┌─────────────┐
                    │   gateway   │  :8000  REST 入口 + gRPC BFF
                    └──────┬──────┘
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
    ┌────────────┐  ┌────────────┐  ┌────────────┐
    │    user    │  │    blog    │  │    rpg     │
    │ :5002 gRPC │  │ :5001 + WS │  │   :5003    │
    └──────┬─────┘  └──────┬─────┘  └──────┬─────┘
           └───────────────┴───────────────┘
                           │
                    MySQL + Redis（同上）
```

| 服务 | 端口 | 职责 |
|------|------|------|
| gateway | 8000 | JWT 验签、HTTP 代理、gRPC BFF（学习用，非生产主路径） |
| user | 5002 / gRPC 50052 | 认证、RBAC、敏感词、操作日志 |
| blog | 5001 | 文章、互动、资源、通知、WebSocket `/realtime` |
| rpg | 5003 | RPG、支付、公开主页 |

路由全表见 [`docs/api-routes.md`](docs/api-routes.md)。Swagger UI 见 [`docs/12-swagger-api-doc.md`](docs/12-swagger-api-doc.md)；**日常开发**以单体 `http://127.0.0.1:5000/api/v1/doc/index.html` 为准（微服务 `:8000` 为对照可选）。

## 环境要求

| 项 | 说明 |
|----|------|
| Go | **1.26+**（见 `go.mod`） |
| MySQL | 8.x，本地开发库 `x_my_blog`，表前缀 `x_` |
| Redis | 7.x，开发默认 **db=1**（与 Nest `blog-server` 一致） |
| Node | 可选，跑 `newman` / `ws-smoke.mjs` 冒烟时需要 |
| buf | 可选，仅改 `proto/` 时需要 |
| OS | **Windows**：PowerShell + `.\scripts\*.ps1`（无 `make`）；**Linux/macOS**：`make …`（见 [常用命令](#常用命令)） |

同机联调 Nest 时，可对照 sibling 仓库 [`blog-server`](../blog-server) 的 `.env.development` 填 JWT 等配置；**仅跑 Go 服务不强制克隆该仓库**（见下文数据库说明）。

## 首次拉取与启动

> 在 **`blog-server-go` 根目录**执行。Windows 示例为主；Mac/Linux 将 `.\scripts\xxx.ps1` 换为 `make xxx`。

**一条龙（首次本地 · 推荐单体）**：

`setup-config` → 填 yaml → **一次性建表** → `dev.ps1` → `curl health`

| 步骤 | 命令 / 动作 |
|------|-------------|
| 0 | Go / MySQL / Redis 就绪 |
| 1 | `git clone` + `go mod download` |
| 2 | `.\scripts\setup-config.ps1` |
| 3 | 编辑 `configs/monolith.yaml`（MySQL、JWT；微服务 yaml 学 gateway 时再填） |
| 4 | **一次性**建库 + 建表（见 §4，与 blog-server 无关亦可） |
| 5 | `.\scripts\dev.ps1`（`:5000`） |
| 6 | `curl.exe` health + `dev_login.go` |

### 0. 前置检查

```powershell
go version          # ≥ 1.26
mysql --version     # 8.x
redis-cli ping      # PONG
```

确认 **3306 / 6379** 可用；**5000** 未被占用（学微服务时再检查 5001–5003、8000）。

若 PowerShell 提示「禁止运行脚本」，执行一次（当前用户）：

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

### 1. 克隆与 Go 依赖

```powershell
git clone <repo-url> blog-server-go
cd blog-server-go
go mod download
```

工作区在 `myGithub` monorepo 内时，路径通常为 `myGithub/blog-server-go`，与 `blog-server` 并列。

### 2. 生成本地配置

仓库**不提交**真实 `configs/*.yaml`（已在 `.gitignore`），首次须从模板生成：

```powershell
.\scripts\setup-config.ps1
```

会创建 `configs/monolith.yaml`、`configs/{user,blog,rpg,gateway}.yaml` 等（已存在则跳过）。

### 3. 填写连接信息

编辑 **`configs/monolith.yaml`** 中至少以下字段（跑四微服务时再编辑 `user/blog/rpg/gateway.yaml`，且 **`jwt.secret` 须四处一致**）：

| 字段 | 说明 |
|------|------|
| `mysql.user` / `mysql.password` | 本机 MySQL 账号（如 `jiangxia`） |
| `jwt.secret` | 自行设定字符串即可（四处 yaml 保持一致）；与 Nest 混跑时才需相同 |
| `oauth.*` / `mail.*` / `pay.*` | 按需；纯 API 联调可先留空 |

**从 Nest 迁移时**：可对照 `blog-server/.env.development` 填 JWT 等；字段映射见 [`configs/README.md`](configs/README.md)。

`crypto.rsa_private_key` 留空即可使用内置开发密钥（支持 `dev_login.go` 登录）。

### 4. 准备 MySQL 库（一次性）

#### 会不会自动建表？

**不会。** `dev-all` / 各服务启动时**不会**跑 Ent migrate，也**不会**自动执行 SQL。  
首次须人工完成 **建库 + 建表**；之后日常只启服务即可。

仓库内现状：

| 能力 | 有 / 无 |
|------|---------|
| 建空库 `x_my_blog` | ✅ `001_create_x_my_blog.sql` 或 bootstrap 内 `CREATE DATABASE` |
| 完整建表 SQL 打包在仓库 | ❌ 未提供全量 dump |
| 启动时 Ent 自动 migrate | ❌ 未接入（migrate 代码仅测试用） |
| 从**任意已有 MySQL 源库**克隆结构 | ✅ `bootstrap-db.ps1` |
| 从源库导入数据 + 测试账号 | ✅ `sync-data-x-my-blog.ps1`（可选） |

因此：**不依赖 blog-server 仓库**；但若本机 MySQL 里**没有任何可复制的源库**，仍需自备一份同 schema 的库或 SQL 备份。

Go 使用库名 **`x_my_blog`**，表前缀 **`x_`**（与 Nest 常用库 `myblog` 分离）。

#### 方式 1：从已有 MySQL 源库克隆（当前默认脚本）

源库可以是：曾跑过 Nest 的 `myblog`、同事备份、远程 dev 库等——**与是否 clone blog-server 无关**。

```powershell
# 可选：root 建库授权
mysql -u root -p < deploy/sql/local/001_create_x_my_blog.sql

# 从源库拷表结构 → x_my_blog，并生成 monolith Ent
.\scripts\bootstrap-db.ps1

# 可选：拷数据（含用户/文章；否则需自行 INSERT 或调注册接口）
.\scripts\sync-data-x-my-blog.ps1
```

源库不叫 `myblog`：

```powershell
$env:BOOTSTRAP_SOURCE_DB='你的源库名'
$env:SYNC_SOURCE_DB='你的源库名'
.\scripts\bootstrap-db.ps1
.\scripts\sync-data-x-my-blog.ps1
```

#### 方式 2：直接导入 SQL 备份

若已有 `x_my_blog` 或带 `x_` 前缀表的 **mysqldump**，导入后跳过 bootstrap：

```powershell
mysql -u jiangxia -p x_my_blog < your-backup.sql
```

无需 blog-server；确保表结构与 Ent schema 一致即可。

#### 方式 3：完全没有源库时

须先让 MySQL 里出现完整表结构，任选：

- 问同事要 **dump / 共用 dev 库**（推荐，与 blog-server 代码无关）
- 或临时跑 Nest 生成 `myblog` 后再 `bootstrap-db`（仅作造库工具）

**尚无**「零依赖、一条命令 Ent 建全表 + 种子数据」的脚本；若需要可另开任务接入 `ent migrate` + dev seed。

### 5. 启动单体（主路径）

确认 MySQL、Redis 已运行；**8000** 未被占用（Nest 仍用 **5000**，与单体可并行）。

```powershell
.\scripts\dev.ps1   # :8000
```

| 项 | 值 |
|----|-----|
| API 基址 | `http://127.0.0.1:8000` |
| 登录脚本 | 默认即 `:8000`；或 `$env:DEV_LOGIN_BASE='http://127.0.0.1:8000'` |
| 冒烟 | `.\scripts\monolith-smoke.ps1` |

### 6. 验证是否启动成功

```powershell
curl.exe -sf http://localhost:8000/api/v1/health
# 期望：{"code":200,...,"data":"ok"}

go run scripts/dev_login.go --token-only
# 期望输出 JWT（默认账号见下表）
```

前端联调示例（**推荐单体**）：

| 项目 | 配置 |
|------|------|
| `blog-home-nuxt` | API 指 `http://127.0.0.1:8000` |
| `blog-home-uniapp` | `pnpm dev`（默认 `env/.env.development` → `:8000`） |
| `blog-admin` | 同上 |
| WebSocket | `ws://127.0.0.1:8000/api/v1/realtime` |

### 7. 可选：四微服务（架构学习）

学习 gateway / gRPC BFF / 多进程部署时：

```powershell
.\scripts\dev-all.ps1        # user → blog → rpg → gateway
.\scripts\dev-all-stop.ps1   # 停止
```

| 项 | 值 |
|----|-----|
| API 基址 | `http://127.0.0.1:8000` |
| 登录脚本 | `$env:DEV_LOGIN_BASE='http://127.0.0.1:8000'` |
| 日志 | `.dev-logs/*.log` |
| 四窗口看日志 | `.\scripts\dev-all.ps1 -Windows` |

> 微服务代码可能落后于单体（如统计大屏、RAG Tool 深度等），**不以微服务 parity 作为交付标准**。Nest 对等矩阵见 [`docs/nest-parity-matrix.md`](docs/nest-parity-matrix.md)（以单体为准）。

### 常见问题

| 现象 | 处理 |
|------|------|
| `dev-all` 报端口占用 | `.\scripts\dev-all-stop.ps1` 或 `netstat` + `Stop-Process`（勿杀 3306/6379） |
| `bootstrap` 报源库无表 | MySQL 里还没有可复制的源库；用 dump 导入或指定 `BOOTSTRAP_SOURCE_DB` |
| 登录 401 | 空库未 sync 数据：跑 `sync-data` 或自行插入用户；用 `dev_login.go` 勿手写明文密码 |
| `curl` 参数报错 | Windows 用 **`curl.exe`**，不要用 PowerShell 的 `curl` 别名 |

完整冒烟见 [冒烟验收](#冒烟验收)。

## 常用命令

在 **`blog-server-go` 根目录**、PowerShell 下执行（Windows **无 `make`**）。

| 说明 | 命令 |
|------|------|
| **单体 monolith（主）** | `.\scripts\dev.ps1` |
| 单体冒烟 | `.\scripts\monolith-smoke.ps1` |
| 四微服务启 / 停（学习） | `.\scripts\dev-all.ps1` / `.\scripts\dev-all-stop.ps1` |
| 单元 + 覆盖率（PR 门禁） | `.\scripts\test-run.ps1 -UnitOnly` |
| 全量四层测试 | `.\scripts\test-run.ps1` |
| 四窗口看日志 | `.\scripts\dev-all.ps1 -Windows` |
| 初始化库 + Ent | `.\scripts\bootstrap-db.ps1` |
| 同步 Nest 数据 | `.\scripts\sync-data-x-my-blog.ps1` |
| 本地配置 | `.\scripts\setup-config.ps1` |
| JWT | `go run scripts/dev_login.go --token-only` |
| Docker 全栈（本地/CI，含 MySQL/Redis） | `docker compose -f deploy/docker/docker-compose.yml up -d --build` |
| **生产远程部署（PM2 单体 · 主）** | `make deploy-monolith` 或 `deploy.ps1 -EnvFileName deploy.monolith.local.env` |
| 生产远程部署（PM2 四微服务 · 学习） | `make deploy` 或 `deploy/pm2/deploy.ps1` |
| proto | `buf generate` |
| 整理依赖 | `go mod tidy` |

单启某服务（把 `user` 换成 `gateway` / `blog` / `rpg`）：

```powershell
$env:CONFIG_PATH='configs/user.yaml'; go run ./services/user/cmd/main.go
```

改 Ent schema 后（`user` → `blog` / `rpg` 同理）：

```powershell
cd services/user/ent; go generate ./...; cd ../../..
go run github.com/google/wire/cmd/wire@latest ./services/user/internal/app
```

编译到 `bin/`：

```powershell
$env:CGO_ENABLED=0; md bin -Force | Out-Null
'gateway','user','blog','rpg','monolith' | % { go build -ldflags="-s -w" -o "bin/$_" "./services/$_/cmd" }
```

端口被占用：`netstat -ano | findstr :8000` → `Stop-Process -Id <PID> -Force`（勿停 3306/6379）。

<details>
<summary>Linux / macOS（make 对照）</summary>

`make dev-all` · `make dev-all-stop` · `make dev` · `make bootstrap-db` · `make build` · `make deploy` · `make proto` · `make ent-gen-user` · `make wire-user` · `make up` · `make down`

</details>

## 本地测试账号

| 字段 | 值 |
|------|-----|
| 用户名 | `18888888888` |
| 密码 | `super`（须 RSA 加密，用 `dev_login.go`） |

单体默认 `:5000` 无需设 `DEV_LOGIN_BASE`。学微服务时指定 gateway：

```powershell
$env:DEV_LOGIN_BASE='http://127.0.0.1:8000'
go run scripts/dev_login.go --token-only
```

## 冒烟验收

PR 合并前至少通过单元门禁；合并 `main` 后 CI 自动跑冒烟/集成/E2E。

```powershell
# 本地全量（行业规范一键）
.\scripts\test-run.ps1 -UnitOnly   # 仅 PR 级
.\scripts\test-run.ps1             # 四层全量（Docker 测试库）

# 或已有 dev-all 时分层
go test -tags=smoke ./test/smoke/... -count=1 -v
```

完整说明见 [`test/README.md`](test/README.md) 与 [`.github/workflows/ci.yml`](.github/workflows/ci.yml)。

## 目录结构

```
blog-server-go/
├── proto/                 # user / blog / rpg gRPC 定义
├── pkg/                   # config、jwtauth、response、usersvc 等共享包
├── services/
│   ├── gateway/           # API Gateway
│   ├── user/              # 用户域 + gRPC server
│   ├── blog/              # 博客域 + WebSocket
│   ├── rpg/               # RPG + 支付
│   └── monolith/          # 单体 Nest 替换主入口（Plan 22，生产基准）
├── configs/               # 本地 yaml（gitignore，见 *.example.yaml）
├── deploy/
│   ├── docker/            # docker-compose（本地/CI 全栈）
│   ├── pm2/               # 生产：二进制 + PM2 远程部署
│   └── postman/           # newman 冒烟集合
├── docs/                  # Plan 01–16 阶段交付文档
├── scripts/               # bootstrap、dev-all、ws-smoke 等
└── blog-server-go-重构方案.md   # 架构总方案 v3
```

## 文档

| 文档 | 说明 |
|------|------|
| [`docs/README.md`](docs/README.md) | 阶段交付文档索引（Plan 01–16） |
| [`docs/api-routes.md`](docs/api-routes.md) | HTTP / gRPC 路由全表 |
| [`docs/12-swagger-api-doc.md`](docs/12-swagger-api-doc.md) | Swagger / OpenAPI（swaggo） |
| [`docs/11-微服务代码物理拆分.md`](docs/11-微服务代码物理拆分.md) | 当前微服务目录、BFF、验收命令 |
| [`blog-server-go-重构方案.md`](blog-server-go-重构方案.md) | 架构与技术选型总方案 |
| [`.cursor/plans/README.md`](.cursor/plans/README.md) | 实施计划索引 |

## 相关仓库

| 仓库 | 关系 |
|------|------|
| `blog-server` | NestJS 原版 API 与数据模型来源 |
| `blog-home-nuxt` / `blog-admin` / `blog-home-uniapp` | 前端，对接 `/api/v1/*` |

## 生产部署

**2G 机器推荐**：Go 二进制 + PM2，流程与 Nest `blog-server/deploy/pm2/` 对齐。生产 env 见 `deploy/pm2/env.production` → [`deploy/pm2/README.md`](deploy/pm2/README.md)。

**单体（生产主路径 · 已远程部署）** — 与 Nest 同端口 `:5000`，单进程 `BlogGo_Monolith`：

```powershell
cp deploy/pm2/deploy.monolith.local.env.example deploy/pm2/deploy.monolith.local.env
# 填写 SSH 与 env.production 后：
make deploy-monolith
```

**四微服务（架构学习）** — gateway `:8000`：

```powershell
make deploy
```

Nginx 切流、Nest 下线 checklist 见 [`docs/22-单体服务Nest对齐补齐.md`](docs/22-单体服务Nest对齐补齐.md)。本地全栈 Docker 见 [`deploy/docker/README.md`](deploy/docker/README.md)。
