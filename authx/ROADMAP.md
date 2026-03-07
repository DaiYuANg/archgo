# AuthX Roadmap

## 1. 项目定位

**AuthX** 的定位不是一个完整的安全 framework，而是一个：

> 基于 Authboss + Casbin 的、面向业务 API 的 Go 安全内核，并提供对第三方 HTTP/RPC 生态的适配能力。

项目目标止步于第二层，即：

- 提供稳定的 `authx-core`
- 提供对常见 Go Web / RPC 生态的 `integration adapters`
- 不接管整个运行时，不做全家桶 framework

这意味着 AuthX 关注的是：

- 统一认证抽象
- 统一授权抽象
- 统一安全上下文模型
- 策略装载与热更新
- 对第三方 HTTP 框架的轻量适配

而 **不** 关注：

- 自建完整 Web Framework
- 接管路由系统
- 接管完整 Session Server 生命周期
- 接管应用整体异常处理体系
- 提供类似 Spring Security 的全栈运行时控制面

---

## 2. 分层设计

AuthX 目标分为两层：

### 2.1 第一层：`authx-core`

核心安全内核，负责安全领域模型与执行逻辑。

包含：

- Identity / Authentication / SecurityContext
- Credential / Authenticator
- Authorization Request / Authorizer
- PolicySnapshot / PolicySource
- IdentityProvider
- SubjectResolver
- PolicyMerger
- Error Model
- Event Contract
- Diagnostics Interface

### 2.2 第二层：`authx-integrations`

适配层，负责与第三方 HTTP / RPC / token/session 生态集成。

建议拆分为多个子模块，例如：

- `authx-http`
- `authx-chi`
- `authx-gin`
- `authx-huma`
- `authx-grpc`
- `authx-apikey`
- `authx-bearer`

适配层职责：

- 从请求中提取 credential
- 注入 / 读取 security context
- 将认证授权错误映射到具体协议响应
- 提供轻量级 middleware / interceptor
- 提供与第三方路由元数据的集成 helper

---

## 3. 设计原则

### 3.1 Core First
先把核心语义做稳，再做集成层。  
所有 adapter 必须依赖 `authx-core`，而不是反向污染 core。

### 3.2 Integration Friendly
以“适配第三方生态”为目标，而不是替代第三方生态。

### 3.3 Explicit over Magic
避免隐式注册、隐式扫描、隐式路由推断。  
Go 生态中，安全能力应尽量显式组装。

### 3.4 Stable Domain Model
优先保证安全领域模型稳定，而不是快速堆功能。

### 3.5 Thin Adapter
适配层尽量薄，只做：
- 提取
- 转换
- 注入
- 响应映射

不要把业务逻辑塞进 adapter。

### 3.6 Policy-Driven Authorization
授权优先采用策略驱动模型，默认以 RBAC / ACL 为主，ABAC 只作为后续扩展方向预留。

---

## 4. 当前阶段判断

**当前状态：Phase 1 已完成，Phase 2 进行中（约 50%）**

### 4.1 Phase 1 完成情况（✅ 100%）

Phase 1「核心内核稳定化」已全部完成，交付了以下能力：

**核心模型与扩展点：**
- ✅ 稳定的 `Identity` / `Authentication` / `SecurityContext` / `Request` 模型
- ✅ `SubjectResolver` 接口及多种实现（Default/Tenant/Prefix/Mapped/Composed）
- ✅ `PolicyMerger` 接口及多种实现（Default/Priority/ConflictDetecting/Strict/DenyOverrides）
- ✅ `CasbinAuthorizer` 支持多种匹配模式（exact/prefix/glob/keyMatch/keyMatch2）

**错误与事件模型：**
- ✅ 完整的 typed error 系统（`ErrorCode` + `*Error`）
- ✅ 支持 `errors.Is` / `errors.As` / `errors.Unwrap`
- ✅ 辅助函数：`IsUnauthorized()` / `IsForbidden()` / `IsNotFound()`
- ✅ **完全复用 `eventx` 的事件系统**（支持异步/并行/订阅/中间件）
- ✅ 日志事件处理器

**诊断能力：**
- ✅ `Diagnostics` 结构体暴露运行时信息
- ✅ `DiagnosticsTracker` 记录 reload 历史

**测试覆盖：**
- ✅ 核心组件单元测试完备（50+ 测试用例）
- ✅ 集成测试验证端到端流程

### 4.2 Phase 2 完成情况（🟡 约 50%）

Phase 2「基础适配层建设」已启动，当前完成：

**已完成：**
- ✅ `MemoryPolicySource` - 内存策略源（支持动态更新）
- ✅ `StaticPolicySource` - 只读静态策略源
- ✅ `MutablePolicySource` - 支持 mutator 函数的可变策略源
- ✅ `FilePolicySource` - 文件策略源（支持 fsnotify 热加载）
- ✅ `PolicySourceChain` - 多策略源合并链
- ✅ `FallbackPolicySource` - 主备 fallback 机制
- ✅ `ConditionalPolicySource` - 条件选择策略源
- ✅ `CachedPolicySource` - 缓存包装器
- ✅ **`EventPublisher` 支持外部自定义实例注入**

**待完成：**
- ⏸️ `authx-http` 中间件层（从 HTTP 请求提取 credential、注入 SecurityContext）
- ⏸️ 认证方式扩展（API Key / Bearer Token verify-only）
- ⏸️ 数据库策略源（Database Policy Source）
- ⏸️ 远程 HTTP 策略源（Remote HTTP Policy Source）
- ⏸️ 示例项目（examples/basic、examples/http-chi 等）

### 4.3 当前主要短板

- ⏸️ 集成层尚未开始系统化建设（authx-http 待实现）
- ⏸️ 缺少生产级认证方式（仅有 password，缺少 API Key / Bearer）
- ⏸️ 缺少完整的示例和文档
- ⏸️ 可观测性能力待增强（metrics / tracing / audit）

---

## 5. 路线图总览

Roadmap 分四个阶段：

- Phase 0：定位与语义收敛
- Phase 1：核心内核稳定化
- Phase 2：基础适配层建设
- Phase 3：生产可用与生态扩展

---

## 6. Phase 0：定位与语义收敛

### 6.1 目标

将项目从“概念可运行”推进到“定义清晰、边界明确”。

### 6.2 核心任务

#### A. 明确项目定位
补充 README 与设计文档，明确说明：

- AuthX 是 security core + integrations
- 不是 full-stack security framework
- 当前支持什么
- 当前不支持什么

#### B. 收敛术语
重点处理以下命名问题：

- `UserDetails.Principal`
- `Identity.Principal()`
- `Permissions`
- `Request.Attributes`

建议原则：

- 登录标识使用 `Login` / `Username` / `SubjectKey`
- 认证后的业务负载使用 `Claims` / `Payload`
- 授权输入使用 `Subject / Resource / Action`

#### C. 收敛默认授权语义
明确当前默认模型为：

- RBAC / ACL-first
- 默认仅承诺 `subject + resource + action`
- attributes 暂不默认参与授权判定

#### D. 明确错误分类
至少区分：

- 认证失败
- 授权拒绝
- provider/source 内部错误
- 配置错误
- 可 fallback 的错误
- 不可 fallback 的错误

### 6.3 交付物

- README 重构
- `docs/architecture.md`
- `docs/concepts.md`
- `docs/non-goals.md`

### 6.4 完成标准

- 外部读者可以在 5 分钟内理解 AuthX 是什么、不是什
- 核心命名不再出现明显歧义
- 默认授权语义有明确文档约束

---

## 7. Phase 1：核心内核稳定化

### 7.1 目标

将 `authx-core` 做成一个真正可扩展、可长期演进的安全内核库。

---

### 7.2 核心任务一：稳定核心模型

#### A. 重构 `Identity`
明确 Identity 的职责：

- 表示认证成功后的身份视图
- 为授权与上下文传播提供标准结构
- 不承载过多与底层 provider 强绑定的信息

建议收敛字段：

- `ID`
- `Type`
- `Name`
- `Claims/Payload`
- `Roles`（若保留，需明确只是辅助信息还是授权输入）
- `Attributes`（若保留，需明确默认不参与授权）

#### B. 稳定 `Authentication`
明确 Authentication 代表：

- 某次认证结果
- 包含认证时间
- 包含策略版本或认证相关元数据

#### C. 稳定 `SecurityContext`
将 `SecurityContext` 作为运行时唯一可信上下文对象。  
避免仅通过 `Identity` 反推完整认证态。

#### D. 稳定 `Request`
明确授权请求模型：

- `Subject`
- `Resource`
- `Action`
- 可选 `Attributes`

并明确默认实现是否允许 `Subject` 自动从 `Identity` 解析。

---

### 7.3 核心任务二：新增关键扩展点

#### A. `SubjectResolver`
用于将 `Identity` 解析为授权使用的 subject。

建议接口：

```go
type SubjectResolver interface {
    ResolveSubject(ctx context.Context, identity Identity) (string, error)
}
```
用途：

多租户 subject

服务账号映射

外部 ID 转内部 ID

特殊 subject 命名策略

B. PolicyMerger

用于统一合并多个策略源的快照。

建议接口：

type PolicyMerger interface {
    Merge(ctx context.Context, snapshots ...PolicySnapshot) (PolicySnapshot, error)
}

用途：

去重

来源优先级

冲突检测

合并策略可配置

C. AuthorizationModel

允许替换默认授权模型，而不仅仅替换策略数据。

D. RequestMapper

允许将更高层的业务请求映射为授权请求。

7.4 核心任务三：增强授权模型
A. 支持 wildcard / path match

至少支持一种资源匹配策略：

prefix match

glob match

Casbin keyMatch/keyMatch2

B. 支持自定义 Casbin 函数注册

用于未来扩展：

own-resource

tenant-scope

dynamic match

domain-aware authorization

C. 明确角色与权限语义

二选一，必须明确：

角色/权限只作为展示信息，实际授权完全依赖 policy snapshot

角色/权限参与授权决策，并进入统一 policy compile 流程

7.5 核心任务四：错误模型与事件模型
A. 错误模型

建议引入 typed error / code：

invalid_credential

principal_not_found

bad_password

unauthenticated

forbidden

provider_unavailable

policy_invalid

policy_merge_conflict

B. 事件模型

定义基础事件契约：

authentication success

authentication failed

authorization allowed

authorization denied

policy loaded

policy replaced

provider fallback triggered

7.6 核心任务五：诊断能力

建议新增 diagnostics API，至少暴露：

当前 authorizer 类型

当前 authenticator 类型

当前 policy version

当前策略来源

最近一次 reload 时间

最近一次 reload 结果

provider/source 状态摘要

7.7 交付物

authx-core 稳定 public API

核心扩展点接口

错误模型

事件模型

诊断接口

更完善的单元测试与集成测试

### 7.8 完成标准

**状态：✅ 已完成**

达到以下标准时，Phase 1 结束：

- ✅ 核心模型清晰稳定
- ✅ 默认授权语义明确
- ✅ 支持 subject resolver（多种实现）
- ✅ 支持 policy merger（多种实现）
- ✅ 支持至少一种 wildcard resource match（prefix/glob/keyMatch/keyMatch2）
- ✅ 完整的错误模型和事件模型
- ✅ 诊断接口实现
- ✅ 对外可以称为"可扩展安全内核库"

**交付物清单：**

| 模块 | 文件 | 状态 |
|------|------|------|
| 核心模型 | identity.go, authentication.go, security_context.go, request.go | ✅ |
| SubjectResolver | subject_resolver.go | ✅ |
| PolicyMerger | policy_merger.go | ✅ |
| 错误模型 | errors.go | ✅ |
| 事件模型 | events.go | ✅ |
| 诊断能力 | diagnostics.go | ✅ |
| Casbin Authorizer | casbin_authorizer.go | ✅ |
| 策略源接口 | policy_source.go | ✅ |

8. Phase 2：基础适配层建设
8.1 目标

在不做 framework 的前提下，让 AuthX 能轻量接入主流 Go 服务生态。

8.2 核心任务一：authx-http

先做协议无关但面向 HTTP 的基础适配层。

职责：

Credential Extractor

SecurityContext Injector

Unauthorized/Forbidden Translator

Middleware Helper

Request → Authorization Request 映射

建议提供：

Middleware

OptionalMiddleware

Require(action, resource)

CurrentIdentity

CurrentSecurityContext

8.3 核心任务二：认证方式扩展

建议顺序如下。

第一批

Password

API Key

Anonymous

第二批

Bearer Token（仅 verify，不负责签发）

Static Token / Service Token

暂缓

Refresh Token

Remember Me

Full Session Lifecycle

MFA

原因：

verify-only 集成简单

issue & lifecycle 会显著放大系统复杂度

8.4 核心任务三：框架适配器

按优先级建议：

A. authx-chi

中间件接入简单

Go 生态接受度高

很适合作为第一个适配器

B. authx-huma

更符合你当前的技术关注点

适合 API-first 项目

可以把 AuthX 的授权抽象很好地嵌进去

C. authx-gin

用户面大

适合扩大使用面

D. authx-grpc

提供 interceptor 支持

适合服务间调用安全场景

### 8.5 核心任务四：策略源实现

**状态：✅ 已完成**

建议提供几个标准实现：

- ✅ Memory Policy Source - `policy_source_memory.go`
- ✅ File Policy Source - `policy_source_file.go`（支持 fsnotify 热加载）
- ✅ Policy Source Chain - `policy_source_chain.go`
- ✅ Fallback Policy Source - `policy_source_chain.go`
- ✅ Conditional Policy Source - `policy_source_chain.go`
- ✅ Cached Policy Source - `policy_source_chain.go`

待完成：

- ⏸️ Database Policy Source（支持 MySQL/PostgreSQL/SQLite）
- ⏸️ Remote HTTP Policy Source（从远程 HTTP 端点加载）
- ⏸️ Watchable Policy Source（支持 etcd/consul 等配置中心）

**交付物清单：**

| 策略源类型 | 文件 | 状态 |
|------------|------|------|
| MemoryPolicySource | policy_source_memory.go | ✅ |
| StaticPolicySource | policy_source_memory.go | ✅ |
| MutablePolicySource | policy_source_memory.go | ✅ |
| FilePolicySource | policy_source_file.go | ✅ |
| PolicySourceChain | policy_source_chain.go | ✅ |
| FallbackPolicySource | policy_source_chain.go | ✅ |
| ConditionalPolicySource | policy_source_chain.go | ✅ |
| CachedPolicySource | policy_source_chain.go | ✅ |
| DatabasePolicySource | 待实现 | ⏸️ |
| RemoteHTTPPolicySource | 待实现 | ⏸️ |

8.6 核心任务五：开发者体验

需要补足：

示例项目

最小接入示例

多适配器示例

常见使用模式示例

建议至少提供：

examples/basic

examples/http-chi

examples/http-huma

examples/apikey

examples/policy-reload

8.7 交付物

authx-http

authx-chi

authx-huma

authx-apikey

authx-bearer（verify-only）

file/db policy source

example projects

8.8 完成标准

达到以下条件时，Phase 2 完成：

可在至少 2 种主流 Go HTTP 生态中无侵入接入

可通过 middleware / interceptor 注入安全上下文

支持 password / apikey / bearer verify-only

策略可来自 file / memory / db

具备清晰 example 与文档

9. Phase 3：生产可用增强
9.1 目标

让 AuthX 成为可在中大型业务项目中稳定使用的安全基础组件。

9.2 核心任务
A. 可观测性增强

metrics

tracing hooks

structured audit events

B. 配置能力增强

builder options

diagnostics endpoint helper

health summary helper

C. 缓存与性能优化

policy compile cache

subject resolve cache

source refresh optimization

D. 多租户能力预留

即便暂不全面支持，也建议在模型上预留：

tenant-aware subject

tenant-aware policy namespace

tenant-aware request attributes

E. 更强测试矩阵

race test

benchmark

adapter integration tests

invalid policy tests

reload consistency tests

9.3 完成标准

可支撑生产落地

具备基础审计与诊断能力

在典型并发场景下可稳定运行

文档和示例足以支持他人接入

10. 建议的模块结构
authx/
  core/
    identity.go
    authentication.go
    security_context.go
    credential.go
    request.go
    errors.go
    events.go
    diagnostics.go

  authn/
    authenticator.go
    password/
    apikey/
    bearer/

  authz/
    authorizer.go
    casbin/
    subject_resolver.go
    request_mapper.go

  policy/
    snapshot.go
    source.go
    merger.go
    loader.go

  integration/
    http/
    chi/
    huma/
    gin/
    grpc/

  examples/
  docs/

说明：

core 放领域模型

authn 放认证实现

authz 放授权实现

policy 放策略体系

integration 放适配器

不要把所有实现全塞进一个包里，否则后期会越来越难维护。

## 11. 版本规划建议

**当前版本：v0.3**

### v0.1 ✅ 已完成

目标：最小可运行版本

- ✅ password auth
- ✅ casbin authz
- ✅ security context
- ✅ basic policy reload

### v0.2 ✅ 已完成

目标：语义收敛版

- ✅ README 重构
- ✅ 核心命名清理
- ✅ typed errors
- ✅ subject resolver
- ✅ policy merger
- ✅ diagnostics contract

### v0.3 ✅ 已完成（当前版本）

目标：内核稳定版

- ✅ wildcard resource match（prefix/glob/keyMatch/keyMatch2）
- ✅ authorization model extension
- ✅ event contract
- ✅ richer tests
- ✅ cleaner package layout
- ✅ 完整的策略源实现（Memory/File/Chain/Fallback/Conditional/Cached）
- ✅ **完全复用 `eventx` 的事件系统**（异步/并行/订阅/中间件）
- ✅ **`EventPublisher` 支持外部自定义实例注入**（依赖注入模式）

### v0.4 🟡 进行中（当前重点）

目标：基础适配版

- ⏸️ authx-http（中间件层）- **下一步重点**
- ⏸️ authx-chi
- ⏸️ authx-huma
- ⏸️ apikey auth
- ⏸️ bearer verify-only

### v0.5 ⏸️ 计划中

目标：可落地版

- ⏸️ file/db policy source
- ⏸️ better examples
- ⏸️ diagnostics helper
- ⏸️ audit hooks
- ⏸️ performance improvements

### v0.6 ⏸️ 计划中

目标：生产强化版

- ⏸️ observability
- ⏸️ benchmark
- ⏸️ caching
- ⏸️ adapter hardening
- ⏸️ better non-happy-path support

## 12. 优先级建议

**更新：2026 年 3 月**

如果当前只做最有价值的部分，建议优先级如下：

### P0 ✅ 已完成

- ✅ 项目定位文档
- ✅ 核心命名清理
- ✅ 错误模型
- ✅ SubjectResolver
- ✅ PolicyMerger

### P1 ✅ 已完成

- ✅ wildcard/path match
- ✅ diagnostics
- ✅ event contract
- ✅ package restructuring
- ✅ 完整的策略源实现
- ✅ **事件系统完全复用 `eventx`**
- ✅ **EventPublisher 支持外部实例注入**

### P2 🟡 当前优先级

- ⏸️ authx-http（中间件层）- **当前重点**
- ⏸️ authx-chi
- ⏸️ authx-huma
- ⏸️ apikey / bearer verify-only

### P3 ⏸️ 计划中

- ⏸️ database policy source
- ⏸️ examples
- ⏸️ metrics / audit hooks

13. 非目标范围

以下内容不建议在当前路线中优先投入：

自建完整 Web Framework

注解式全自动 method security

完整 session server 托管

自建用户体系替代 Authboss

自建权限引擎替代 Casbin

一开始就做复杂 ABAC 平台

一开始就做多租户全栈模型

一开始就做 JWT 签发中心

这些能力都可能有价值，但不属于当前路线的最优先目标。

14. 最终结论

AuthX 最合理的发展路线是：

先把它做成一个稳定、清晰、可扩展的 security core library，再围绕主流 Go HTTP/RPC 生态补齐轻量适配层，而不是把它推向一个重型 framework。

这条路线的优点在于：

符合 Go 生态习惯

更容易控制复杂度

更容易稳定 public API

更适合逐步积累用户与使用场景

不会过早陷入 framework 级维护成本

AuthX 的成功关键不在于“功能做得多像 Spring Security”，而在于：

模型是否清晰

边界是否稳定

扩展点是否正确

适配层是否轻薄好用