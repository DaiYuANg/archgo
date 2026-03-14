# ArcGo 文档版本管理

当前阶段，ArcGo 文档采用**单版本**策略：只维护 `/docs` 主线内容。

## 当前约定

- 当前版本号：`v0.3.0`
- 站点文档入口：`/docs`
- `docs/data/versions.yaml` 只保留一个 `current: true` 条目
- 未发布历史版本不保留在仓库中

## 目录结构（当前）

```text
docs/
├── content/
│   └── docs/                      # 主线文档
├── data/
│   └── versions.yaml              # 仅当前版本
├── layouts/
│   └── _partials/
│       └── navbar/version-switcher.html
└── scripts/
    ├── sync-versions.go
    ├── git.go
    ├── versions.go
    ├── versions_file.go
    └── filesystem.go
```

## versions.yaml 示例

```yaml
versions:
  - name: "v0.3.0"
    release: "v0.3.0"
    path: "/docs"
    current: true
```

## 本地预览

```bash
cd docs
go tool hugo server --buildDrafts --disableFastRender
```

入口：`http://localhost:1313/docs/`

## 未来发布策略（可选）

如果后续需要公开历史版本，再启用多版本：

1. 将当前 `/docs` 快照复制到 `content/versioned/<tag>/docs`
2. 在 `versions.yaml` 增加历史版本条目
3. 保持主线版本继续使用 `/docs`
