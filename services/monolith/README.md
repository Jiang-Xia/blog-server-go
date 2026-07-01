# monolith（已 deprecated）

> **Plan 11 起**：新业务须在 `services/{user,blog,rpg}/` 开发，勿再向本目录追加功能。

## 用途

- `make dev` 单体本地调试（`services/monolith/cmd/main.go`）
- 迁移窗口内的**回滚入口**（四微服务不可用时）
- 集成测试 / 对照 Nest 行为

## 与微服务关系

| 域 | 新开发位置 | monolith 状态 |
|----|-----------|---------------|
| user | `services/user/` | 副本保留，勿双维护新功能 |
| blog | `services/blog/` | 副本保留 |
| rpg + pay | `services/rpg/` | 副本保留 |
| gateway | `services/gateway/` | 无 monolith 副本 |

**Go `internal` 规则**：monolith **不能** import `services/*/internal/*`，故迁移期存在代码双份；以微服务目录为准。

## 推荐本地联调

```bash
make dev-all          # user → blog → rpg → gateway
# API baseUrl: http://127.0.0.1:8000
make dev-all-stop
```
