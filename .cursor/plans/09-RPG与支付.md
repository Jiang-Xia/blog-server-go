# Plan 09：RPG 与支付（模块化单体）

## 元信息

| 项 | 内容 |
|----|------|
| **目标** | RPG 全模块 + 支付宝充值、cron 业务任务完善、单体全量回归 |
| **前置依赖** | [08-WebSocket与事件驱动.md](./08-WebSocket与事件驱动.md) 验收通过 |
| **周期** | ~2-2.5 周 |
| **架构形态** | 模块化单体（`internal/rpg/` + `internal/blog/scheduler`） |
| **里程碑** | M4 实时RPG（收尾） |
| **原方案章节** | 五（5.7/5.8）、8.3 第11-12周 |

## 模块范围

| 域 | 对应未来服务 | NestJS 路径 |
|----|-------------|-------------|
| RPG | rpg-service | `modules/rpg/` |
| 支付 | rpg-service | `modules/pay/` |

### 负责表

- rpg 域：`rpg_*` 全部、`pay_order`

> **不在范围**：RAG 模块 — 后续扩展

## API 路径（保持兼容）

| 路径 | 说明 |
|------|------|
| `/api/v1/rpg/*` | RPG 全部接口 |
| `/api/v1/admin/rpg/*` | RPG 后台管理 |
| `/api/v1/pay/*` | 支付宝充值 |

## 关键实现要点

### RPG 子模块（对照 NestJS 22 services）

- 等级/经验、签到、任务、背包/宠物、抽奖
- 公会/赛季、打赏、成就、排行榜、活动、buff 等
- admin RPG 管理接口
- RPG 事件 → Redis Stream（Plan 08 总线）→ notification/WS

### 支付

- smartwalle/alipay SDK
- 充值下单/回调/订单查询
- pay_order 表 CRUD
- 与 RPG 充值联动

### cron 完善

- 签到重置、赛季结算、活动通知等
- 对照 NestJS scheduled-task 业务 job

## 任务清单

### Week 1-2：RPG 核心

- [ ] 等级/经验系统
- [ ] 签到
- [ ] 任务系统
- [ ] 背包/宠物
- [ ] 抽奖
- [ ] 公会/赛季
- [ ] 打赏
- [ ] admin RPG 管理接口
- [ ] RPG 事件 → Redis Stream → notification/WS

### Week 2-3：支付 + 回归

- [ ] 支付宝 SDK（smartwalle/alipay）
- [ ] 充值下单/回调/订单查询
- [ ] pay_order 表 CRUD
- [ ] cron 完善：签到重置、赛季结算等
- [ ] 全量回归：C 端 + admin + WS + RPG + 支付
- [ ] Postman/newman 契约测试全量跑通
- [ ] 与 NestJS 并行对比测试（关键接口响应 diff）

## 验收标准

- [ ] RPG 升级/打赏可收到 WS 推送
- [ ] 支付宝沙箱充值链路通过
- [ ] 单体全功能回归通过（对照 NestJS 功能清单）
- [ ] cron 业务 job 正常运行
- [ ] 附录 A API 路径在单体阶段全部可用

### 可脚本化验收

```bash
newman run deploy/postman/full-regression.json --env-var baseUrl=http://localhost:5000
make test   # 单元 + 集成子集
```

## 本计划不做

- WebSocket Hub 实现 — Plan 08 已完成
- 物理拆 4 微服务 — Plan 10
- OpenTelemetry / docker-compose 生产部署 — Plan 10
- RAG 模块 — v3 范围外

## 生产切换提示

本计划完成后，**整个 Go 单体可替代 NestJS**：

- Nginx 全量切流到 Go 单体（`:5000` 或自定义端口）
- NestJS 只读观察 1 周，确认无问题后下线
- Plan 10 是在 Go 单体内部拆 4 服务，不再涉及 NestJS

## 风险与注意点

| 风险 | 对策 |
|------|------|
| RPG 逻辑复杂 | 逐模块对照 NestJS service 迁移，写集成测试 |
| 支付宝回调 URL | 沙箱/生产 config 分离 |
| lottery/admin 大文件 | 优先核心路径，admin 可分批 |
| 全量回归遗漏 | 对照附录 A 路径清单逐项勾选 |

## 文档交付

完成验收后须在 [`docs/09-RPG与支付.md`](../../docs/09-RPG与支付.md) 记录实现要点、接口一览、验收命令与已知限制，并更新 [`docs/README.md`](../../docs/README.md) 索引。

- [ ] 文档已写入 `docs/`

## 下一步

完成验收后进入 [10-微服务拆分与生产上线.md](./10-微服务拆分与生产上线.md)。
