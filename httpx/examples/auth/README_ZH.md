# httpx 认证示例

这个示例演示如何使用当前 `httpx` API 配置认证相关的 OpenAPI 能力。

它展示了：

- 使用 `WithSecurity(...)` 注册 security schemes
- 使用 `RegisterGlobalHeader(...)` 注册请求追踪 header
- 在 group 上设置默认认证与标签
- 通过 typed input 绑定 `Authorization`、`X-API-Key`、`X-Request-Id`

## 运行

在仓库根目录执行：

```bash
go run ./httpx/examples/auth
```

## 路由

- `GET /api/health`
- `GET /api/secure/profile`

## 文档

- Docs UI：`http://localhost:8080/docs`
- OpenAPI JSON：`http://localhost:8080/openapi.json`

## 试用

Bearer 示例：

```bash
curl http://localhost:8080/api/secure/profile ^
  -H "Authorization: Bearer demo-token" ^
  -H "X-Request-Id: req-1"
```

API key 示例：

```bash
curl http://localhost:8080/api/secure/profile ^
  -H "X-API-Key: demo-key" ^
  -H "X-Request-Id: req-2"
```

## 关注点

`GET /api/secure/profile` 这个 operation 会被文档化为：

- 可选认证：`BearerAuth` 和 `ApiKeyAuth`
- 标签：`auth`
- 全局请求头：`X-Request-Id`
