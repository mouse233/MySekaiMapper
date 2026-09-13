<!-- GENERATED from README.md; do not edit directly. -->

# Notifications and static files

Create local configuration from `config/push_map.example.json` and `config/bark_map.example.json`. These files contain player/device identifiers and are ignored by Git.

### Player routing

`config/push_map.json` maps player IDs to `telegram`, Bark aliases, `none`, `+tg` strings, or arrays of methods:

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
