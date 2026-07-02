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
3. Nginx：将 `jiang-xia.top.conf` 中 `/x-blog/api/v1/` 的 `proxy_pass` 从 `:5000` 改为 `:8000`（或去掉 `/x-blog-go/` 联调路径后统一走 `:8000`）
4. `curl -sf http://127.0.0.1:8000/api/v1/health`

**切流前本地验 Go 线上**：Nginx 已提供 `/x-blog-go/api/v1/` → `:8000`（见 `blog-server/deploy/nginx/README.md`「本地前端联调」），无需停 Nest。

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

## PM2 实例命名

| 服务 | PM2 name |
|------|----------|
| gateway | `BlogGo_Gateway` |
| user | `BlogGo_User` |
| blog | `BlogGo_Blog` |
| rpg | `BlogGo_Rpg` |

`deploy.local.env` 中 `DEPLOY_PM2_APPS=BlogGo_User,BlogGo_Blog,BlogGo_Rpg,BlogGo_Gateway`（启动/校验顺序：先依赖后 gateway）。

## 零停机发布（releases + current + pm2 reload）

1. 新版本解压到 `releases/YYYYMMDD-HHMMSS/`
2. `ln -sfn` 切换 `current` 软链（原子切版本）
3. `pm2 reload`（**非 delete**）按顺序：User → Blog/Rpg → Gateway  
   - `ecosystem.config.js` 的 `cwd` 固定为 `$DEPLOY_REMOTE_DIR/current`，reload 后自动读新 release 下 `./bin/*`
4. 回滚：`rollback.ps1` 指回旧 release + reload

**首次启用本机制**：若 PM2 仍绑定旧 release 绝对路径，部署脚本会 **一次性 recreate**，之后均为 reload。

> fork 单实例下 reload 为优雅重启，空窗极短；非 socket 热备意义上的绝对 0ms。

## 故障排查

### 部署后 PM2 四服务 `not online` / `waiting restart`

1. 看日志：`tail -50 /opt/jxapp/server/blog-server-go/releases/logs/user-err.log`
2. **常见：PM2 cwd 仍指向旧 release**（`pm2 show BlogGo_User` 的 `exec cwd` 不是 `.../current`）。应走 reload；若 cwd 过期脚本会一次性 recreate。手动恢复：
   ```bash
   cd /opt/jxapp/server/blog-server-go/current
   export DEPLOY_REMOTE_DIR=/opt/jxapp/server/blog-server-go
   pm2 delete BlogGo_User BlogGo_Blog BlogGo_Rpg BlogGo_Gateway
   pm2 start ecosystem.config.js --env production
   pm2 save
   ```
   若仍有旧名 `user`/`blog`/`rpg`/`gateway` 进程，一并 `pm2 delete` 清理。
3. **常见：verify 误报 not online**。Go 冷启动 + 独立 SSH 验证有竞态；`release-lib.sh` 已带重试。若 `pm2 list` 已 online 可忽略单次 verify 报错。
4. **常见：库名错误**。旧包可能连 `myblog`，Go 生产须 **`x_my_blog`**（`env.production` → `db_database = x_my_blog`）。见上文「生产库」。
