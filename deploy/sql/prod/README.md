# 生产库 x_my_blog（方案 B：与 Nest myblog 分库）

Go 服务使用 **`x_my_blog`** 库 + **`x_` 表前缀**；Nest 继续使用 **`myblog`**，互不影响。

## 一次性初始化（在服务器 SSH 内）

### 0. 准备代码与 env

```bash
cd /opt/jxapp/server/blog-server-go   # 或 git clone 后进入
# 确保存在 deploy/pm2/env.production，且 db_database = x_my_blog
```

本地 Windows 改完 env 后需重新 `deploy.ps1` 上传，或直接在服务器编辑 `deploy/pm2/env.production`。

### 1. root 建库 + 授权 jxblog

**1a. 建库**（只 CREATE DATABASE，避免 GRANT 报错）：

```bash
sudo mysql -u root -p < deploy/sql/prod/001_create_x_my_blog.sql
```

**1b. 查 jxblog 实际 host**（Nest 能连 myblog 的那条）：

```bash
sudo mysql -u root -p -e "SELECT user, host FROM mysql.user WHERE user='jxblog';"
```

**1c. 对查到的 host 授权**（把 `127.0.0.1` 换成你看到的 host）：

```bash
sudo mysql -u root -p -e "GRANT ALL PRIVILEGES ON \`x_my_blog\`.* TO 'jxblog'@'127.0.0.1'; FLUSH PRIVILEGES;"
```

若报 **`You are not allowed to create a user with GRANT`**：说明 `'jxblog'@'你写的host'` 不存在，请 **GRANT 到上一步 SELECT 里已有的 host**，不要写错 `localhost` / `127.0.0.1`。

验证 jxblog 能进新库：

```bash
mysql -u jxblog -p -h 127.0.0.1 -e "USE x_my_blog; SELECT 1;"
```

### 2. 导入数据（二选一）

**方式 A：导入转换后的 SQL 备份（推荐，无需 Go）**

本机已生成：`deploy/sql/prod/x_my_blog_import_from_myblog_backup.sql`（由 `myblog_backup_1782985224776.sql` 转换，48 张 `x_*` 表）。

上传到服务器后：

```bash
mysql -u jxblog -p -h 127.0.0.1 x_my_blog < deploy/sql/prod/x_my_blog_import_from_myblog_backup.sql
```

重新转换（Windows）：

```powershell
go run scripts/transform_myblog_dump.go `
  -in D:/data/onlineBlogServeData/database/myblog_backup_1782985224776.sql `
  -out deploy/sql/prod/x_my_blog_import_from_myblog_backup.sql
```

**方式 B：Go 脚本从 myblog 克隆**

```bash
go run scripts/bootstrap_prod_db.go --env deploy/pm2/env.production
```

### 3. 验证

```bash
mysql -u jxblog -p -h 127.0.0.1 -e "SHOW TABLES FROM x_my_blog LIKE 'x_%';" | head
mysql -u jxblog -p -h 127.0.0.1 -e "SELECT COUNT(*) FROM x_my_blog.x_user;"
mysql -u jxblog -p -h 127.0.0.1 -e "SELECT COUNT(*) FROM myblog.user;"
```

### 3b. 导入后列名修复（若 RPG 接口 500）

早期 `x_my_blog_import_from_myblog_backup.sql` 曾把 `rpg_item_config.category` 误写成 `x_category`，导致 Go Ent 查询 `category` 列失败。已在服务器执行：

```bash
sudo mysql x_my_blog < deploy/sql/prod/002_fix_rpg_item_config_category.sql
```

脚本幂等，可重复执行。

### 3c. 公会成员 role 列（若 guild/my 500）

早期导入脚本将 `rpg_user_guild_member.role` 误写成 `x_role`（与权限表 `x_role` 冲突），导致 Go 查询 `role` 列失败：

```bash
sudo mysql x_my_blog < deploy/sql/prod/003_fix_rpg_user_guild_member_role.sql
```

### 4. 部署 Go 四服务

本地 Windows：

```powershell
powershell -ExecutionPolicy Bypass -File deploy/pm2/deploy.ps1
```

切流：`pm2 stop blog-server` → Nginx `:8000` → `curl health`。

## env.production 关键项

```ini
db_database = x_my_blog
db_username = jxblog
db_host = 127.0.0.1
```

修改后本地执行：

```powershell
go run ./scripts/sync_pm2_config_from_nest.go
```

再 `deploy.ps1`。

## 注意

- **不会修改/删除 `myblog`**，仅读取后写入 `x_my_blog.x_*`。
- 生产默认 **跳过 Redis FLUSHDB**；Go 稳定上线后如需刷新权限缓存：`redis-cli -n 1 FLUSHDB`（与 Nest 共用 db=1 时谨慎操作）。
