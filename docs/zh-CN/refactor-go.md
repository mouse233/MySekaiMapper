<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# Go 重构说明

当前运行时仅使用 Go。模块采用包含 `cmd/`、`internal/`、`go.mod` 和 `go.sum` 的标准根目录结构；Python 源码、依赖和 CI 已被移除。归档的参考实现仍保留在 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 分支和 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 标签中。

HTTP 端点、环境变量、输出名称、归档布局和路由文件格式保持兼容。Go 渲染器使用固定画布，因此生成的 PNG 不保证与此前的 Matplotlib 输出逐像素完全一致。
