# blog-server-go PM2 生产部署（仅单体）

对齐 `blog-server/deploy/pm2/` 的**配置格式与发布流程**；生产 env **在本仓库独立维护**，不依赖 sibling `blog-server`。

## 部署模式

| 模式 | 端口 | PM2 | 用途 |
|------|------|-----|------|
| **monolith**（**唯一上线路径**） | `:8000` | `BlogGo_Monolith` | 单进程全路由；对接 uniapp 等大前端 |
| microservices | — | — | **不上生产**；仅本地 WSL Docker / `dev-all` 学习（见 [`deploy/docker/README.md`](../docker/README.md)） |

仓库内仍保留四微服务的 `deploy.ps1` / ecosystem 脚本，便于学习对照；**线上只跑单体**。

## 一条命令

### 单体（生产 · 推荐）

```powershell
# deploy/pm2/deploy.monolith.local.env 中 DEPLOY_MODE=monolith
powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1 -EnvFileName deploy.monolith.local.env
# 或 make deploy-monolith
```

### 四微服务脚本（勿用于线上）

仅本机/学习环境需要时才执行；生产机请使用上方单体命令。

```powershell
powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1
# 或 make deploy
```

## 首次配置

### 1. SSH

**单体**：

```bash
cp deploy/pm2/deploy.monolith.local.env.example deploy/pm2/deploy.monolith.local.env
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

`deploy.ps1` 部署前自动把 `env.production` 同步为 `deploy/pm2/configs/*.yaml`（含 **monolith.yaml**），无需手改 yaml。

### 3. 服务器初始化（一次）

服务器须已安装 **nvm + pm2**（Nest 部署过通常已有）。若 `pm2: command not found`：

```bash
bash deploy/pm2/server-init.sh
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
| `deploy/pm2/configs/*.yaml` | ❌ | 部署时自动生成（含 monolith.yaml） |
| `deploy/pm2/deploy.monolith.local.env` | ❌ | 单体 SSH + `DEPLOY_MODE=monolith` |

## 切流（单体 · 生产）

1. `pm2 stop blog-server`（Nest，若仍在跑）
2. 部署单体：`deploy.ps1 -EnvFileName deploy.monolith.local.env`
3. Nginx：`/x-blog/api/v1/` 的 `proxy_pass` → **`127.0.0.1:8000`**（`BlogGo_Monolith`）
4. 验证：`curl -sf http://127.0.0.1:8000/api/v1/health`

**切流前本地验 Go 线上**：Nginx `/x-blog-go/api/v1/` → `:8000`（见 `blog-server/deploy/nginx/README.md`「本地前端联调」）

## 回滚

```powershell
powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1 -EnvFileName deploy.monolith.local.env -List
powershell -ExecutionPolicy Bypass -File deploy/pm2/rollback.ps1 -EnvFileName deploy.monolith.local.env
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
| `auth_jwtSecret` | monolith `jwt.secret` |
| `serve_corsOrigins` | monolith CORS |
| `app_blogHome`、`app_github*`、`app_email*` | monolith |
| `pay_*`、`file_filePath` | monolith |
| `rag_*` | monolith |

生产对外 **仅 monolith `:8000`**；uniapp / admin / home 均指向该端口。

## PM2 实例命名

| 服务 | PM2 name | 端口 |
|------|----------|------|
| **monolith** | **`BlogGo_Monolith`** | **8000** |

单体：`DEPLOY_PM2_APPS=BlogGo_Monolith`，`DEPLOY_ECOSYSTEM_FILE=ecosystem.monolith.config.js`。

## 零停机发布（releases + current + pm2 reload）

1. 新版本解压到 `releases/YYYYMMDD-HHMMSS/`
2. `ln -sfn` 切换 `current` 软链（原子切版本）
3. `pm2 reload` `BlogGo_Monolith`（**非 delete**）；`ecosystem.monolith.config.js` 的 `cwd` 固定为 `$DEPLOY_REMOTE_DIR/current`
4. 回滚：`rollback.ps1` 指回旧 release + reload

**首次启用本机制**：若 PM2 仍绑定旧 release 绝对路径，部署脚本会 **一次性 recreate**，之后均为 reload。

> fork 单实例下 reload 为优雅重启，空窗极短；非 socket 热备意义上的绝对 0ms。

## 故障排查

### 部署后 PM2 `not online` / `waiting restart`

1. 看日志：`tail -50 /opt/jxapp/server/blog-server-go/logs/monolith-err.log`
2. **常见：PM2 cwd 仍指向旧 release**。应走 reload；若 cwd 过期脚本会一次性 recreate。手动恢复：
   ```bash
   cd /opt/jxapp/server/blog-server-go/current
   export DEPLOY_REMOTE_DIR=/opt/jxapp/server/blog-server-go
   pm2 delete BlogGo_Monolith
   pm2 start ecosystem.monolith.config.js --env production
   pm2 save
   ```
3. **常见：库名错误**。Go 生产须 **`x_my_blog`**（`env.production` → `db_database = x_my_blog`）。见上文「生产库」。
