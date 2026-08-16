# 快速上手

先完成安装与 `.env` 基础配置，再按你想要的推送方式选择路径：

- **路径 A（仅 Telegram Bot 推送）**：配置最少，推荐先跑通这条；
- **路径 B（启用 Bark 推送）**：在路径 A 基础上，需要额外配置 Bark key、玩家路由与静态文件服务器。

## 1. 安装

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 可选:安装 mysekai 命令(等价于 python cli.py ...)
venv/bin/pip install -e .
```

## 2. 配置 .env（必填项）

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` 为 MySekai 存档的 AES-128-CBC 解密密钥（各 16 字节），无论走哪条路径都必须填写。其余变量按选择的路径配置：

| 变量 | 必填 | 说明 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai 存档的 AES-128-CBC 密钥，各 16 字节 |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 选* | Telegram 推送（默认渠道）需要，来自 [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | 选 | Bark 通知图标 URL |
| `BARK_IMAGE_BASE` | 选 | 静态文件服务器根地址（推送 Bark 图片直链用，见下文） |
| `FALLBACK_IMAGE_BASE` | 选 | 未配置 `BARK_IMAGE_BASE` 时的图片直链兜底地址 |

::: warning
\* 若只想用 Bark 收通知：可留空 Telegram 配置，但**必须在 `config/push_map.json` 里把玩家路由到 Bark 别名**，否则未配置玩家默认走 Telegram，而 Telegram 缺配置时只会打印一行警告并跳过，结果是什么都不推。
:::

## 3. 路径 A：仅 Telegram Bot 推送（最简）

适用场景：只要在 Telegram 收到地图与统计，不折腾其他组件。

1. 在 `.env` 中填写 Telegram 配置（来自 [@BotFather](https://t.me/BotFather)）：

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 手动跑一遍验证：

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 日常使用：启动上传服务；抓包客户端（MitM 模块 / Reqable 上报服务器）按[上传接口](/zh-CN/guide/upload-api)分片上传后，自动生成地图并推送：

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

路径 A **不需要**：`config/push_map.json`、`config/bark_map.json`、静态文件服务器、`BARK_IMAGE_BASE`。未配置的玩家默认就推送到 Telegram。

## 4. 路径 B：启用 Bark 推送（需额外配置）

在路径 A 的基础上（Telegram 配置可保留，也可留空只推 Bark），按顺序补齐：

1. **配置 Bark key**：在 `config/bark_map.json` 中为每个别名配置设备 key（模板见同目录 `bark_map.example.json`）。
2. **配置玩家路由**：在 `config/push_map.json` 中把玩家 ID 路由到 Bark 别名，例如：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ::: warning
   **必须配置**：未配置的玩家默认走 Telegram；若此时 Telegram 又未配置，只会打印警告并跳过，结果什么都不推。
   :::
3. **搭建静态文件服务器**：把项目的 `data/` 目录暴露为公网可达的 HTTP(S) 服务，并在 `.env` 设置 `BARK_IMAGE_BASE=https://<域名或IP:端口>`。否则 Bark 通知不带地图图片（详见[静态文件服务器](/zh-CN/guide/static-server)）。
4. 验证与日常使用同路径 A（第 2、3 步）。
