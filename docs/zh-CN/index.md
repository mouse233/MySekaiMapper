# MySekaiMapper

MySekaiMapper 是一个 Go 服务：解密 MySekai 存档、绘制采集点地图、归档结果，并推送到 Telegram 或 Bark。

## 从这里开始

1. 安装 Go 1.25 或更高版本。
2. 将 `.env.example` 复制为 `.env`，填写 `AES_KEY` 和 `AES_IV`。
3. 构建并启动服务：

```bash
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

请阅读[快速上手](./guide/quickstart)、[命令行参考](./guide/cli)与 [Go 重构说明](./guide/refactor-go)。
