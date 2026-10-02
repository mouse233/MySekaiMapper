# MySekaiMapper

🌐 **Languages**: [English](README.md) · [简体中文](doc/README.zh-CN.md) · [繁體中文](doc/README.zh-TW.md) · [日本語](doc/README.ja-JP.md) · [한국어](doc/README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/>

A Go service that turns encrypted *Project SEKAI* MySekai saves into resource-gathering maps and sends the result to Telegram, Bark (Day.app), or AstrBot (QQ).

It works with a MitM capture client or Reqable's **Report Server**: the capture tool uploads a MySekai save, the service decrypts and parses it, renders maps and a rare-resource summary, archives the artifacts, and dispatches notifications without a manual processing step.

The usual MySekai areas produce `site_5.png` (Grassland), `site_6.png` (Beach), `site_7.png` (Flower Garden), `site_8.png` (Memorial Place), and `rare_resources.txt`. The renderer and notifier also handle any additional regular `site_*.png` outputs.

The capture flow has been verified on the CN and TW servers operated by Nuverse. Availability on other regions depends on their API path and save format.

## How it works

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```

## Quick start

Choose the notification path that fits your setup:

- **Path A — Telegram only**: simplest option; no player-routing file or public image server is needed.
- **Path B — Bark enabled**: configure Bark keys, player routing, and a public static-file server for images.
- **Path C — AstrBot (QQ)**: configure Push Lite and recipient aliases; see the AstrBot section under Notifications and static files. No public image server is needed.

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
| `ASTRBOT_PUSH_URL`, `ASTRBOT_PUSH_TOKEN` | AstrBot only | Push Lite base URL and API token; see notifications below |
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
2. Create `config/push_map.json` from `config/push_map.example.json`, mapping player IDs to `bark:<alias>`, `telegram`, `none`, or a combination:

   ```json
   {
     "1234567890123456789": ["bark:klee"],
     "1234567890123456790": ["telegram", "bark:klee"],
     "1234567890123456791": "none"
   }
   ```

3. Expose the repository's `data/` directory through a public HTTP(S) static-file server and set its public root as `BARK_IMAGE_BASE`:

   ```dotenv
   BARK_IMAGE_BASE=https://maps.example.com
   ```

An unconfigured player defaults to Telegram. If Telegram is not configured, an unconfigured player therefore receives no notification; explicitly assign a Bark alias for Bark-only use.

## Running the service

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The server prints ready URLs and writes lifecycle logs for upload/report acceptance, queueing, parsing, rendering, archiving, notifications, elapsed time, task ID, and `player_id`. It deliberately avoids logging archive bodies, secrets, tokens, or complete notification URLs.

The process handles `SIGINT` and `SIGTERM`: it stops accepting HTTP requests, then drains already accepted jobs for up to 15 seconds.

A compiled binary can run outside the checkout with `--root /path/to/MySekaiMapper`; otherwise the repository root is discovered from the working directory.

## Upload API

`POST /uploadMySekai` accepts the encrypted MySekai response body directly. A single upload is normally enough; ordered chunks remain supported for capture-client compatibility.

| Header | Required | Description |
| --- | --- | --- |
| `X-Upload-Id` | Yes | Task identifier matching `^[A-Za-z0-9_-]{1,64}$` |
| `X-Chunk-Index` | Yes | Zero-based chunk index |
| `X-Total-Chunks` | Yes | Total chunk count, from 1 through 10 |
| `X-Original-Url` | No | Original game URL; `/user/<id>` supplies the player route |
| `X-Script-Version` | No | Accepted for capture-client compatibility and ignored by the service |

The encrypted archive, each chunk, and the merged upload are limited to 1 MiB. A successfully accepted request returns plain-text `OK`; rendering and notification continue in the background.

### Single-upload example

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### Chunked-upload example

Use a shared `X-Upload-Id`, ordered indices, and at most ten chunks:

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

Common responses are `200 OK` for an accepted upload, `400 Bad Request` for invalid identifiers or chunk ranges, `413 Payload Too Large` for a size limit, and `422 Unprocessable Entity` for missing or non-integer required upload headers.

## Reqable Report Server

Reqable v2.20.0+ can POST each captured HTTP session to this service as HAR JSON. The report endpoint is enabled by default and coexists with `/uploadMySekai`.

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| Variable | Default | Description |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | Set `0`, `false`, `no`, or `off` to disable reports |
| `REPORT_PATH` | `/reqable/report` | Endpoint path configured in Reqable |
| `REPORT_MAX_SIZE` | `1` | Maximum decompressed HAR body size in MiB |
| `REPORT_TOKEN` | empty | Optional value required in `X-Report-Token` |

### Processing flow

For each report, the service:

1. Decompresses `identity`, `gzip`, `br`, `zstd`, or `zstandard` content and parses the HAR. Streamed zstd frames without a content-size field are supported.
2. Walks `log.entries` and accepts the first response body (falling back to its request body) that decrypts with `AES_KEY`/`AES_IV` and validates as a MySekai archive.
3. Extracts `player_id` from `/user/<id>` in the matched session URL.
4. Saves the encrypted archive in `data/raw_mysekai/` and starts the same render → archive → notify pipeline used by uploads.

> Reqable reports each session once and does not retry. Keep the service available and watch `[REPORT]` logs. A syntactically valid HAR with no MySekai archive still receives `ok`; only the first valid archive in a report is processed.

### Configure Reqable

- **Matching rule**: `https://<game-api-domain>/api/user/*/mysekai*`
- **Server URL**: `http://<your-server>:9478/reqable/report` (or your custom `REPORT_PATH`)

| Server | Game API domain |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

The matching pattern has been verified for CN. If your region uses another MySekai API path, inspect its captured URL and adjust the rule.

### Security

Reqable cannot add the custom `X-Report-Token` header. Use a long random `REPORT_PATH` such as `/reqable/report/<random>` and restrict access through a reverse proxy or firewall; do not expose the default endpoint publicly without controls.

### Manual gzip HAR test

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```

## Notifications and static files

Create local configuration from `config/push_map.example.json` and `config/bark_map.example.json`. These files contain player/device identifiers and are ignored by Git.

### Player routing

`config/push_map.json` maps player IDs to `telegram`, `bark:<alias>`, `astrbot:<alias>`, `none`, prefixed `+tg` strings, or arrays of methods:

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

Players without an available routing value default to Telegram.

Bark routes must use `bark:<alias>`; AstrBot routes use `astrbot:<alias>`. When upgrading, change bare Bark targets such as `"sls"` and `"xufan"` to `"bark:sls"` and `"bark:xufan"` in `push_map.json`. Keys in `bark_map.json` stay unchanged (`"sls"`, `"xufan"`). Bare aliases are rejected with a routing error; other valid selected targets are still attempted. The string shorthand `"bark:klee+tg"` is supported, but `["bark:klee", "telegram"]` is recommended.

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
  "1234567890123456790": ["astrbot:qq_me", "telegram", "bark:klee"],
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

## Command-line reference

Build the binary once:

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

All commands load `.env` by default and accept `--env /path/to/file`. `--root` may appear anywhere after the subcommand.

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

Decrypts and parses a save, then prints a safe aggregate JSON summary without writing maps.

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

Decrypts the archive, extracts drops, and writes `site_*.png` plus `rare_resources.txt`. `--output` defaults to `data/latest`; `--assets` can override the asset directory.

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

`--output` is required. `--task-id` and `--player-id` default to `unknown`; pass the actual player ID whenever player-specific routing is required.

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

Starts the upload and report HTTP endpoints. Defaults are `0.0.0.0:9478`.

## Directory structure

```text
.
├── cmd/mysekaimapper/       # CLI entry point
├── internal/
│   ├── har/                 # Reqable HAR parsing and decompression
│   ├── mapper/              # AES, MsgPack, resources, and rendering
│   ├── notify/              # Telegram, Bark, and AstrBot dispatch
│   ├── server/              # Upload and report HTTP endpoints
│   └── service/             # Queue, storage, and archive pipeline
├── assets/                  # Font and resource icons
├── config/                  # Local routing templates
│   ├── astrbot_map.example.json
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # Ignored runtime data
│   ├── tmp/                 # Upload staging
│   ├── raw_mysekai/         # Encrypted source archives
│   ├── archive/             # Historical artifacts by player and timestamp
│   └── latest/              # Latest generated artifacts
├── docs/                    # VitePress documentation
├── go.mod / go.sum          # Go module definition
└── .env.example             # Configuration template
```

`data/`, `.env`, `config/bark_map.json`, `config/push_map.json`, and `config/astrbot_map.json` are private runtime data and are ignored by Git.

## Testing

```bash
go test ./...
go build -o /tmp/mysekaimapper ./cmd/mysekaimapper
npm run docs:build
```

GitHub Actions runs the Go test suite and build for pushes and pull requests.

## Go refactor

The active runtime is Go-only. The module follows the standard root layout with `cmd/`, `internal/`, `go.mod`, and `go.sum`; Python source, dependencies, and CI were removed. The archived reference implementation remains in the [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) branch and [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) tag.

The HTTP endpoints, environment variables, output names, archive layout, and routing-file formats remain compatible apart from the required `bark:` route prefix. The Go renderer uses a fixed canvas, so its generated PNGs are not guaranteed to be pixel-identical to the former Matplotlib output.

## Disclaimer

This tool is for personal learning and entertainment only. Do not use it for commercial purposes or in ways that violate the game's terms of service. Game data and assets belong to their respective owners.

## License

Project code is licensed under [MIT](LICENSE) (Copyright © 2025 mouse233). Game assets and game data under `assets/` belong to their respective owners and are not covered by this license.
