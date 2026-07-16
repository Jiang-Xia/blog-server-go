# Docker 部署

## 微服务全栈（学习用）

本地或测试机一键拉起 **4 微服务 + MySQL + Redis + Jaeger**：

```bash
docker compose -f deploy/docker/docker-compose.yml up -d --build
# 或 make up
```

Jaeger UI：`http://localhost:16686`

## 单体 + uniapp + admin（本地 WSL 试验 · 推荐）

拉起 **monolith(:8000) + MySQL + Redis + uniapp H5(:8008) + blog-admin(:9856)**，不占用宿主机已有的 3306/6379：

```bash
# 1. 从模板生成 docker 配置（已有 configs/docker/monolith.yaml 可跳过）
pwsh scripts/setup-config.ps1
# 编辑 configs/docker/monolith.yaml：mysql 用户/密码须与 compose 默认 blogdev/changeme 一致
# CORS 须含 localhost:8008（uniapp）与 localhost:9856（admin）

# 2. initdb（首次空数据卷自动执行）
#    01-schema.sql     表结构
#    02-seed-base.sql  必备基础数据（角色/菜单/权限/部门/分类标签/敏感词/超管账号）
#    对齐 Nest：blog-server/deploy/sql/init.sql 种子段
#    默认登录：18888888888 / super
#    若卷已存在需重灌：compose down -v 后再 up

# 3. WSL 中启动（无 systemd 时先确保 dockerd 在跑）
cd /mnt/d/study/myGithub/blog-server-go
sudo dockerd >/tmp/dockerd.log 2>&1 &   # 若 docker info 失败再执行
docker compose -f deploy/docker/docker-compose.monolith.yml up -d --build
# 或 make up-monolith
```

| 服务 | 地址 |
|------|------|
| Go 单体 health | http://localhost:8000/health |
| Swagger | http://localhost:8000/api/v1/doc/index.html |
| uniapp H5 | http://localhost:8008/ |
| blog-admin | http://localhost:9856/ |
| MySQL（DBeaver） | `127.0.0.1:3308` / `blogdev` / `changeme` / `x_my_blog` |

停止：`make down-monolith`（或 `docker compose -f deploy/docker/docker-compose.monolith.yml down`）。

- uniapp 构建读 `blog-home-uniapp/env/.env.docker`
- admin 构建读 `blog-admin/.env.docker`（构建时覆盖为 `.env.production`）
- 浏览器均直连 `localhost:8000`

**2G 生产机**请用 **PM2 + 二进制**，见 [`deploy/pm2/README.md`](../pm2/README.md)（复用宿主机 MySQL/Redis，内存更省）。

可选：部署前在服务器执行 `deploy/docker/setup-swap.sh` 增加 1G swap。
