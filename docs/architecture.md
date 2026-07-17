# 架构与部署定位

> NestJS [`blog-server`](../../blog-server) 的 Go 重构：**Hertz + Ent + Kitex**，对外 `/api/v1/*` + `{code,message,data}`。  
> 详细启动见根 [`README.md`](../README.md)；功能对等见 [`nest-parity-matrix.md`](./nest-parity-matrix.md)；路由见 [`api-routes.md`](./api-routes.md)。

## 双形态（定稿）

| 形态 | 端口 | 定位 |
|------|------|------|
| **`services/monolith`** | **`:8000`** | **线上唯一路径**；对接 `blog-home-uniapp` / nuxt / admin；新功能只在此落地 |
| **gateway + user / blog / rpg** | gateway `:8000`（与单体二选一） | **仅本地 WSL 学习**；与单体 **代码不共用**；不上生产；不按内存预算裁剪 |

```
线上 / 日常开发                    本地 WSL 学习（可选）
─────────────────                  ────────────────────
uniapp / nuxt / admin              make up / dev-all
        │                                  │
        ▼                                  ▼
   monolith :8000                    gateway :8000
        │                           user/blog/rpg + Nacos
   MySQL + Redis                    MySQL + Redis + Nacos
```

## 技术选型（现行）

| 类别 | 选型 |
|------|------|
| HTTP | CloudWeGo Hertz |
| ORM | Ent（表前缀 `x_`，库 `x_my_blog`） |
| 缓存 / 事件 | Redis（rueidis；Stream 领域事件） |
| 内部 RPC | **Kitex + protobuf**（仅微服务学习路径）；服务发现 **Nacos** |
| DI | wire |
| 配置 | Viper（`configs/*.yaml`） |
| 部署 | **PM2 + 单体二进制**（[`deploy/pm2/`](../deploy/pm2/)）；Docker 见 [`deploy/docker/`](../deploy/docker/) |

## 目录边界

| 路径 | 说明 |
|------|------|
| `services/monolith/` | 线上主入口；`internal/{user,blog,rpg,pay,rag,…}` |
| `services/{gateway,user,blog,rpg}/` | 学习用四进程；**勿 import 进 monolith**（Go `internal`） |
| `pkg/` | 跨服务可复用包（config、jwtauth、response、kitexmeta、kitexreg 等） |
| `proto/` | Kitex protobuf IDL；生成物在 `proto/kitex/`（`make kitex`） |

## Nest 关系

- Nest 仍可跑在 `:5000`，与 Go 单体 `:8000` **并行**。
- 切流：Nginx `/x-blog/api/v1/` → `127.0.0.1:8000`（`BlogGo_Monolith`），见 [`deploy/pm2/README.md`](../deploy/pm2/README.md)。

## 历史说明

Plan 01–22 阶段交付文档与执行计划正文已归档删除（完整内容见 git 历史）。里程碑摘要见 [`.cursor/plans/README.md`](../.cursor/plans/README.md)。
