# blog-server-go Cursor Rules

Hertz + Ent + wire 单体/微服务重构项目的 Agent 规则。从 **blog-server-go** 子目录或 **myGithub 根目录**打开工作区均生效（根目录有指针规则）。

## 规则索引

| 文件 | 说明 | 作用域 |
|------|------|--------|
| `hertz-00-core.mdc` | 技术栈、API 兼容、强制通则 | always |
| `hertz-01-architecture.mdc` | 目录分层、hz 生成物、checklist | `services/**` `pkg/**` |
| `hertz-02-go-style.mdc` | Go 风格、错误、并发、测试 | `**/*.go` |
| `hertz-03-commenting-hard-requirement.mdc` | 中文注释（与 Nest server-08 对齐） | always |
| `hertz-04-handler-contract.mdc` | Hertz handler、响应体、路由 | `**/handler/**` |
| `hertz-05-request-validation.mdc` | dto + validator | `**/dto/**` |
| `hertz-06-service-ent.mdc` | service/repo/Ent | `**/service/**` `**/repo/**` |
| `hertz-07-config-env.mdc` | Viper + env | `pkg/config` `configs/**` |
| `hertz-08-middleware.mdc` | 中间件链、JWT | `**/middleware/**` |
| `hertz-09-wire-di.mdc` | wire 装配 | `wire.go` `cmd/**` |
| `hertz-10-redis-event-ws.mdc` | Stream + WS Hub | `ws/**` `event/**` |
| `hertz-11-dev-service-cleanup.mdc` | 验证后关 dev 进程 | always |
| `hertz-12-commit-message.mdc` | Conventional Commits 中文 subject | always |
| `hertz-13-plan-docs.mdc` | 计划验收后写入 `docs/` 交付文档 | always |
| `hertz-14-dev-test-account.mdc` | 本地联调默认账号 18888888888/super | always |
| `karpathy-guidelines.mdc` | LLM 行为准则（简化为先、小 diff） | always |

## 相关文档

- [`blog-server-go-重构方案.md`](../blog-server-go-重构方案.md)
- [`.cursor/plans/`](../plans/)
- Nest 对照：[`blog-server/.cursor/rules/`](../../blog-server/.cursor/rules/)
