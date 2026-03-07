# httpx 后续迭代设计文档（路线 A / 统一 API + Huma 原生能力暴露）

## 1. 文档目标

本文档用于指导 `httpx` 的后续迭代方向。

本次设计明确选择以下路线：

- `httpx` 保持轻量，不演进为重型 web framework
- `httpx` 需要提供一套完整、统一、稳定的 API，用于组织 HTTP 服务能力
- `httpx` 原生集成 OpenAPI 与文档能力
- `httpx` 提供对 Huma 高级能力的完整接入面
- `httpx` 同时保留底层 Huma API 的直接访问能力
- `httpx` 同时支持：
  - adapter 原生生态 middleware
  - Huma middleware
  - 后续自带常用 middleware（待定）

因此，`httpx` 的目标不是简单做一层“薄透传”，也不是替代 Huma，而是：

> 在 Huma 之上提供一致的服务组织 API，同时完整承接 Huma 的 OpenAPI、文档、operation、middleware 与高级扩展能力。

---

## 2. 正式定位

`httpx` 的正式定位定义为：

> 一个以 Huma 为 OpenAPI 与 typed operation 核心、以 adapter 为运行时承载、以统一 API 面组织 HTTP 服务能力的轻量库。

更具体地说：

- Huma 负责：
  - typed operation
  - request/response schema
  - OpenAPI
  - JSON Schema
  - 文档生成
  - Huma middleware / group / operation modifier / response transformer
- adapter 负责：
  - 承载底层运行时
  - 暴露底层 router / handler / middleware 生态
  - 提供 listen / serve / runtime 接入能力
- httpx 负责：
  - 统一 server / group / endpoint / middleware / docs / openapi 的组织 API
  - 暴露完整的 Huma 配置能力
  - 提供必要的高级 helper
  - 同时保留直接访问 Huma API 的能力

---

## 3. 设计目标

`httpx` 后续迭代应同时满足以下目标：

### 3.1 一致性

无论底层 adapter 是 std/http、gin、echo 还是 fiber，`httpx` 都应提供一致的高层 API 来组织以下能力：

- server
- route
- group
- endpoint
- docs
- openapi
- middleware
- native handler escape hatch

`httpx` 不需要抹平所有底层差异，但必须提供统一的组织入口。

---

### 3.2 OpenAPI 原生集成

Huma 的核心价值之一是 OpenAPI 3.1 与 JSON Schema 原生集成。Huma API 本身允许直接访问并修改 OpenAPI 文档，且文档路径、OpenAPI 路径、schemas 路径都由 `huma.Config` 控制。:contentReference[oaicite:2]{index=2}

因此 `httpx` 必须将 OpenAPI 视为核心能力，而不是附属功能。

---

### 3.3 文档能力一等公民

Huma 内置支持使用第三方 renderer 生成交互式文档，默认是 Stoplight Elements，也支持 Scalar 与 Swagger UI；同时允许禁用内置 docs 后自行接管 renderer。:contentReference[oaicite:3]{index=3}

因此 `httpx` 必须在接口层完整暴露这些能力，而不是只保留一个模糊的“开启 Swagger”开关。

---

### 3.4 高级能力既要统一 API，也要保留原生出口

对于以下高级能力：

- security schemes
- components
- tags
- docs config
- global headers
- OpenAPI patch
- operation modifier
- group middleware
- response transformer

`httpx` 需要做两件事：

1. 提供一套统一的 `httpx` API
2. 同时暴露底层 Huma API / `ConfigureOpenAPI`

也就是说：

- 简单场景，用户使用 `httpx` 的统一 API
- 高级场景，用户可以直接下沉到 Huma

---

### 3.5 middleware 双层模型

Huma 提供 API 级 middleware，并在 Group 上支持 middleware、operation modifier 和 response transformer。:contentReference[oaicite:4]{index=4}

因此 `httpx` 后续必须明确 middleware 设计为双层模型：

#### 第一层：adapter native middleware
面向底层生态：

- gin middleware
- echo middleware
- chi middleware
- fiber middleware
- std/http wrapper

#### 第二层：Huma middleware
面向 operation 语义：

- Huma API middleware
- Huma Group middleware
- operation modifier
- response transformer

`httpx` 不应强行把两者抹平成同一种机制，但必须为两者提供统一而清晰的接入面。

---

## 4. 当前问题总结

### 4.1 顶层抽象存在语义漂移

当前 `httpx` 的一部分 API 更像历史遗留入口，而不是当前真实架构的准确映射。  
这会导致：

- 文档与行为不一致
- 配置点分散
- 用户不清楚某项配置应在构建时还是构建后生效

---

### 4.2 对 Huma 能力的暴露面不足

这是当前最需要补齐的点。

Huma 已经提供了：

- `API.OpenAPI()`，用于直接访问 OpenAPI 文档，并在服务启动前修改它。:contentReference[oaicite:5]{index=5}
- `huma.Config`，用于控制 OpenAPI、docs、schemas 暴露路径。:contentReference[oaicite:6]{index=6}
- `DocsRenderer`，用于切换文档 UI renderer。:contentReference[oaicite:7]{index=7}
- Group middleware / modifier / transformer。:contentReference[oaicite:8]{index=8}

但当前 `httpx` 尚未把这些能力完整而系统地暴露出来。

---

### 4.3 docs/UI 能力没有上升到正式 API

Huma 对文档 UI 的支持不只是“是否启用文档”，而是包含：

- 文档路径
- OpenAPI 路径
- schema 路径
- renderer 类型
- 自定义 renderer / 自行接管 docs 页面

这些能力必须进入 `httpx` 的正式 API 设计，而不是停留在 adapter 细节里。:contentReference[oaicite:9]{index=9}

---

### 4.4 middleware 模型尚未成形

当前若只支持 adapter middleware，会丢失 Huma operation 级能力；  
若只支持 Huma middleware，又无法接入底层框架生态。

因此 `httpx` 必须明确支持双层 middleware 模型。

---

## 5. 核心设计原则

### 5.1 `httpx` 提供统一 API，不隐藏 Huma

`httpx` 必须提供统一 API，但不能以牺牲 Huma 原生能力为代价。

也就是说：

- `httpx` 不只是 helper 集合
- `httpx` 也不是 Huma 的封闭替代层
- `httpx` 是一个统一组织层 + Huma 原生能力承接层

---

### 5.2 统一 API 与原生 API 并存

对于高级能力，统一原则如下：

- 高频能力：提供 `httpx` 统一 API
- 特殊能力：提供 `ConfigureOpenAPI`
- 全量能力兜底：提供 `HumaAPI()` 访问

这三层都要有。

---

### 5.3 构建期配置与构建后配置并存

这是本次设计中的一个关键约定。

对于 OpenAPI、docs、security、headers、components 等能力，`httpx` 应同时支持：

#### 构建期配置
在 `NewServer(...)` 或 builder/options 中配置。

适合：

- title / version / description
- docs path
- docs UI
- 默认 security scheme
- 全局 tags
- 全局 servers
- 全局 header
- OpenAPI patch

#### 构建后配置
在 `Server` 构建完成后调用。

适合：

- `ConfigureOpenAPI(...)`
- `ConfigureDocs(...)`
- `RegisterSecurityScheme(...)`
- `UseHumaMiddleware(...)`
- `PatchOpenAPI(...)`

这样用户既可以走 declarative 构建方式，也可以走 imperative 配置方式。

---

### 5.4 docs/UI 不是附属功能，而是正式能力域

Huma 支持 Stoplight Elements、Scalar 和 Swagger UI。默认 renderer 为 Stoplight Elements，也可以通过 `DocsRenderer` 切换，或将 `DocsPath` 置空后自行接管 docs 页面。:contentReference[oaicite:10]{index=10}

因此在 `httpx` 中，docs/UI 必须被视为正式能力域，而不是附属选项。

---

### 5.5 middleware 必须分层，不强制统一实现机制

`httpx` 只统一使用方式，不强制统一底层模型。

因此：

- adapter middleware 继续由 adapter 侧承载
- Huma middleware 由 `httpx` / Huma 侧承载
- 两类 middleware 在接口层并列出现
- `httpx` 不试图把 gin middleware 和 Huma middleware 编造成完全同构的一套内部模型

---

## 6. 目标能力模型

`httpx` 后续 API 设计建议围绕以下能力域展开。

### 6.1 Server 能力域

- 创建与持有 adapter
- 暴露 Huma API
- 统一 route / group / endpoint 注册
- docs / openapi / middleware 的统一入口
- 构建期与构建后配置入口

---

### 6.2 OpenAPI 能力域

- title / version / description / contact / license
- tags
- servers
- components
- security schemes
- global headers / common params
- OpenAPI patch
- 直接访问 `huma.OpenAPI`

Huma 的 `API.OpenAPI()` 明确支持直接访问并编辑 OpenAPI 文档。:contentReference[oaicite:11]{index=11}

---

### 6.3 Docs 能力域

- docs enable / disable
- docs path
- openapi path
- schemas path
- docs UI type
- custom docs renderer
- 完全自定义 docs 页面

Huma 默认 docs 使用 Stoplight Elements，并支持 Scalar / Swagger UI。:contentReference[oaicite:12]{index=12}

---

### 6.4 Middleware 能力域

#### adapter middleware
- UseAdapterMiddleware(...)
- NativeAdapter() / Router() escape hatch

#### Huma middleware
- UseHumaMiddleware(...)
- Group.UseHumaMiddleware(...)
- UseOperationModifier(...)
- UseResponseTransformer(...)

Huma Group 官方已支持 middleware、modifier、transformer。:contentReference[oaicite:13]{index=13}

---

### 6.5 Group / Endpoint 能力域

Huma Group 不只是 prefix，还支持：

- path prefixes
- middleware
- operation modifiers
- response transformers
- documentation customization。:contentReference[oaicite:14]{index=14}

因此 `httpx.Group` 不应再只是简单 prefix wrapper，而应承接这些能力。

---

## 7. 正式架构职责

### 7.1 httpx

职责：

- 统一组织 HTTP 服务能力
- 统一暴露 docs / openapi / middleware / route / group / endpoint
- 暴露完整 Huma 高级能力入口
- 提供合理的 helper
- 提供构建期与构建后配置机制

不负责：

- 替代底层 router
- 实现重型 framework 生命周期
- 抹平所有底层框架差异

---

### 7.2 adapter

职责：

- 持有底层框架对象
- 提供 listen / serve
- 暴露框架原生 middleware / native route 能力
- 提供 Huma runtime 接入点

---

### 7.3 huma integration

职责：

- typed operation
- request/response schema
- OpenAPI
- docs
- operation documentation
- Huma middleware / modifier / transformer

---

## 8. 推荐 API 设计方向

这里不写具体签名定稿，只定义能力层次。

---

### 8.1 构建期配置 API

建议 `httpx` 在构建时支持下列能力：

- `WithOpenAPI(...)`
- `WithDocs(...)`
- `WithSecurity(...)`
- `WithGlobalHeaders(...)`
- `WithOpenAPIPatch(...)`
- `WithHumaMiddleware(...)`
- `WithAdapterMiddleware(...)`

其中：

#### OpenAPI 配置
用于设置：

- title
- version
- description
- contact
- license
- servers
- tags

#### Docs 配置
用于设置：

- enabled
- docs path
- openapi path
- schemas path
- docs renderer
- custom docs renderer

#### Security 配置
用于设置：

- 注册 security schemes
- 默认 security requirements
- group / operation 默认安全策略

#### Global header 配置
用于设置：

- 公共请求头
- 公共响应头
- 公共 parameter
- OpenAPI component header / parameter

---

### 8.2 构建后配置 API

建议 `Server` 在构建完成后支持以下方法：

- `HumaAPI()`
- `ConfigureOpenAPI(fn)`
- `PatchOpenAPI(fn)`
- `ConfigureDocs(fn)`
- `RegisterSecurityScheme(name, scheme)`
- `RegisterGlobalHeader(...)`
- `UseHumaMiddleware(...)`
- `UseOperationModifier(...)`
- `UseResponseTransformer(...)`
- `UseAdapterMiddleware(...)`

这里的设计重点是：

- 构建时可以配置
- 构建后也可以继续细化
- 始终保留 Huma 兜底出口

---

### 8.3 Huma 直接访问 API

`httpx` 必须明确支持：

- `Server.HumaAPI()`
- `Group.HumaGroup()` 或等价能力
- `Server.OpenAPI()`（可选，直接返回 `*huma.OpenAPI`）

原因是 Huma 官方 API 明确允许用户通过 `API.OpenAPI()` 直接访问和修改 OpenAPI 文档。:contentReference[oaicite:15]{index=15}

---

### 8.4 Docs UI API

建议 `httpx` 正式暴露 docs UI 选择能力，例如：

- Stoplight Elements
- Scalar
- Swagger UI
- Custom

因为 Huma 官方已将这几类 renderer 视为标准支持项。:contentReference[oaicite:16]{index=16}

---

### 8.5 Group API

`httpx.Group` 建议至少支持：

- Prefix
- Huma middleware
- operation modifiers
- response transformers
- group-level docs/openapi customization
- 默认 tags
- 默认 security

这与 Huma Group 官方能力是对齐的。:contentReference[oaicite:17]{index=17}

---

## 9. 中间件设计

### 9.1 中间件分层模型

#### Adapter Middleware
用于集成底层生态，例如：

- logging
- recovery
- CORS
- gzip
- framework-native auth
- metrics

#### Huma Middleware
用于集成 operation 语义，例如：

- auth context 注入
- operation 级 tracing
- request metadata enrich
- typed request 前后处理

#### Operation Modifier
用于注册期修改 operation，例如：

- 默认 tags
- 默认 summary
- 安全定义
- response metadata
- OpenAPI 扩展字段

#### Response Transformer
用于输出阶段修改返回值。Huma Group 官方已支持此能力。:contentReference[oaicite:18]{index=18}

---

### 9.2 统一 API 设计原则

`httpx` 应统一“入口体验”，但不统一“底层实现模型”。

也就是说：

- 用户能一眼看出哪些是 adapter middleware
- 哪些是 Huma middleware
- 哪些是 operation modifier
- 哪些是 response transformer

这比强行做一个全部兼容的 `Use(...)` 更可持续。

---

### 9.3 内置 middleware

`httpx` 后续可以提供常用 middleware，但这是次级目标。

当前优先级应是：

1. 先定义 middleware 模型
2. 先把 adapter middleware 与 Huma middleware 接口做清楚
3. 再决定是否提供内置 middleware

---

## 10. OpenAPI 与文档能力的正式化

### 10.1 OpenAPI 是核心配置对象

Huma 的 `Config` 内嵌 `*OpenAPI`，并通过 `API.OpenAPI()` 暴露运行时可修改的 OpenAPI 文档。:contentReference[oaicite:19]{index=19}

因此 `httpx` 的 OpenAPI 设计应该明确两层：

#### 声明式层
- 用于构建期配置
- 适合统一项目配置

#### 编程式层
- 用于构建后 patch
- 适合灵活扩展与复杂场景

---

### 10.2 文档能力的正式化

`httpx` 应将文档能力定义为正式子系统，而不是几个散乱 option：

- 启用/禁用 docs
- docs path
- OpenAPI path
- schemas path
- renderer
- custom renderer
- custom docs page
- docs route protection strategy

同时需要注意 Huma 官方文档提到某些 router middleware 可能干扰 `/openapi.json` 或 `/openapi.yaml` 之类路径，因此 `httpx` 文档中也应提醒这类冲突。:contentReference[oaicite:20]{index=20}

---

## 11. 不建议做的事情

### 11.1 不要把 Huma 完全包死

`httpx` 必须允许用户随时回到：

- `HumaAPI()`
- `OpenAPI()`
- `ConfigureOpenAPI(...)`

否则高阶能力会永远不完整。

---

### 11.2 不要只做“几个零散 helper”

如果只提供：

- 一个 `WithSwagger()`
- 一个 `WithBearer()`
- 一个 `WithHeader()`

但没有形成系统 API，最终会比现在更乱。

因此这次设计必须承认：

> `httpx` 需要一套完整、正式的 API 面来承载 docs / openapi / security / middleware。

---

### 11.3 不要把 adapter middleware 和 Huma middleware 混成一种语义

这两者本质不同，统一使用入口可以，但不要统一实现语义。

---

### 11.4 不要在当前阶段引入重型 framework 机制

例如：

- 模块生命周期系统
- 自动 DI 容器
- 复杂 hook 总线
- 统一 runtime pipeline 引擎

路线 A 不需要这些。

---

## 12. 实施优先级

### P0：语义收敛

- 清理 no-op API
- 文档与行为对齐
- 明确 `httpx` 的正式定位
- 明确 Huma / adapter / httpx 三层职责

### P1：正式补齐核心 API 面

- docs 配置 API
- OpenAPI 配置 API
- `ConfigureOpenAPI`
- `HumaAPI()`
- security schemes API
- global headers / components API

### P2：middleware 体系成型

- adapter middleware API
- Huma middleware API
- group middleware / modifier / transformer API
- endpoint 级默认能力

### P3：稳定与裁剪

- 合并重复入口
- 删除临时 helper
- 固化最小长期 API 面

---

## 13. 成功标准

若本轮迭代成功，`httpx` 应达到以下状态：

1. 用户能用统一的 `httpx` API 组织 HTTP 服务能力
2. 用户能直接配置和控制 OpenAPI 与 docs
3. 用户能在接口层选择文档 UI
4. 用户能同时使用 adapter middleware 与 Huma middleware
5. 用户在高级场景下可以直接访问 Huma API
6. `httpx` 不会因为封装不足而限制 Huma 的高级能力
7. `httpx` 仍保持轻量，而不是演进成重型 framework

---

## 14. 总结

本次路线重新定义后，`httpx` 的目标不是“少量 helper + 透传”，而是：

> 提供一套统一、正式、完整的 API 来组织 HTTP 服务能力，并原生集成 OpenAPI、文档与 Huma 高级能力，同时保留 Huma 原生访问出口。

因此后续迭代应围绕三条主线推进：

1. 建立完整的 `httpx` 正式 API 面
2. 完整暴露 Huma 的核心能力域
3. 明确 adapter middleware 与 Huma middleware 的双层模型

---

## 15. 当前实现状态（2026-03-07）

以下状态用于记录 roadmap 对应能力在当前代码中的落地情况，便于后续继续迭代时直接对照。

### 15.1 已完成

#### Server / OpenAPI / docs

- `Server.HumaAPI()`
- `Server.OpenAPI()`
- `Server.ConfigureOpenAPI(...)`
- `Server.PatchOpenAPI(...)`
- `WithOpenAPIInfo(...)`
- `WithOpenAPIDocs(...)`
- `WithDocs(...)`
- `Server.Docs()`
- `Server.ConfigureDocs(...)`
- docs/OpenAPI/schema 路由已支持 adapter 层重绑定
- 已支持内置 docs renderer:
  - Stoplight Elements
  - Scalar
  - Swagger UI

#### Security / components / global parameters

- `WithSecurity(...)`
- `RegisterSecurityScheme(...)`
- `SetDefaultSecurity(...)`
- `RegisterComponentParameter(...)`
- `RegisterComponentHeader(...)`
- `RegisterGlobalParameter(...)`
- `RegisterGlobalHeader(...)`
- `AddTag(...)`

#### Group 能力

- `Group.HumaGroup()`
- `Group.UseHumaMiddleware(...)`
- `Group.UseOperationModifier(...)`
- `Group.UseSimpleOperationModifier(...)`
- `Group.UseResponseTransformer(...)`
- `Group.DefaultTags(...)`
- `Group.DefaultSecurity(...)`
- `Group.DefaultParameters(...)`
- `Group.DefaultSummaryPrefix(...)`
- `Group.DefaultDescription(...)`
- `Group.RegisterTags(...)`
- `Group.DefaultExternalDocs(...)`
- `Group.DefaultExtensions(...)`

#### 请求层行为

- `WithPanicRecover(...)` 已接到 typed handler 包装层
- `WithAccessLog(...)` 已接到 `httpx.Server` 请求入口层
- 请求 access log 已支持记录：
  - method
  - path
  - status
  - duration
  - route pattern
  - handler name

#### 文档 / 示例

- `README.md` / `README_ZH.md` 已与当前正式 API 面基本对齐
- 已补充并更新 examples:
  - `quickstart`
  - `auth`
  - `organization`
  - `config`
- `auth` / `organization` 已有中英文 README

---

### 15.2 已明确的设计收敛

#### timeout / max header bytes

这类配置不再视为 `httpx` 顶层统一配置。

原因：

- 它们属于 adapter / 底层 server library 的构造期能力
- 不同 adapter 的承载方式不同
- 放在 `httpx/options.ServerOptions` 会制造“能配但不生效”的假 API

当前结论：

- `httpx/options.ServerOptions` 不再承载这些字段
- timeout 设计应下沉到各 adapter 的构造期 `Options`
- 不做 `httpx` 层的运行期热修改 API

#### panic recover

保留在 `httpx` typed handler 层，而不是放到 adapter-native middleware 层统一处理。

当前结论：

- `httpx` 只保证自己注册的 typed handler 不因 panic 直接打崩
- adapter-native route 仍由各框架自己的 recover 机制负责

#### access log

保留在 `httpx.Server` 请求入口层，而不是下沉到每个 adapter 各自实现一套。

当前结论：

- `httpx` 统一记录 typed operation 请求
- 不承诺完全覆盖 adapter-native middleware 的日志体系

#### logger

`httpx.WithLogger(...)` 与 adapter logger 需要明确区分，但桥接层应尽量保持一致。

当前结论：

- `httpx.Server.logger` 属于 `httpx` 层
- adapter 自己的 logger 属于 adapter 构造期配置
- 当 adapter 支持时，`httpx.WithLogger(...)` 可以将 logger 下推到 adapter bridge 层
- 不自动承诺完全接管 gin / echo / fiber / chi 原生日志体系

---

### 15.3 当前发现并已处理/正在处理的问题

#### 已确认的历史问题

- `ReadTimeout` / `WriteTimeout` / `IdleTimeout` / `MaxHeaderBytes` 曾经只是死配置，已从 `httpx/options.ServerOptions` 移除
- `EnablePanicRecover` 之前未真正接线，现已生效
- `EnableAccessLog` 之前未真正接线，现已生效
- `WithPrintRoutes(...)` 之前未真正触发，现已接到路由注册路径
- `ContextOptions.CancelOnPanic` 之前没有实际消费逻辑，已从 options 语义中移除

#### 正在继续收敛的问题

- adapter 构造期 `Options` 正在补齐：
  - logger
  - timeout / shutdown timeout
  - max header bytes（适用于 net/http-based adapter）
- `httpx.WithLogger(...)` 对外部传入 adapter 的自动下推语义正在统一
- `echo` / `gin` / `fiber` / `std` 的 adapter logger 语义仍需继续文档化，明确哪些日志属于：
  - `httpx` 层
  - adapter bridge 层
  - 框架原生日志层

---

### 15.4 还未完全完成的 roadmap 项

#### Adapter middleware 正式 API

roadmap 中已经明确需要：

- `UseAdapterMiddleware(...)`
- 更清晰的 adapter-native middleware 正式入口

当前状态：

- 仍以 adapter escape hatch 为主
- 还没有统一的 `httpx` 层 adapter middleware 正式 API

#### Group / endpoint 更细的默认能力

当前已经完成了 tags / security / parameters / summary / description / external docs / extensions，  
但仍可继续补充：

- group-level 默认响应 metadata
- endpoint 级默认能力收口
- 更多 operation helper 的正式 API

#### adapter 构造期配置文档

timeout/logger 已经完成设计收敛，但文档层还需要继续补充：

- `std` adapter 的 server config
- `gin` adapter 的 server config
- `echo` adapter 的 server config
- `fiber` adapter 的 app config

---

### 15.5 当前建议的下一步

后续迭代建议按以下顺序继续：

1. 完成各 adapter 的构造期 `Options` 收口，并补测试
2. 明确 adapter logger 与框架原生日志的边界，并更新 README / examples
3. 继续补齐 adapter middleware 的正式 API
4. 持续清理剩余占位注释与旧文档表述

只要这三点成立，`httpx` 就能成为一个真正可持续演进的统一 HTTP 组织层，而不是一个能力不完整的封装层。
