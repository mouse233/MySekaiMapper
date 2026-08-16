# Quick Start

First finish the installation and basic `.env` configuration, then pick the path that matches your push setup:

- **Path A (Telegram Bot only)**: fewest configs, recommended to get running first;
- **Path B (enable Bark push)**: Path A plus Bark keys, player routing, and a static file server.

## 1. Install

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# Optional: install the mysekai command (equivalent to python cli.py ...)
venv/bin/pip install -e .
```

## 2. Configure .env (required)

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` are the AES-128-CBC decryption keys for MySekai saves (16 bytes each) — required on every path. The remaining variables depend on your chosen path:

| Variable | Required | Description |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | AES-128-CBC keys for MySekai saves, 16 bytes each |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | Optional* | Needed for Telegram push (default channel); from [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | Optional | Icon URL for Bark notifications |
| `BARK_IMAGE_BASE` | Optional | Root URL of the static file server (for Bark image links; see below) |
| `FALLBACK_IMAGE_BASE` | Optional | Fallback base URL for image links when `BARK_IMAGE_BASE` is not set |

::: warning
\* If you only want Bark notifications: you may leave the Telegram config empty, but you **must route the player to a Bark alias in `config/push_map.json`**, otherwise unconfigured players default to Telegram — and with Telegram unconfigured, only a warning is printed and nothing is pushed.
:::

## 3. Path A: Telegram Bot only (simplest)

Use when: you just want maps and stats in Telegram without setting up anything else.

1. Fill in the Telegram config in `.env` (from [@BotFather](https://t.me/BotFather)):

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. Run it once manually to verify:

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. Daily use: start the upload service; the capture client (MitM module / Reqable Report Server) uploads chunks per the [Upload API](/guide/upload-api), then maps are generated and pushed automatically:

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

Path A does **not** need: `config/push_map.json`, `config/bark_map.json`, a static file server, or `BARK_IMAGE_BASE`. Unconfigured players are pushed to Telegram by default.

## 4. Path B: enable Bark push (extra configuration)

On top of Path A (the Telegram config may stay, or be left empty to push only to Bark), set up in order:

1. **Configure Bark keys**: give each alias a device key in `config/bark_map.json` (template: `bark_map.example.json` in the same directory).
2. **Configure player routing**: route player IDs to Bark aliases in `config/push_map.json`, for example:

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ::: warning
   **Required**: unconfigured players default to Telegram; if Telegram is also unconfigured, only a warning is printed and nothing is pushed.
   :::
3. **Set up a static file server**: expose the project's `data/` directory as a publicly reachable HTTP(S) service and set `BARK_IMAGE_BASE=https://<domain-or-ip:port>` in `.env`. Otherwise Bark notifications carry no map images (see [Static file server examples](/guide/static-server)).
4. Verify and use daily the same as Path A (steps 2 and 3).
