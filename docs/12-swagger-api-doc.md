# Swagger / OpenAPI 接口文档

> **更新日期**：2026-07-02  
> **工具链**：[swaggo/swag](https://github.com/swaggo/swag) + [swaggo/http-swagger](https://github.com/swaggo/http-swagger)（Hertz HTTP Adaptor）

## 访问地址

开发环境默认开启（`configs/*.yaml` → `swagger.enabled: true`），路径对齐 Nest **`/api/v1/doc`**：

| 服务 | 端口 | Swagger UI | OpenAPI JSON |
|------|------|------------|--------------|
| gateway | `:8000` | http://127.0.0.1:8000/api/v1/doc/index.html | `/api/v1/doc/doc.json` |
| user-service | `:5002` | http://127.0.0.1:5002/api/v1/doc/index.html | 同上 |
| blog-service | `:5001` | http://127.0.0.1:5001/api/v1/doc/index.html | 同上 |
| rpg-service | `:5003` | http://127.0.0.1:5003/api/v1/doc/index.html | 同上 |

生产环境请在配置中设置 `swagger.enabled: false`（`deploy/pm2/env.production.example` 中 `swagger_enabled = false` 语义一致）。

## 路由覆盖范围

| OpenAPI  spec | 路由数 | 说明 |
|---------------|--------|------|
| `services/user/docs/` | 60 | 用户/RBAC/敏感词/操作日志 |
| `services/blog/docs/` | 77 | 文章/互动/资源/通知/WS |
| `services/rpg/docs/` | 101 | RPG/支付/公开主页 |
| `services/gateway/docs/` | 8 | BFF 本地路由 + 健康检查 |

路由注释桩由 `docs/api-routes.md` 自动生成，与 [api-routes.md](./api-routes.md) 保持同步。Gateway 未单独列出经代理转发的 `/api/v1/*`，请查阅对应微服务 spec。

**不在 Swagger 中的例外**（与 Nest 一致）：

- WebSocket `/realtime`（Upgrade，非 JSON 包装）
- 支付异步通知 `/pay/notice`（第三方回调格式）
- Gateway 反向代理通配 `/api/v1/*`

## 维护流程

1. 修改 `services/*/internal/handler/register_*.go` 或 gateway 路由
2. 同步更新 [`docs/api-routes.md`](./api-routes.md)
3. 重新生成文档：

```powershell
make swag-all
# 或分服务：make swag-user / swag-blog / swag-rpg / swag-gateway
```

4. 提交 `internal/apidoc/routes_gen.go` 与各 `services/*/docs/` 生成物

## 目录结构

```
pkg/apidoc/           # Swagger UI 挂载、通用响应类型
services/{svc}/internal/apidoc/
  doc.go              # @title / @BasePath / BearerAuth（swag 入口）
  routes_gen.go       # 自动生成路由注释桩（勿手改）
services/{svc}/docs/  # swag init 输出（docs.go / swagger.json / swagger.yaml）
scripts/gen_swag_apidoc.go
```

## 响应格式说明

文档中 `@Success 200` 使用 `apidoc.SuccessBody`，与 `pkg/response.Body` 一致：

```json
{ "code": 200, "bizCode": 200, "message": "success", "data": {} }
```

HTTP 状态码恒为 **200**（业务错误亦如此，`code`/`bizCode` 表达语义）。

## 后续增强（可选）

- 为高频接口在 handler 上补充 `@Param body` 具体 DTO struct（当前多为 `object` 占位）
- Gateway 聚合多服务 OpenAPI 为单一 spec
- CI 增加 `make swag-all` + git diff 校验，防止文档漂移
