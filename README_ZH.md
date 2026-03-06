# arcgo

`arcgo` 是模块化的 Go 后端基础设施工具集，可按需引入。

[English](./README.md) | Chinese

## 文档结构

仓库采用 README 优先模式，原 `docs/` 目录已移除。  
各包文档与代码同目录维护。

## 包文档导航

| 包 | 作用 | English | Chinese | 可运行 Quickstart |
| --- | --- | --- | --- | --- |
| `collectionx` | 泛型集合与并发安全结构 | [collectionx/README.md](./collectionx/README.md) | [collectionx/README_ZH.md](./collectionx/README_ZH.md) | [collectionx/examples/quickstart](./collectionx/examples/quickstart) |
| `configx` | 分层配置加载与校验 | [configx/README.md](./configx/README.md) | [configx/README_ZH.md](./configx/README_ZH.md) | - |
| `eventx` | 进程内强类型事件总线 | [eventx/README.md](./eventx/README.md) | [eventx/README_ZH.md](./eventx/README_ZH.md) | - |
| `httpx` | 多框架统一强类型 HTTP 路由 | [httpx/README.md](./httpx/README.md) | [httpx/README_ZH.md](./httpx/README_ZH.md) | [httpx/examples/quickstart](./httpx/examples/quickstart) |
| `logx` | 结构化日志与 `slog` 互通 | [logx/README.md](./logx/README.md) | [logx/README_ZH.md](./logx/README_ZH.md) | - |

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

## 说明

- 代码注释统一为英文。
- 中文文档统一使用 `README_ZH.md`。
