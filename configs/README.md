# 配置文件说明

与 NestJS `blog-server` 环境变量一一对应，便于同机切换或并行对比。

## 文件对照

| Go 本地文件（**gitignore，勿提交**） | Go 仓库模板 | Nest 对照 |
|--------------------------------------|-------------|-----------|
| `configs/monolith.yaml` | `monolith.example.yaml` | `.env.development` |
| `configs/monolith.production.yaml` | `monolith.production.example.yaml` | `deploy/pm2/env.production` |
| `configs/{user,blog,rpg,gateway}.yaml` | `*.example.yaml` | 同上（**仅本地 WSL 微服务学习**；与 monolith 代码不共用） |
| `configs/docker/*.yaml` | `configs/docker/*.example.yaml` | Docker 内网主机名版（含 MySQL/Redis 容器） |
| `deploy/pm2/env.production` | `deploy/pm2/env.production.example` | 生产 PM2（与 Nest `deploy/pm2/env.production` **同格式**，本仓库独立维护） |
| `deploy/pm2/configs/*.yaml` | （`deploy.ps1` 从 `env.production` 自动生成） | 打包进 tar |

## 首次初始化

```powershell
# 从模板生成本地 yaml（已存在的文件不会覆盖）
pwsh scripts/setup-config.ps1
```

然后按 `blog-server` 中已有真实配置填写：

- 开发：`blog-server/.env.development`（或本仓库 `configs/*.yaml`）
- 生产 PM2：本仓库 `deploy/pm2/env.production`（格式对照 `blog-server/deploy/pm2/env.production`，可一次性拷贝后独立维护）

也可直接把你现有的 `configs/*.yaml` 保留在本地——它们已被 `.gitignore`，不会被提交。

## 字段对照（节选）

| Nest `.env` | Go YAML |
|-------------|---------|
| `auth_jwtSecret` | `jwt.secret` |
| `app_blogHome` | `app.blog_home` |
| `app_githubClientId` | `oauth.github_client_id` |
| `app_githubClientSecret` | `oauth.github_client_secret` |
| `app_emailHost/Port/User/Pass` | `mail.*` |
| `pay_alipayAppId` | `pay.alipay_app_id` |
| `pay_alipayPrivateKey` | `pay.alipay_private_key` |
| `pay_alipayPublicKey` | `pay.alipay_public_key` |
| `pay_alipayNotifyUrl` | `pay.alipay_notify_url` |
| `file_filePath` | `storage.upload_path` |
| `rag_enabled` | `rag.enabled` |
| `rag_api_key` | `rag.llm.api_key` |
| `rag_api_base_url` | `rag.llm.base_url` |
| `rag_chat_model` | `rag.llm.model` |
| `rag_embedding_api_key` | `rag.embedding.api_key` |
| `rag_embedding_api_base_url` | `rag.embedding.remote_url` |
| `rag_embedding_model` | `rag.embedding.model` |
| `rag_daily_query_limit` | `rag.daily_quota` |
| `rag_top_k` | `rag.top_k` |
| `rag_allow_local_fallback` | `rag.allow_local_fallback` |

## 微服务共享存储（仅学习环境）

本地 WSL 跑四微服务时，`user` / `blog` / `rpg` 须指向**同一 MySQL 库**与**同一 Redis 实例及 db**。线上只跑 **monolith**，用 `configs/monolith.yaml` / `deploy/pm2` 生成配置即可。

学习环境注意：

- 登录 token / 验证码在 user 写入、gateway 校验失败
- RPG 禁言、收藏点赞跨服务数据不一致

| 字段 | 本地开发约定 | CI Docker 测试 |
|------|-------------|----------------|
| `mysql.database` | `x_my_blog` | `x_my_blog` |
| `mysql.host:port` | `127.0.0.1:3306` | `127.0.0.1:3307` |
| `redis.addr` | `127.0.0.1:6379` | `127.0.0.1:6380` |
| `redis.db` | **1**（与 Nest 对齐） | **2**（隔离本地 db 1） |
| `jwt.secret` | 四服务 + gateway **完全一致** | `ci-integration-test-secret` |

`gateway` 不连 MySQL/Redis，但 `jwt.secret` 须与其它服务一致。

修改任一学习用服务的 `mysql` / `redis` / `jwt` 后，请同步其余 `configs/{user,blog,rpg}.yaml`。单体日常开发只维护 `configs/monolith.yaml`。

`test-run.ps1` 默认不启 Docker，保留本地 `configs/*.yaml`；migrate/seed 从 `configs/user.yaml` 读取 MySQL 凭据。

## 差异说明

- **MySQL 库名**：Go 本地与生产均默认 **`x_my_blog`**（带 `x_` 表前缀）；Nest 仍用 `myblog`，首次迁数据见 `scripts/sync_data_x_my_blog.go` / `deploy/sql/prod/`。
- **微服务 JWT**：`user` / `blog` / `rpg` / `gateway` 的 `jwt.secret` 须一致。
- **RSA 登录密钥**：`crypto.rsa_private_key` 为空时使用 `pkg/config` 内置开发密钥对（与 Nest `ssh.ts` 一致）；**生产须配置独立私钥**。

## 环境变量覆盖

Viper `AutomaticEnv`，示例：

```powershell
$env:OAUTH_GITHUB_CLIENT_SECRET="your-secret"
$env:PAY_ALIPAY_APP_ID="your-app-id"
$env:CONFIG_PATH="configs/monolith.production.yaml"
go run ./services/monolith/cmd/main.go
```

完整前缀见 `.cursor/rules/hertz-07-config-env.mdc`。

## 开源 / Public 仓库注意

- 仓库内**只提交** `*.example.yaml`，禁止提交含密码/密钥的 `configs/*.yaml`。
- 若历史提交曾含密钥，公开前须轮换凭据并清理 git 历史。
