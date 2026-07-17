# 实施计划（已完成 · 历史摘要）

> Plan 01–22 **全部已交付并归档删除**。功能真相源：[`docs/nest-parity-matrix.md`](../../docs/nest-parity-matrix.md)；架构定位：[`docs/architecture.md`](../../docs/architecture.md)。  
> 完整计划正文见 git 历史（勿再新增 `docs/NN-*.md` 阶段交付文档）。

## 演进原则（回顾）

**单体先行 → 验证模块边界 → 拆 4 服务（学习）→ 代码物理拆分 → Plan 22 单体复主（Nest 替换）**

**现行部署**：monolith `:8000` = 线上（对接 uniapp）；gateway + 四微服务 = 仅本地 WSL 学习，代码不共用。

## 里程碑（已完成）

| 里程碑 | 内容 | 对应计划 |
|--------|------|----------|
| M1 基础 | 脚手架、认证登录 | 01–02 |
| M2 用户域 | RBAC、敏感词 | 03–04 |
| M3 博客域 | 文章、互动、周边 | 05–07 |
| M4 实时与 RPG | WS、事件、RPG/支付 | 08–09 |
| M5 拆分 | 四进程 + 代码物理拆分 | 10–11 |
| M6 补齐 | 定时任务、RAG、公开主页、跨服务、事件等 | 12–18 |
| M7 RPG 深度 | 文章等级、惩罚链、实时通知 | 19–21 |
| M8 单体复主 | Nest 对齐、RAG/定时任务进 monolith | 22 |

## 仍不在范围（明确不做）

微信支付 JSAPI、`/pub/ai-stream`、K8s、强制单体↔微服务 parity 等——见 [`nest-parity-matrix.md`](../../docs/nest-parity-matrix.md) §3.6。

## 新功能文档约定

变更对外 API / Nest 对等状态时，更新：

1. [`docs/api-routes.md`](../../docs/api-routes.md)
2. [`docs/nest-parity-matrix.md`](../../docs/nest-parity-matrix.md)（若影响对等）
3. 根 [`README.md`](../../README.md) 或 [`deploy/*/README.md`](../../deploy/)（若影响启动/部署）
