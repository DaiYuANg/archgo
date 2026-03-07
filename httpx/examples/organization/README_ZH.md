# httpx Organization 示例

这个示例把较新的 `httpx` 服务组织 API 串成了一个完整可运行的服务。

它展示了：

- 自定义 docs 路径与 docs renderer
- OpenAPI security 注册
- 全局请求 header
- group 级 tags、security、parameters、summary/description 默认值
- group 级 external docs 与 OpenAPI extensions

## 运行

在仓库根目录执行：

```bash
go run ./httpx/examples/organization
```

## 路由

- `GET /api/health`
- `GET /api/admin/tenants/{id}`

## 文档

- Docs UI：`http://localhost:8080/reference`
- OpenAPI JSON：`http://localhost:8080/spec.json`
- OpenAPI YAML：`http://localhost:8080/spec.yaml`

## 关注点

`/api/admin/tenants/{id}` 这个 operation 通过 group 注册，并继承了：

- tags：`admin`、`tenants`
- security：`bearerAuth`
- header parameter：`X-Tenant`
- summary 前缀：`Admin`
- 共享 description
- external docs 元数据
- extension：`x-owner=platform`

服务还注册了一个全局请求头：

- `X-Request-Id`
