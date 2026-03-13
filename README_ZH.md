# arcgo

`arcgo` 是模块化的 Go 后端基础设施工具集，按包组织、可按需引入，并允许包间依赖组合。

[English](./README.md) | Chinese

> **文档**: 访问 [Hugo 文档站点](./docs/) 获得统一的文档体验。

## 包概览

| 包 | 描述 |
| --- | --- |
| `authx` | 基于 Authboss + Casbin 的 Opinionated 安全抽象层 |
| `collectionx` | 泛型集合与并发安全结构 |
| `configx` | 分层配置加载与校验 |
| `eventx` | 进程内强类型事件总线 |
| `httpx` | 多框架统一强类型 HTTP 路由 |
| `logx` | 结构化日志与 `slog` 互通 |
| `observabilityx` | 可选可观测性抽象（OTel/Prometheus） |

## 快速选择

- 需要容器/数据工具：从 `collectionx` 开始
- 需要基于 Authboss + Casbin 的认证/授权抽象：从 `authx` 开始
- 需要从 `.env` + 文件 + 环境变量加载配置：从 `configx` 开始
- 需要进程内带类型的事件总线：从 `eventx` 开始
- 需要跨框架的统一 HTTP 路由：从 `httpx` 开始
- 需要结构化日志和日志轮转：从 `logx` 开始
- 需要可选的遥测抽象（OTel/Prometheus）：从 `observabilityx` 开始

## 常见组合

- API 服务：`httpx + configx + logx`
- 单体内事件驱动：`eventx + logx`
- 数据结构/工具层：`collectionx + configx`

## 常用命令

```bash
go tool task fmt
go tool task lint
go tool task test
go tool task check
```

## 文档脚本（跨平台）

```bash
go run ./scripts/deploy-docs help
go run ./scripts/deploy-docs sync
go run ./scripts/deploy-docs build
go run ./scripts/deploy-docs serve
go run ./scripts/deploy-docs deploy
# 可选：
# DOCS_REMOTE=origin DOCS_BRANCH=gh-pages go run ./scripts/deploy-docs deploy
```

## clientx 示例

```bash
go run ./clientx/examples/edge_http
go run ./clientx/examples/internal_rpc_tcp
go run ./clientx/examples/low_latency_udp
```

## Git 提交前 Hook

仓库使用 `lefthook`（通过 `go tool` 管理）。

每个 clone 只需执行一次安装：

```bash
go tool task git:hooks:install
```

手动执行 hook：

```bash
go tool task git:hooks:run
```

`pre-commit` 会在提交前执行：

- `go tool task fmt`
- `go tool task lint`
```

## 说明

- 代码注释统一为英文。
- 所有文档统一在 Hugo 文档站点维护。
