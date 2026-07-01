# 配置文件说明

与 NestJS `blog-server` 环境变量一一对应，便于同机切换或并行对比。

| 文件 | 对应 Nest 来源 | 用途 |
|------|----------------|------|
| `monolith.yaml` | `.env.development` | 本地开发（默认 `CONFIG_PATH`） |
| `monolith.production.yaml` | `deploy/pm2/env.production` | 生产/PM2 切 Go 时使用 |
| `monolith.example.yaml` | `.env.example` | 无密钥模板 |

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
| `pay_alipayMiniCashierPage` | `pay.alipay_mini_cashier_page` |
| `file_filePath` | `storage.upload_path` |

## 差异说明

- **MySQL 库名**：Go 本地默认 `x_my_blog`（带 `x_` 表前缀）；Nest 开发库为 `myblog`。数据同步见 `scripts/sync_data_x_my_blog.go`。
- **支付宝 notify**：与 Nest 相同，指向公网 `https://jiang-xia.top/x-zone/api/v1/pay/notice`（本地 dev 也需支付宝能回调到该地址）。
- **GitHub Secret**：`.env.development` 仓库内为脱敏占位时，请设置环境变量 `OAUTH_GITHUB_CLIENT_SECRET` 或写入本地 yaml。

## 环境变量覆盖

Viper `AutomaticEnv`，示例：

```powershell
$env:OAUTH_GITHUB_CLIENT_SECRET="your-secret"
$env:PAY_ALIPAY_APP_ID="2021005172682351"
$env:CONFIG_PATH="configs/monolith.production.yaml"
go run ./services/monolith/cmd/main.go
```

完整前缀见 `.cursor/rules/hertz-07-config-env.mdc`。
