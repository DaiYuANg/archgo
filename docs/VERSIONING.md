# ArcGo 版本文档管理

本文档说明如何管理 ArcGo 的多版本文档。

## 📁 目录结构

```text
docs/
├── content/
│   ├── docs/                      # 当前版本文档
│   └── versioned/                 # 历史版本文档
│       ├── v0.2.1/
│       ├── v0.2.0/
│       └── ...
├── data/
│   └── versions.yaml              # 版本配置
├── layouts/
│   ├── _partials/                 # Hextra v0.12+ 覆盖模板
│   │   ├── navbar.html
│   │   └── navbar/version-switcher.html
└── scripts/
    ├── sync-versions.go            # 入口
    ├── git.go                      # git 元信息读取
    ├── versions.go                 # 版本模型与 semver 比较
    ├── versions_file.go            # versions.yaml 写入
    └── filesystem.go               # 目录/文件复制
```

## ✅ 运行前提（Hextra）

- 主题版本：`github.com/imfing/hextra v0.12.1`
- Hugo 最低版本：`0.146.0`（extended）
- 推荐命令：`go tool hugo`（避免全局 `hugo` 版本不一致）

## 🚀 快速开始

### 发布新版本后同步

```bash
cd docs
go run ./scripts
```

脚本会自动：
1. 读取 git tags
2. 更新 `data/versions.yaml`
3. 初始化缺失的版本目录

## 🧩 versions.yaml 规范

```yaml
versions:
  - name: "v0.3.0"
    release: "v0.3.0"
    path: "/docs"
    current: true

  - name: "v0.2.2"
    release: "v0.2.2"
    path: "/versioned/v0.2.2/docs"
    current: false
```

要点：
- 当前版本路径固定用 `/docs`
- 历史版本路径用 `/versioned/<tag>/docs`
- 列表顺序建议新到旧

## 🔧 本地预览

```bash
cd docs
go tool hugo server --buildDrafts --disableFastRender
```

访问：
- 当前文档入口：`http://localhost:1313/docs/`
- 历史版本入口：`http://localhost:1313/versioned/v0.2.2/docs/`

## 🐛 故障排查

### 多版本切换器不显示

1. 检查 `docs/data/versions.yaml` 是否存在且合法
2. 检查 `docs/layouts/_partials/navbar.html` 中是否有 `type: versions` 分支
3. 检查 `docs/layouts/_partials/navbar/version-switcher.html` 是否存在
4. 执行 `go tool hugo version`，确认 Hugo >= `0.146.0`（extended）

### 切换后跳到错误版本或 404

1. 检查 `versions.yaml` 的 `path` 是否为 `/docs` 和 `/versioned/<tag>/docs`
2. 检查 `docs/content/versioned/` 下是否存在误嵌套目录（例如 `v0.2.1/v0.2.2`）
3. 清理缓存后重启：

```bash
cd docs
go tool hugo --gc
go tool hugo server --buildDrafts --disableFastRender
```
