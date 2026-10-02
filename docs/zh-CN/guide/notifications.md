<!-- GENERATED from doc/README.zh-CN.md; do not edit directly. -->

# 通知与静态文件

从 `config/push_map.example.json`、`config/bark_map.example.json` 创建本地配置。这些文件包含玩家/设备标识，已被 Git 忽略。

### 玩家路由

`config/push_map.json` 将玩家 ID 映射为 `telegram`、Bark 别名、`astrbot:<别名>`、`none`、`+tg` 字符串或方法数组：

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

### AstrBot Push Lite（QQ）

在现有 AstrBot 实例中安装 [astrbot_plugin_push_lite](https://github.com/Raven95676/astrbot_plugin_push_lite)。QQ 使用 OneBot v11 接入 AstrBot，并配置插件的 API token 和监听端口（默认 `9966`）。从 `config/astrbot_map.example.json` 创建 `config/astrbot_map.json`：

```json
{
  "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"},
  "qq_group": {"platform_id": "qq_main", "type": "group", "qq": "987654321"},
  "session": {"umo": "qq_main:FriendMessage:123456789"}
}
```

`platform_id` 是 AstrBot 中实际的机器人／平台实例 ID，不一定是适配器类型 `aiocqhttp`，也不是机器人 QQ 号。`type` 必须为 `private`（私聊）或 `group`（群聊）；`qq` 用**字符串**填写接收人的 QQ 号或群号。程序自动构造 `platform_id:FriendMessage:qq` 或 `platform_id:GroupMessage:qq`。只需通过 `/sid` 确认一次平台 ID，无需每个接收人执行命令。也可仅填写 `umo` 使用完整会话标识，但不能与 `platform_id`、`type`、`qq` 混用。参见 [AstrBot 会话文档](https://docs.astrbot.app/use/command.html#name)。

在 `config/push_map.json` 中使用 `astrbot:<别名>` 选择接收人，可与现有渠道组合：

```json
{
  "1234567890123456789": ["astrbot:qq_me", "astrbot:qq_group"],
  "1234567890123456790": ["astrbot:qq_me", "telegram", "klee"],
  "1234567890123456791": "none"
}
```

在 `.env` 配置推送服务：

```dotenv
ASTRBOT_PUSH_URL=http://astrbot:9966
ASTRBOT_PUSH_TOKEN=your-plugin-api-token
ASTRBOT_ALLOW_INSECURE_HTTP=1
```

`ASTRBOT_PUSH_URL` 填写服务**根地址**，不包含 `/send`，支持反向代理路径前缀。默认要求 HTTPS；仅在可信本机或 Docker 私有网络使用 HTTP 时设置 `ASTRBOT_ALLOW_INSECURE_HTTP=1`。不同容器应加入同一网络并使用 AstrBot 服务名，`localhost` 指向当前容器；宿主机部署则使用可达地址和插件映射端口。请求使用 Bearer token，程序不会跟随重定向。

`serve` 和 `notify` 都会对每个接收人先提交文字摘要，再将全部常规 `site_*.png` 分别以 Base64 图片提交，无需公网图片服务器或共享目录。每张图片编码前最多 8 MiB，反向代理也需允许相应请求大小。同一路由中重复的 AstrBot 别名只发送一次；某个接收人或请求失败，不会阻止其他接收人和渠道的尝试。

成功响应表示**已进入 Push Lite 队列**，并不代表 QQ 已送达；日志会明确记录 `request queued`。本接入不使用送达回调，也不自动重试。插件队列在内存中，重启可能丢失待发消息，最终发送结果请查看 AstrBot 日志。需显式配置 AstrBot 路由，没有可用路由值的玩家仍默认 Telegram。`.env` 和 `config/astrbot_map.json` 已被 Git 忽略，请保持私密。

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
