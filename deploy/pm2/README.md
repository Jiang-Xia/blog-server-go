# blog-server-go PM2 生产部署（2G 机器推荐）

对齐 `blog-server/deploy/pm2/` 的**配置格式与发布流程**；生产 env **在本仓库独立维护**，不依赖 sibling `blog-server`。

## 一条命令

```powershell
powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1
# 或 make deploy
```

## 首次配置

### 1. SSH

```bash
cp deploy/pm2/deploy.local.env.example deploy/pm2/deploy.local.env
```

### 2. 生产 env（与 blog-server 同格式，本仓库一份）

```powershell
powershell -ExecutionPolicy Bypass -File scripts/setup-config.ps1
```

会生成 `deploy/pm2/env.production`（gitignore）。填写真实值：

- 从模板手填：`deploy/pm2/env.production.example`
- **或一次性参考拷贝**（之后只改本仓库）：
  ```bash
  cp ../blog-server/deploy/pm2/env.production deploy/pm2/env.production
  ```

`deploy.ps1` 部署前自动把 `env.production` 同步为 `deploy/pm2/configs/*.yaml`，无需手改 yaml。

### 3. 服务器初始化（一次）

服务器须已安装 **nvm + pm2**（Nest 部署过通常已有）。若 `pm2: command not found`：

```bash
bash deploy/pm2/server-init.sh
# 或手动：
# curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.4/install.sh | bash
# source ~/.nvm/nvm.sh && npm install -g pm2
```

验证（SSH 登录后）：

```bash
source ~/.nvm/nvm.sh && nvm use default && pm2 -v
```

## 环境变量文件

| 文件 | Git | 说明 |
|------|-----|------|
| `deploy/pm2/env.production.example` | ✅ | 模板 |
| `deploy/pm2/env.production` | ❌ | 真实生产配置（与 Nest PM2 同 key 命名） |
| `deploy/pm2/configs/*.yaml` | ❌ | 部署时自动生成 |

## 切流

1. `pm2 stop blog-server`
2. `deploy.ps1`
3. Nginx：`5000` → `8000`
4. `curl -sf http://127.0.0.1:8000/api/v1/health`

## 回滚

```powershell
powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1 -List
powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1
```

## 生产库（方案 B：x_my_blog 与 myblog 分库）

Go 用 **`x_my_blog`**，Nest 继续 **`myblog`**。一次性初始化见 [`deploy/sql/prod/README.md`](../sql/prod/README.md)：

```bash
sudo mysql -u root -p < deploy/sql/prod/001_create_x_my_blog.sql
go run scripts/bootstrap_prod_db.go --env deploy/pm2/env.production
```

`deploy/pm2/env.production` 中须 **`db_database = x_my_blog`**。

## 字段映射（env.production → Go yaml）

| env key | Go |
|---------|-----|
| `db_*` / `redis_*` | mysql / redis |
| `auth_jwtSecret` | 四服务 `jwt.secret` |
| `serve_corsOrigins` | gateway CORS |
| `app_blogHome`、`app_github*`、`app_email*` | user |
| `pay_*`、`file_filePath` | rpg / blog |

Gateway 固定 `:8000`，代理本机 `5001/5002/5003`。
