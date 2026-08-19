# Go 重构说明

生产实现已从 Python 迁移至 Go，以降低启动与资源消耗、限制后台渲染并发，并提供单一可执行文件部署。

## 变化

- 仓库根目录现在就是 Go module：`cmd/`、`internal/`、`go.mod`、`go.sum`。
- 服务已包含 AES/MsgPack 解析、HAR 接入、渲染、存储以及 Telegram/Bark 通知。
- `notify` 可作为独立 Go 子命令使用。
- 通知会枚举全部常规 `site_*.png` 输出。
- Go CI 会在 push 与 PR 时测试并构建服务。

## 兼容性

HTTP 端点、`.env` 变量、输出名称、归档布局和路由配置保持兼容。Go 使用固定画布绘图，不保证与 Matplotlib 输出像素级一致。手动通知采用显式 flags，玩家专属路由需传入 `--player-id`。

Python 参考实现已归档在 `legacy/python` 分支与 `python-v0.2.0` 标签中，不再属于活跃运行时。
