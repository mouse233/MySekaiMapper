# Push Mechanism

## Telegram Bot by default

- Players not configured in `config/push_map.json` **always default to Telegram**; the same applies when `push_map.json` is missing.
- Telegram uses the Bot API `sendMediaGroup` to upload the 4 local PNGs directly as multipart — **no public image URL and no static file server needed**; if `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` are missing it just prints a warning and skips, without affecting the Bark channel.

## Bark push requires public image links

Images in Bark (Day.app) notifications are **URL links**: `notify.py` encodes the image address into the `image=` parameter sent to `api.day.app`, and the Bark server fetches that image itself. The URL must therefore be **publicly reachable (HTTPS recommended)**, otherwise Bark notifications have no images.

The 4 map links are composed by `notify.py` with this precedence:

```python
base = image_base or BARK_IMAGE_BASE or FALLBACK_IMAGE_BASE
image_url = base.rstrip("/") + f"/site_{i}.png"   # i = 5..8
```

| Scenario | base value | Image link form |
| --- | --- | --- |
| Server flow (recommended) | `BARK_IMAGE_BASE` + `/archive/by-id/<user_id>/<timestamp>` | `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<timestamp>/site_{5..8}.png` |
| Manual CLI push | `BARK_IMAGE_BASE` or `FALLBACK_IMAGE_BASE` | `<base>/site_{5..8}.png` (expose `data/latest/` under `<base>/`) |

::: tip
The server flow only composes archive-path links when `BARK_IMAGE_BASE` is configured; with only `FALLBACK_IMAGE_BASE` set, the server pushes `<FALLBACK_IMAGE_BASE>/site_{5..8}.png` links too.
:::

See [Static File Server](/guide/static-server) for how to expose the images, and [Player Routing](/guide/routing) for how a player is assigned to Telegram vs Bark.
