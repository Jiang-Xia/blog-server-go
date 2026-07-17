# Docker 部署

## 定位

| Compose | 用途 |
|---------|------|
| **`docker-compose.monolith.yml`** | **推荐**：单体 `:8000` + uniapp + admin，本地联调 / 对照线上形态 |
| **`docker-compose.yml`** | **仅本地 WSL 学习**：四微服务 + gateway；**不上生产**，**不按服务器内存裁剪** |

## 微服务全栈（仅本地 WSL 学习）

本机或 WSL 一键拉起 **4 微服务 + MySQL + Redis + etcd + etcd-ui + Jaeger**，用于学 Kitex BFF / etcd 发现 / 多进程拆分。与 `services/monolith` **代码不共用**，功能可能落后于单体。

```bash
docker compose -f deploy/docker/docker-compose.yml up -d --build
# 或 make up
```

| 入口 | 地址 |
|------|------|
| Gateway | `http://localhost:8000` |
| Jaeger UI | `http://localhost:16686` |
| **etcd UI** | `http://localhost:8888`（已连 compose 内 `etcd:2379`） |
| etcd gRPC | `localhost:2379` |

Kitex 注册 key 前缀为 `kitex/registry-etcd/`（**无**前导 `/`），在 UI 中按此前缀浏览即可看到 `blog.user` / `blog.blog` / `blog.rpg` 实例。

仅查本机 Windows etcd（`D:\env\etcd`，未跑 Docker 栈）时可用：

```bash
# WSL / Docker Desktop：宿主机 etcd 映射为 host.docker.internal
docker run --rm -p 8888:8888 phanna/etcd-ui:1.0.0 -endpoint=host.docker.internal:2379
```

停止：`make down`（会停 etcd / etcd-ui）。

### 多实例（学习用）

默认 compose 把 `5001/5002/5003/50052` 绑到宿主机，**不能直接 `--scale`**。叠加 [`docker-compose.scale.yml`](./docker-compose.scale.yml) 取消业务服务的宿主机端口后可扩副本。

```bash
# WSL 中，仓库根目录 blog-server-go
cd /mnt/d/study/myGithub/blog-server-go

# user/blog/rpg/gateway 各扩到 3 实例（edge nginx 对外 :8000）；或 make up-scale
docker compose -f deploy/docker/docker-compose.yml \
  -f deploy/docker/docker-compose.scale.yml \
  up -d --build \
  --scale user=3 --scale blog=3 --scale rpg=3 --scale gateway=3

# 冒烟：etcd 多实例 + 反复打 pub/stats 看命中分布
bash scripts/scale-user-smoke.sh
```

说明：

| 路径 | 多实例表现 |
|------|------------|
| Kitex（etcd） | `blog.user` / `blog.blog` / `blog.rpg` 各注册多条，gateway 客户端负载均衡 |
| gateway HTTP | 不再直接暴露；由 **edge(nginx)** 反代到 `gateway:8000` 多副本 |
| blog / rpg | cron **会×N 双跑**，仅学习用 |
| MySQL | scale overlay 挂载 `initdb/`，宿主机映射改为 `3309`；**勿与 monolith compose 同时占 `:8000`** |

对外仍只暴露 gateway `:8000`；业务 HTTP/Kitex 端口不再映射到宿主机。

停止（与平时相同，两文件都要带上以免残留 scale 状态）：

```bash
make down-scale
# 或
docker compose -f deploy/docker/docker-compose.yml \
  -f deploy/docker/docker-compose.scale.yml down
```

## 单体 + uniapp + admin（本地联调 · 对齐线上）

拉起 **monolith(:8000) + MySQL + Redis + uniapp H5(:8008) + blog-admin(:9856)**，不占用宿主机已有的 3306/6379。形态与线上「单体 + 大前端」一致。

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

**线上生产**请用 **PM2 + 单体二进制**（`:8000`），见 [`deploy/pm2/README.md`](../pm2/README.md)。四微服务不部署到生产机。
