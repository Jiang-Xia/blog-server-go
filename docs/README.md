# Blog-Server-Go 文档

> **入口**：根 [`README.md`](../README.md)（启动、常用命令、部署一句话）  
> **架构定位**：[`architecture.md`](./architecture.md)（单体线上 `:8000` / 微服务仅 WSL 学习）  
> **Nest 对等**：[`nest-parity-matrix.md`](./nest-parity-matrix.md)  
> Plan 01–22 阶段交付文档已删除；历史见 git / [`.cursor/plans/README.md`](../.cursor/plans/README.md)。

## 现行文档

| 文档 | 说明 |
|------|------|
| [architecture.md](./architecture.md) | 双形态定位、技术选型、目录边界 |
| [api-routes.md](./api-routes.md) | HTTP / Kitex 路由全表 |
| [nest-parity-matrix.md](./nest-parity-matrix.md) | Nest ↔ Go 功能对等矩阵 |
| [swagger.md](./swagger.md) | Swagger / OpenAPI（swaggo） |

## 部署与子域

| 文档 | 说明 |
|------|------|
| [`deploy/pm2/README.md`](../deploy/pm2/README.md) | 生产：PM2 单体 `:8000` |
| [`deploy/docker/README.md`](../deploy/docker/README.md) | Docker：单体联调 / 微服务 WSL |
| [`services/monolith/README.md`](../services/monolith/README.md) | 单体服务说明 |
| [`configs/README.md`](../configs/README.md) | 配置模板与约定 |
| [`test/README.md`](../test/README.md) | 四层测试 |

## 新功能文档同步

1. 新/改对外路由 → 更新 [`api-routes.md`](./api-routes.md)，必要时 `make swag-all`
2. 影响 Nest 对等 → 更新 [`nest-parity-matrix.md`](./nest-parity-matrix.md)
3. 影响启动/部署 → 更新根 README 或 `deploy/*/README.md`
4. **不再**按计划序号新增 `docs/NN-*.md`
