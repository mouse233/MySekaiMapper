# 快速上手

## 前置条件

- Go 1.25+
- 从游戏客户端取得的 AES-128-CBC 密钥与 IV
- 可选的 Telegram 或 Bark 凭据

## 配置

```bash
cp .env.example .env
```

填写 16 字节的 `AES_KEY`、`AES_IV`。`.env`、`config/bark_map.json` 和 `config/push_map.json` 均为私密配置，不应提交。

## 构建并启动

```bash
mkdir -p bin
go test ./...
go build -o bin/mysekaimapper ./cmd/mysekaimapper
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

服务提供 `POST /uploadMySekai`，以及默认开启的 `POST /reqable/report`。任务在受限的后台 worker 中渲染，成品发布到 `data/archive/` 和 `data/latest/`。

## 手动生成

```bash
bin/mysekaimapper generate --input data/raw_mysekai/save.bin --output data/latest
```
