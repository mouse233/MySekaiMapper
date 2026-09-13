<!-- GENERATED from README.md; do not edit directly. -->

# Quick start

Choose the notification path that fits your setup:

- **Path A — Telegram only**: simplest option; no player-routing file or public image server is needed.
- **Path B — Bark enabled**: configure Bark keys, player routing, and a public static-file server for images.

### 1. Requirements and build

Go **1.25 or newer** is required.

```bash
go version
cp .env.example .env
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

`AES_KEY` and `AES_IV` in `.env` are required 16-byte AES-128-CBC values. Do not commit `.env` or local routing files.

### 2. Configure `.env`

| Variable | Required | Description |
| --- | --- | --- |
| `AES_KEY`, `AES_IV` | Yes | 16-byte MySekai AES-128-CBC key and IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | Telegram only | Bot credentials and target chat ID from [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | Optional | Icon URL included in Bark notifications |
| `BARK_IMAGE_BASE` | Bark images | Public base URL for archived map images |
| `FALLBACK_IMAGE_BASE` | Optional | Image-base fallback when `BARK_IMAGE_BASE` is unset |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | Optional | Reqable report-endpoint settings |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | Optional | Override the default repository directories |

### 3. Path A — Telegram only

1. Set the Telegram variables in `.env`:

   ```dotenv
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. Optionally verify parsing and notification with an existing encrypted save:

   ```bash
   bin/mysekaimapper generate --input data/raw_mysekai/mysekai.bin
   bin/mysekaimapper notify \
     --output data/latest \
     --task-id manual-001 \
     --player-id 1234567890123456789
   ```

3. Start the service for normal operation:

   ```bash
   bin/mysekaimapper serve --host 0.0.0.0 --port 9478
   ```

Players absent from `config/push_map.json` default to Telegram. Path A does not require a Bark map, a push map, or a public image server.

### 4. Path B — enable Bark

In addition to the Path A configuration (Telegram may be omitted for Bark-only routes):

1. Create `config/bark_map.json` from `config/bark_map.example.json`, mapping a Bark alias to each device key.
2. Create `config/push_map.json` from `config/push_map.example.json`, mapping player IDs to a Bark alias, `telegram`, `none`, or a combination:

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"],
     "1234567890123456791": "none"
   }
   ```

3. Expose the repository's `data/` directory through a public HTTP(S) static-file server and set its public root as `BARK_IMAGE_BASE`:

   ```dotenv
   BARK_IMAGE_BASE=https://maps.example.com
   ```

An unconfigured player defaults to Telegram. If Telegram is not configured, an unconfigured player therefore receives no notification; explicitly assign a Bark alias for Bark-only use.
