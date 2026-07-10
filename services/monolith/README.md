# monolith（Nest 替换主入口）

> **Plan 22 起**：单体为 **Nest 替换与本地开发主入口**（`:5000`）；与四微服务保持功能对齐，新能力优先在单体落地后再按需同步至 `services/{user,blog,rpg}/`。

## 用途

- `make dev` 本地开发与联调（`services/monolith/cmd/main.go`）
- **替换 `blog-server`（NestJS）** 的生产部署目标
- 集成测试 / E2E 全栈验收

## 与微服务关系

| 域 | 微服务目录 | monolith 状态 |
|----|-----------|---------------|
| user | `services/user/` | 进程内 `internal/user/`，功能对齐 |
| blog | `services/blog/` | 进程内 `internal/blog/` + RAG + 定时任务 |
| rpg + pay | `services/rpg/` | 进程内 `internal/rpg/` + `internal/pay/` |
| gateway | `services/gateway/` | 单体无 gateway；BFF 能力进程内等价 |

**Go `internal` 规则**：monolith **不能** import `services/*/internal/*`，故与微服务存在代码双份；Plan 22 起以 **monolith 为准** 补齐，再反向同步微服务（可选）。

## 本地启动

```bash
make dev              # 单体 :5000
# API baseUrl: http://127.0.0.1:5000/api/v1
```

四微服务联调（可选对照）：

```bash
make dev-all          # gateway :8000
make dev-all-stop
```
