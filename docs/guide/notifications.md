<!-- GENERATED from README.md; do not edit directly. -->

# Notifications and static files

Create local configuration from `config/push_map.example.json` and `config/bark_map.example.json`. These files contain player/device identifiers and are ignored by Git.

### Player routing

`config/push_map.json` maps player IDs to `telegram`, Bark aliases, `astrbot:<alias>`, `none`, `+tg` strings, or arrays of methods:

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "klee"],
  "1234567890123456791": "none"
}
```

Players without an available routing value default to Telegram.

### Telegram

Telegram uploads all generated regular `site_*.png` files as a local multipart media group. It requires `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID`, but does not require a public image server. Telegram failures do not prevent configured Bark attempts.

### AstrBot Push Lite (QQ)

Install [astrbot_plugin_push_lite](https://github.com/Raven95676/astrbot_plugin_push_lite) in your existing AstrBot instance. For QQ, connect AstrBot through OneBot v11 and configure the plugin's API token and listening port (default `9966`). Create `config/astrbot_map.json` from `config/astrbot_map.example.json`:

```json
{
  "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"},
  "qq_group": {"platform_id": "qq_main", "type": "group", "qq": "987654321"},
  "session": {"umo": "qq_main:FriendMessage:123456789"}
}
```

`platform_id` is the actual bot/platform instance ID in AstrBot, not necessarily the adapter type `aiocqhttp` or the bot's QQ number. `type` must be `private` or `group`; `qq` is a **string** containing the recipient's QQ number or group number. The notifier constructs `platform_id:FriendMessage:qq` or `platform_id:GroupMessage:qq`. You can confirm the platform ID once with `/sid`; each recipient does not need to run the command. Alternatively, set `umo` alone to an existing full session identifier. Do not combine it with `platform_id`, `type`, or `qq`. See [AstrBot's session documentation](https://docs.astrbot.app/use/command.html#name).

Select aliases using `astrbot:<alias>` in `config/push_map.json`:

```json
{
  "1234567890123456789": ["astrbot:qq_me", "astrbot:qq_group"],
  "1234567890123456790": ["astrbot:qq_me", "telegram", "klee"],
  "1234567890123456791": "none"
}
```

Configure the gateway in `.env`:

```dotenv
ASTRBOT_PUSH_URL=http://astrbot:9966
ASTRBOT_PUSH_TOKEN=your-plugin-api-token
ASTRBOT_ALLOW_INSECURE_HTTP=1
```

`ASTRBOT_PUSH_URL` is the service **base URL**, without `/send`; reverse-proxy path prefixes are supported. HTTPS is required by default. Enable `ASTRBOT_ALLOW_INSECURE_HTTP=1` only for trusted local/Docker networks. With separate containers, use the AstrBot service name on a shared network; `localhost` refers to the current container. For a host-based service, use its reachable address and mapped plugin port. The notifier sends the token as a Bearer header and does not follow redirects.

Both `serve` and `notify` submit a text summary followed by every regular `site_*.png` as a separate Base64 image request, per recipient. No public image server or shared filesystem is required. Images are limited to 8 MiB each before encoding; allow sufficient request size in your reverse proxy. Duplicate AstrBot aliases in one route are submitted once. A failed target/request does not prevent other selected targets or channels from being attempted.

A successful response means **queued in Push Lite**, not confirmed QQ delivery; logs explicitly say `request queued`. This integration does not use delivery callbacks or automatically retry. The plugin's queue is held in memory, so pending messages can be lost on restart. Check AstrBot's logs for downstream results. Explicitly assign AstrBot routes: players without a routing value still default to Telegram. Keep `.env` and `config/astrbot_map.json` private; both are ignored by Git.

### Bark

Bark sends summaries of rare resources and separate notifications for all generated regular `site_*.png` files. `config/bark_map.json` maps aliases to device keys:

```json
{ "klee": "paste-your-bark-key-here" }
```

Bark fetches image URLs itself. Automated service tasks should point `BARK_IMAGE_BASE` to the public `data/` root directory. Archive URLs are formatted as follows:

```text
https://maps.example.com/archive/by-id/<player_id>/<timestamp>/site_5.png
```

For manual `notify`, the image root path precedence is `--image-base`, `BARK_IMAGE_BASE`, then `FALLBACK_IMAGE_BASE`; the root path should directly expose the selected output directory publicly.

### Static file server

Bark images cannot use `localhost` or `127.0.0.1`. Use public HTTPS, for example:

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

The notifier ignores symbolic links in the output directory and does not log credentials or complete notification URLs.
