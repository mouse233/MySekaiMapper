# MySekaiMapper

🌐 **Languages**: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-CN/>

MySekaiMapper 是面向 *Project SEKAI* MySekai 存档的 Go 服务：将加密存档转换为采集点地图，并将结果发送到 Telegram 或 Bark（Day.app）。

它可配合 MitM 抓包客户端或 Reqable 的 **上报服务器（Report Server）** 使用：抓包工具上传 MySekai 存档，服务解密、解析并绘制地图和稀有资源摘要，归档产物后自动发送通知，无需手动处理。

常见区域会生成 `site_5.png`（草地）、`site_6.png`（海滩）、`site_7.png`（花园）、`site_8.png`（纪念地）和 `rare_resources.txt`。渲染器与通知器也支持额外的常规 `site_*.png` 文件。

抓包流程已经在朝夕光年运营的国服和台服验证；其他地区是否可用取决于 API 路径和存档格式。

## 工作流程

```text
游戏 API 响应 → MitM 模块 / Reqable 上报服务器
    │  ① POST /uploadMySekai（单次或有序分片上传）
    │  ② POST /reqable/report（HAR，可选 gzip / br / zstd）
    ▼
mysekaimapper serve
    ├─ AES-128-CBC 解密 + MsgPack 解析 + 坐标归一化
    ├─ 绘制 site_*.png + rare_resources.txt
    ├─ 归档到 data/archive/by-id/<player_id>/<timestamp>/
    └─ 发布 data/latest/ 并通知
         ├─ Telegram：以 multipart 媒体组上传本地图片
         └─ Bark：从公开静态文件服务器读取图片 URL
```

## 快速上手

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

## 运行服务

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

服务会输出就绪地址，并记录上传/上报接收、入队、解析、渲染、归档、通知、耗时、任务 ID 与 `player_id` 等生命周期日志。日志不会记录存档正文、密钥、令牌或完整通知 URL。

进程处理 `SIGINT` 和 `SIGTERM`：先停止接收 HTTP 请求，再最多等待 15 秒处理已接收任务。

如果二进制在仓库外运行，请传入 `--root /path/to/MySekaiMapper`；否则会从工作目录自动发现仓库根目录。

## 上传接口

`POST /uploadMySekai` 直接接收加密的 MySekai 响应正文。通常单次上传即可；为兼容抓包客户端，仍支持有序分片。

| 请求头 | 必填 | 含义 |
| --- | --- | --- |
| `X-Upload-Id` | 是 | 匹配 `^[A-Za-z0-9_-]{1,64}$` 的任务标识符 |
| `X-Chunk-Index` | 是 | 从零开始的分片序号 |
| `X-Total-Chunks` | 是 | 分片总数，范围为 1 到 10 |
| `X-Original-Url` | 否 | 游戏原始 URL；`/user/<id>` 用于提供玩家路由 |
| `X-Script-Version` | 否 | 为兼容抓包客户端而接受，服务会忽略 |

加密存档、每个分片和合并后的上传均限制为 1 MiB。成功接收后返回纯文本 `OK`；绘制和通知会在后台继续进行。

### 单次上传示例

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### 分片上传示例

使用相同的 `X-Upload-Id`、有序的索引，且最多十个分片：

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_
total=$(ls /tmp/ms_chunk_* | wc -l | tr -d ' ')

i=0
for chunk in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
    --data-binary @"$chunk"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

常见响应包括：成功接收时为 `200 OK`，标识符或分片范围无效时为 `400 Bad Request`，超过大小限制时为 `413 Payload Too Large`，缺少必填请求头或其值不是整数时为 `422 Unprocessable Entity`。

## Reqable 上报服务器

Reqable v2.20.0+ 可以将捕获的 HTTP 会话作为 HAR JSON POST 到本服务。上报端点默认启用，并与 `/uploadMySekai` 共存。

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | 设置为 `0`、`false`、`no` 或 `off` 可禁用上报 |
| `REPORT_PATH` | `/reqable/report` | 在 Reqable 中配置的端点路径 |
| `REPORT_MAX_SIZE` | `1` | 解压后的 HAR 请求体大小上限，单位为 MiB |
| `REPORT_TOKEN` | 空 | 可选；要求匹配 `X-Report-Token` |

### 处理流程

每次上报时，服务会：

1. 解压 `identity`、`gzip`、`br`、`zstd` 或 `zstandard` 内容并解析 HAR；也支持没有 content-size 字段的流式 zstd 帧。
2. 遍历 `log.entries`，接收第一个能够使用 `AES_KEY`/`AES_IV` 解密并验证为 MySekai 存档的响应体；响应体不匹配时回退检查请求体。
3. 从命中会话 URL 的 `/user/<id>` 中提取 `player_id`。
4. 将加密存档保存至 `data/raw_mysekai/`，并启动与上传接口相同的“绘制 → 归档 → 通知”流水线。

> Reqable 每个会话只上报一次，不会重试。请保持服务可用并关注 `[REPORT]` 日志。语法正确但不含 MySekai 存档的 HAR 仍会返回 `ok`；每次上报只处理第一份有效存档。

### Reqable 侧配置

- **匹配规则**：`https://<游戏 API 域名>/api/user/*/mysekai*`
- **服务器 URL**：`http://<你的服务器>:9478/reqable/report`（或自定义的 `REPORT_PATH`）

| 服务器 | 游戏 API 域名 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

该匹配规则已在国服验证。如果所在地区使用其他 MySekai API 路径，请检查实际抓包 URL 后调整规则。

### 安全

Reqable 无法附加自定义 `X-Report-Token` 请求头。请使用足够长的随机 `REPORT_PATH`，例如 `/reqable/report/<random>`，再通过反向代理或防火墙限制访问；不要在没有保护措施时公开默认端点。

### 手动 gzip HAR 测试

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```

## 通知与静态文件

从 `config/push_map.example.json`、`config/bark_map.example.json` 创建本地配置。这些文件包含玩家/设备标识，已被 Git 忽略。

### 玩家路由

`config/push_map.json` 将玩家 ID 映射为 `telegram`、Bark 别名、`none`、`+tg` 字符串或方法数组：

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "klee"],
  "1234567890123456791": "none"
}
```

没有可用路由值的玩家默认走 Telegram。

### Telegram

Telegram 将全部生成的常规 `site_*.png` 作为本地 multipart 媒体组上传。它需要 `TELEGRAM_BOT_TOKEN`、`TELEGRAM_CHAT_ID`，但不需要公网图片服务器。Telegram 失败不会阻止已配置的 Bark 尝试。

### Bark

Bark 会发送稀有资源摘要，并为全部生成的常规 `site_*.png` 分别通知。`config/bark_map.json` 负责别名与设备密钥的映射：

```json
{ "klee": "paste-your-bark-key-here" }
```

Bark 会自行抓取图片 URL。自动服务任务应将 `BARK_IMAGE_BASE` 指向公开的 `data/` 根目录，归档 URL 如下：

```text
https://maps.example.com/archive/by-id/<player_id>/<timestamp>/site_5.png
```

手动 `notify` 的图片根路径优先级为 `--image-base`、`BARK_IMAGE_BASE`、`FALLBACK_IMAGE_BASE`；该根路径应直接公开所选输出目录。

### 静态文件服务器

Bark 图片不可使用 `localhost` 或 `127.0.0.1`。请使用公网 HTTPS，例如：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    root /path/to/MySekaiMapper/data;
}
```

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

通知器会忽略输出目录中的符号链接，不会记录凭据或完整通知 URL。

## 命令行参考

先构建一次可执行文件：

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

所有命令默认加载 `.env`，并接受 `--env /path/to/file`。`--root` 可放在子命令后的任意位置。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

解密并解析存档，输出安全的聚合 JSON 摘要，不写入地图。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

解密存档、提取掉落点，并写入 `site_*.png` 和 `rare_resources.txt`。`--output` 默认使用 `data/latest`；`--assets` 可覆盖资源目录。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output` 必填。`--task-id` 和 `--player-id` 默认值为 `unknown`；需要玩家专属路由时，请传入实际玩家 ID。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

启动上传和上报 HTTP 端点，默认监听 `0.0.0.0:9478`。

## 目录结构

```text
.
├── cmd/mysekaimapper/       # CLI 入口
├── internal/
│   ├── har/                 # Reqable HAR 解析与解压
│   ├── mapper/              # AES、MsgPack、资源与渲染
│   ├── notify/              # Telegram 与 Bark 通知
│   ├── server/              # 上传与上报 HTTP 端点
│   └── service/             # 队列、存储与归档流水线
├── assets/                  # 字体和资源图标
├── config/                  # 本地路由模板
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # 被忽略的运行时输出
│   ├── tmp/                 # 上传暂存
│   ├── raw_mysekai/         # 加密源存档
│   ├── archive/             # 按玩家和时间戳保存的历史产物
│   └── latest/              # 最新生成的产物
├── docs/                    # VitePress 文档
├── go.mod / go.sum          # Go 模块定义
└── .env.example             # 配置模板
```

`data/`、`.env`、`config/bark_map.json` 和 `config/push_map.json` 是私密的运行时数据，会被 Git 忽略。

## 测试

```bash
go test ./...
go build -o /tmp/mysekaimapper ./cmd/mysekaimapper
npm run docs:build
```

GitHub Actions 会在 push 和拉取请求时运行 Go 测试套件并构建服务。

## Go 重构说明

当前运行时仅使用 Go。模块采用包含 `cmd/`、`internal/`、`go.mod` 和 `go.sum` 的标准根目录结构；Python 源码、依赖和 CI 已被移除。归档的参考实现仍保留在 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 分支和 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 标签中。

HTTP 端点、环境变量、输出名称、归档布局和路由文件格式保持兼容。Go 渲染器使用固定画布，因此生成的 PNG 不保证与此前的 Matplotlib 输出逐像素完全一致。

## 免责声明

本工具仅供个人学习和娱乐使用。请勿将其用于商业目的，或以任何违反游戏服务条款的方式使用。游戏数据和资源归其各自所有者所有。

## 许可证

项目代码采用 [MIT](https://github.com/mouse233/MySekaiMapper/blob/feat/go-rewrite/LICENSE) 许可证（Copyright © 2025 mouse233）。`assets/` 下的游戏资源和数据归其各自所有者所有，不包含在本许可证中。
