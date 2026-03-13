---
title: 'roadmap'
linkTitle: 'roadmap'
description: 'clientx 路线图'
weight: 90
---

## clientx Roadmap（2026-03）

## 定位

`clientx` 是协议导向客户端层，不是重型 RPC 框架。

- 保持协议专属 API 显式（`http` 请求响应、`tcp` 流、`udp` 报文）。
- 统一工程约束而非调用形态（配置校验与默认值、错误模型、策略管线、可观测性 hook）。

## 当前快照（已更新）

本轮已完成：

- 三协议 `New` 构造统一为“标准化 + 校验 + 默认值”。
- `http/tcp/udp` 全部具备统一 `Close()` 生命周期能力。
- `http.Execute` 强制接收 `context.Context`，TCP TLS 拨号链路支持 context 取消。
- 在根包引入能力接口：
  - `clientx.Closer`
  - `clientx.Dialer`
  - `clientx.PacketListener`
- 引入策略管线基础抽象：
  - `clientx.Operation` / `OperationKind`
  - `clientx.Policy` / `PolicyFuncs`
  - `clientx.InvokeWithPolicies`
- `WithPolicies(...)` 已接入 `http/tcp/udp` 三协议客户端。
- 已在 `clientx` 落地内置 timeout-guard、retry/backoff 与 concurrency-limit policy。
- hook/policy 的 panic 隔离默认开启，单个扩展异常不会拖垮客户端数据面主流程。
- `http.Config.Retry` 已映射到统一策略管线，并为 HTTP/TCP/UDP 提供 `WithConcurrencyLimit(...)` / `WithTimeoutGuard(...)` 便捷接入。
- 已提供工程化预设包 `clientx/preset`：`NewEdgeHTTP`、`NewInternalRPC`、`NewLowLatencyUDP`（支持 preset option 覆盖）。

## 版本推进计划（执行导向）

- `v0.3.0-alpha.2`（已完成）
  - 增加内置策略模块：超时护栏、重试退避、并发限制。
  - 增加 hook/policy 的 panic 隔离（默认不影响数据面主流程）。
- `v0.3.0-beta.1`
  - 统一 operation 分类与策略元数据约定。
  - 增加与 `observabilityx` 对齐的统一遥测增强适配层。
- `v0.3.0-rc.1`
  - 补齐服务间 HTTP/TCP/UDP 的端到端 profile 示例。
  - 建立策略开销的回归与性能基线。

## 优先级建议

### P0（当前）

- 固化内置策略集合与默认组合顺序。
- 固化工程化预设默认值与覆盖规则（文档/示例一致化）。
- 补齐策略顺序、错误聚合、取消语义的测试矩阵。

### P1（下一阶段）

- 引入 context-aware hook 契约与 canonical operation 属性。
- 提供策略层幂等判定与重试分类辅助工具。
- 文档与示例切换到“能力接口组合”用法。

### P2（后续）

- 增加可选 transport 扩展点，同时保持核心轻量。
- 增加协议与策略组合的 benchmark 矩阵。

## 非目标

- 不做完全抹平协议语义的一刀切抽象。
- 不替代成熟协议 SDK 的全部能力。
- 不强绑单一遥测或重试实现。





