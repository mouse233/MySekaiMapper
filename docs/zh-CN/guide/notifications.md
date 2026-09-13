<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 通知与静态文件

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
