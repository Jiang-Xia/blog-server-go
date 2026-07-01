# 配置文件说明

与 NestJS `blog-server` 环境变量一一对应，便于同机切换或并行对比。

## 文件对照

| Go 本地文件（**gitignore，勿提交**） | Go 仓库模板 | Nest 对照 |
|--------------------------------------|-------------|-----------|
| `configs/monolith.yaml` | `monolith.example.yaml` | `.env.development` |
| `configs/monolith.production.yaml` | `monolith.production.example.yaml` | `deploy/pm2/env.production` |
| `configs/{user,blog,rpg,gateway}.yaml` | `*.example.yaml` | 同上（微服务拆分） |
| `configs/docker/*.yaml` | `configs/docker/*.example.yaml` | Docker 内网主机名版 |

## 首次初始化

```powershell
# 从模板生成本地 yaml（已存在的文件不会覆盖）
pwsh scripts/setup-config.ps1
```

然后按 `blog-server` 中已有真实配置填写：

- 开发：`blog-server/.env.development`
- 生产：`blog-server/deploy/pm2/env.production`

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

## 差异说明

- **MySQL 库名**：Go 本地默认 `x_my_blog`（带 `x_` 表前缀）；Nest 开发库为 `myblog`。数据同步见 `scripts/sync_data_x_my_blog.go`。
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
