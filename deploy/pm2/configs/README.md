# 生产 PM2 配置（自动生成，gitignore）

`deploy.ps1` 从本仓库 **`deploy/pm2/env.production`** 同步生成四份 yaml（格式与 blog-server 相同，**独立维护，部署不依赖 blog-server 仓库**）。

首次：

```powershell
powershell -ExecutionPolicy Bypass -File scripts/setup-config.ps1
# 编辑 deploy/pm2/env.production
```

若已有 Nest 配置，可一次性拷贝后在本仓库维护：

```bash
cp ../blog-server/deploy/pm2/env.production deploy/pm2/env.production
```

预览同步：

```powershell
go run ./scripts/sync_pm2_config_from_nest.go
```
