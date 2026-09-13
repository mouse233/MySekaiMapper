<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 快速上手

请选择适合你环境的通知方式：

- **路径 A — 仅使用 Telegram**：最简单的选项；不需要玩家路由文件或公开图片服务器。
- **路径 B — 启用 Bark**：配置 Bark 密钥、玩家路由以及用于图片的公开静态文件服务器。

### 1. 前置条件与构建

需要 Go **1.25 或更高版本**。

```bash
go version
cp .env.example .env
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

`.env` 中的 `AES_KEY` 和 `AES_IV` 必须是 16 字节的 AES-128-CBC 密钥和 IV。请勿提交 `.env` 或本地路由文件。

### 2. 配置 `.env`

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `AES_KEY`, `AES_IV` | 是 | 16 字节的 MySekai AES-128-CBC 密钥和 IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | 仅 Telegram | 从 [@BotFather](https://t.me/BotFather) 获取的 Bot 凭据和目标聊天 ID |
| `BARK_ICON` | 可选 | Bark 通知中包含的图标 URL |
| `BARK_IMAGE_BASE` | Bark 图片 | 已归档地图图片的公开基础 URL |
| `FALLBACK_IMAGE_BASE` | 可选 | 未设置 `BARK_IMAGE_BASE` 时使用的图片基础 URL |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | 可选 | Reqable 上报端点设置 |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | 可选 | 覆盖仓库默认目录 |

### 3. 路径 A — 仅使用 Telegram

1. 在 `.env` 中设置 Telegram 变量：

   ```dotenv
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 可选：使用已有加密存档验证解析和通知：

   ```bash
   bin/mysekaimapper generate --input data/raw_mysekai/mysekai.bin
   bin/mysekaimapper notify \
     --output data/latest \
     --task-id manual-001 \
     --player-id 1234567890123456789
   ```

3. 启动服务进行正常运行：

   ```bash
   bin/mysekaimapper serve --host 0.0.0.0 --port 9478
   ```

`config/push_map.json` 中没有配置的玩家默认使用 Telegram。路径 A 不需要 Bark 路由文件、推送路由文件或公开图片服务器。

### 4. 路径 B — 启用 Bark

在路径 A 的配置基础上（仅使用 Bark 的路由可以省略 Telegram）：

1. 从 `config/bark_map.example.json` 创建 `config/bark_map.json`，将 Bark 别名映射到各设备密钥。
2. 从 `config/push_map.example.json` 创建 `config/push_map.json`，将玩家 ID 映射到 Bark 别名、`telegram`、`none` 或它们的组合：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"],
     "1234567890123456791": "none"
   }
   ```

3. 使用公开 HTTP(S) 静态文件服务器提供仓库的 `data/` 目录，并将公开根路径设置为 `BARK_IMAGE_BASE`：

   ```dotenv
   BARK_IMAGE_BASE=https://maps.example.com
   ```

未配置的玩家默认使用 Telegram。因此，如果没有配置 Telegram，未配置的玩家不会收到通知；仅使用 Bark 时，请为玩家显式分配 Bark 别名。
