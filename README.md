# MySekaiMapper

🌐 **Languages**: [简体中文](doc/README.zh-CN.md) · [繁體中文](doc/README.zh-TW.md) · [日本語](doc/README.ja-JP.md) · [한국어](doc/README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/>

A resource-gathering point map generator for the MySekai mode in *Project Sekai* (世界计划 多彩舞台).

**Original intent**: designed to work with MitM modules or Reqable's "Report Server" feature — the capture tool grabs MySekai data packets from the game and automatically uploads them to this service (single POST; chunked upload is also supported). The server decrypts the encrypted saves, extracts the resource drop coordinates of every station, draws gathering maps, and pushes the results (including a rare-resource summary) to the player's Telegram / Bark (iOS Day.app) — no manual intervention required.

Each task produces **4 maps**: `site_5.png` (Grassland), `site_6.png` (Beach), `site_7.png` (Flower Garden), `site_8.png` (Memorial Place), plus a `rare_resources.txt` rare-resource summary.

This project has been tested and verified on the CN and TW servers operated by Nuverse (朝夕光年). Availability on other servers is unknown.

## How it works

```
Game API response → MitM module / Reqable Report Server (captures mysekai data)
   │  ① Auto upload (single POST; chunked supported) → server.py processes automatically
   │  ② Or drop a .bin save manually → cli.py generate
   ▼
parser.py    AES-128-CBC decrypt + msgpack parse + coordinate rotation
   ▼
render.py    Draw site_5.png ~ site_8.png + rare_resources.txt → data/latest/
   ▼
notify.py    Push:
             ├─ Telegram: images uploaded directly as multipart, no public URL needed ← default channel
             └─ Bark: notified with image= URL links, requires a static file server
```

## Quick start

First finish the installation and basic `.env` configuration, then pick the path that matches your push setup:

- **Path A (Telegram Bot only)**: fewest configs, recommended to get running first;
- **Path B (enable Bark push)**: Path A plus Bark keys, player routing, and a static file server.

### 1. Install

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# Optional: install the mysekai command (equivalent to python cli.py ...)
venv/bin/pip install -e .
```

### 2. Configure .env (required)

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

> \* If you only want Bark notifications: you may leave the Telegram config empty, but you **must route the player to a Bark alias in `config/push_map.json`**, otherwise unconfigured players default to Telegram — and with Telegram unconfigured, only a warning is printed and nothing is pushed.

### 3. Path A: Telegram Bot only (simplest)

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

3. Daily use: start the upload service; saves are turned into maps and pushed automatically. Two capture clients are supported:

   - **MitM module**: uploads the save per the [Upload API](#upload-api)
   - **Reqable Report Server**: reports captured sessions to the built-in endpoint (see [Reqable Report Server](#reqable-report-server))

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

Path A does **not** need: `config/push_map.json`, `config/bark_map.json`, a static file server, or `BARK_IMAGE_BASE`. Unconfigured players are pushed to Telegram by default.

### 4. Path B: enable Bark push (extra configuration)

On top of Path A (the Telegram config may stay, or be left empty to push only to Bark), set up in order:

1. **Configure Bark keys**: give each alias a device key in `config/bark_map.json` (template: `bark_map.example.json` in the same directory).
2. **Configure player routing**: route player IDs to Bark aliases in `config/push_map.json`, for example:

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ⚠️ **Required**: unconfigured players default to Telegram; if Telegram is also unconfigured, only a warning is printed and nothing is pushed.
3. **Set up a static file server**: expose the project's `data/` directory as a publicly reachable HTTP(S) service and set `BARK_IMAGE_BASE=https://<domain-or-ip:port>` in `.env`. Otherwise Bark notifications carry no map images (see [Static file server examples](#static-file-server-examples-optional) below).
4. Verify and use daily the same as Path A (steps 2 and 3).

## Upload API

This endpoint receives the captured mysekai response body via `POST /uploadMySekai` (a single POST; chunked upload is kept for compatibility). The same protocol can be debugged manually with curl. Headers:

| Header | Description |
| --- | --- |
| `X-Upload-Id` | Upload task ID (alphanumeric plus `-` / `_`, length 1~64), required |
| `X-Chunk-Index` | Chunk index, starting at 0 (always 0 for a single POST), required |
| `X-Total-Chunks` | Total number of chunks (1~10; use 1 for a single POST), required |
| `X-Original-Url` | The client's original page URL, used to resolve the player ID (e.g. `https://.../user/123456...`); **optional** — if missing, the player ID is recorded as `unknown` |
| `X-Script-Version` | Client script version; ignored by the server, may be omitted |

The request body is the raw binary save data (no multipart needed).

Limits:

- Total file size ≤1MB (`MAX_TOTAL_SIZE`)
- Single chunk ≤1MB (`MAX_CHUNK_SIZE`, returns 413 if exceeded)
- Max 10 chunks (`MAX_CHUNKS`)

> Note: current saves are ~200KB, so a **single POST** is all you need. Chunked upload is kept for compatibility with older capture clients; if used, keep each chunk well below 1MB (e.g. 256KB) so 10 chunks fill the 1MB cap.

Responses:

| Status | Meaning |
| --- | --- |
| `200` | Save received, returns `OK`; the server automatically: merges the save (if chunked) → generates maps → archives to `data/archive/by-id/<user_id>/<timestamp>/` → pushes notifications. No manual intervention. |
| `400` | Invalid parameters (bad upload id format, chunk index out of range, total chunks not in 1~10) |
| `413` | Size limit exceeded (single chunk over 1MB, or cumulative total over 1MB) |

### curl examples

Single POST (all current saves fit in one request):

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

Chunked upload (optional, for compatibility; 256KB per chunk fills the 1MB cap with 10 chunks):

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
total=$(( ($(wc -c < "$file") + 262143) / 262144 ))
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_

i=0
for c in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H "X-Original-Url: https://example.com/user/1234567890123456789" \
    --data-binary @"$c"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

A `200 OK` means the save was accepted; the pipeline (merge if chunked → generate → archive → notify) runs automatically. Replace `127.0.0.1:9478` with your actual service address; `X-Upload-Id` must match `^[a-zA-Z0-9_-]{1,64}$` (e.g. a random string from `openssl rand -hex 5`).

## Reqable Report Server

Instead of a custom capture client, you can use Reqable's built-in **Report Server** feature (Reqable v2.20.0+): it automatically POSTs each captured HTTP session to your server in the [HAR](https://en.wikipedia.org/wiki/HAR_(file_format)) JSON format, optionally compressed with gzip / brotli / zstd. The report endpoint is **enabled by default** and coexists with the chunked upload API — `python cli.py server` serves both. Set `REPORT_ENABLED=0` to disable it:

```bash
python cli.py server
```

Configuration (`.env`):

| Variable | Default | Description |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` (on) | Set to `0` / `false` to disable the report endpoint |
| `REPORT_PATH` | `/reqable/report` | Endpoint path; fill this into the Reqable "Upload Path" field |
| `REPORT_MAX_SIZE` | `1` | Max HAR request body size in MB, same as the chunked upload limit |
| `REPORT_TOKEN` | *(empty)* | Optional shared token; when set, the endpoint requires the `X-Report-Token` header |

What the endpoint does with each report:

1. Decompresses the body (`Content-Encoding: gzip` / `br` / `zstd`) and parses the HAR.
2. Walks `log.entries` and takes the first session whose response body (fallback: request body) decrypts with `AES_KEY`/`AES_IV` and parses as a MySekai save — unrelated API traffic matching the rule is skipped.
3. Resolves the player ID from the session URL (`/user/<id>`, same rule as `X-Original-Url`).
4. Saves the archive to `data/raw_mysekai/` and launches the same generate → archive → notify pipeline as chunked uploads.

Notes:

- Reqable sends each session **exactly once and never retries**, so the endpoint answers `200` as fast as possible; make sure your server is stable and watch the `[REPORT]` log lines.
- Only **one** archive per report is processed (the first valid entry), so a rule matching many endpoints won't cause duplicate pushes.
- Security: the protocol has no built-in auth. Since Reqable cannot attach custom headers, prefer embedding a random secret in `REPORT_PATH` (e.g. `/reqable/report/9f3a…`) or restrict access with a reverse proxy / firewall instead of relying on `REPORT_TOKEN`.

Example Reqable configuration:

- URL matching rule: `https://<game-api-host>/api/user/*/mysekai*`
- Upload path: `http://<your-server>:9478/reqable/report`
- Compression: any of gzip / brotli / zstd (server supports all three)

Game API domains (one per region):

| Region | Game API domain |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

Recommended matching rule: `https://<domain>/api/user/*/mysekai*` (verified on CN). If your region's mysekai API path differs, adjust the rule accordingly.

Manual curl test (gzip-compressed HAR):

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```

## Push mechanism

### Telegram Bot by default

- Players not configured in `config/push_map.json` **always default to Telegram**; the same applies when `push_map.json` is missing.
- Telegram uses the Bot API `sendMediaGroup` to upload the 4 local PNGs directly as multipart — **no public image URL and no static file server needed**; if `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` are missing it just prints a warning and skips, without affecting the Bark channel.

### Bark push requires public image links

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

> Note: the server flow only composes archive-path links when `BARK_IMAGE_BASE` is configured; with only `FALLBACK_IMAGE_BASE` set, the server pushes `<FALLBACK_IMAGE_BASE>/site_{5..8}.png` links too.

## Static file server examples (optional)

Purpose: expose the `data/archive/` directory as a public URL so the Bark server can fetch the four maps.

**Recommended setup**: point the static server root at the project's `data/`, then set `BARK_IMAGE_BASE=https://<your-domain-or-ip:port>` for automatic mapping:

```
data/archive/by-id/<user_id>/<timestamp>/site_5.png
  →  https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<timestamp>/site_5.png
```

Common examples:

Python built-in (simplest; LAN/testing):

```bash
python -m http.server 8000 --directory data
# then set BARK_IMAGE_BASE=http://<server-ip>:8000
```

nginx:

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    # ... ssl certificate config ...
    root /path/to/MySekaiMapper/data;
}
```

Caddy (automatic HTTPS):

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

Notes:

- **Don't use `127.0.0.1` / `localhost`** as the link address; the Bark server must be able to reach it. In general, pick a publicly reachable address; LAN IPs only when connectivity is confirmed.
- **Telegram-only users need no static server at all** — skip this section.
- Manual `cli.py notify` links carry no archive path: expose `data/latest/` under `BARK_IMAGE_BASE` separately, or point `FALLBACK_IMAGE_BASE` at the output directory (e.g. `FALLBACK_IMAGE_BASE=http://<host>:5500/output` → that server mounts `data/latest/` at `/output`).

## Player push routing (optional)

Create local configs under `config/` as needed (formats follow the `*.example.json` templates in the same directory; these files are `.gitignore`d):

- `push_map.json` — player ID → push method: the value can be `"telegram"`, a Bark alias, `"none"` (no push), or a combination like `["alias", "telegram"]` / `"alias+tg"`. **Unconfigured players default to `telegram`**.

  ```json
  {
    "1234567890123456789": ["telegram"],
    "1234567890123456790": ["telegram", "klee"]
  }
  ```

- `bark_map.json` — Bark alias → device key:

  ```json
  { "klee": "paste-your-bark-key-here" }
  ```

## FAQ

- **Bark notifications have no images?** Check whether the link is publicly reachable: open `https://<BARK_IMAGE_BASE>/archive/by-id/<user_id>/<timestamp>/site_5.png` in a browser or over cellular data — it should show the image. LAN addresses, `127.0.0.1`, or HTTPS with certificate problems all make the fetch fail.
- **Nothing was pushed?** Check whether `push_map.json` sets that player to `"none"`; whether Bark-only users forgot to assign a Bark alias to the player (unconfigured players default to Telegram); whether the Telegram channel has a token and chat id; whether the Bark channel lacks a key (look for `[BARK] ... failed` in the logs).
- **Don't want Bark, only Telegram?** Nothing to do — unconfigured players already default to Telegram.

## CLI (cli.py)

Everything can be driven through `cli.py`; after installing (`pip install -e .`), the equivalent `mysekai` command is also available. Commands exit with 0 on success and 1 on error (errors print to stderr).

```bash
python cli.py --help           # subcommand overview
python cli.py <command> --help # show a subcommand's arguments
```

### generate — decrypt a save and generate maps

```bash
python cli.py generate <mysekai_bin>
```

- `<mysekai_bin>`: path to the encrypted save (.bin), required
- Flow: AES-128-CBC decrypt → msgpack parse → extract drop coordinates → draw 4 maps (`site_5.png` ~ `site_8.png`) → write `rare_resources.txt`
- Output goes to `data/latest/`; the actual path is printed at the end
- Requirements: `AES_KEY` / `AES_IV` configured in `.env`; exits with an error if the save contains no drop points

### notify — push maps and stats

```bash
python cli.py notify <output_dir> [task_id]
```

- `<output_dir>`: directory containing `site_*.png` and `rare_resources.txt` (usually `data/latest/`)
- `[task_id]`: optional upload task ID, defaults to `unknown`. Used to look up the player ID from `data/raw_mysekai/`: it first tries to match `mysekai_<playerID>_<task_id>.bin`, otherwise falls back to the newest save in raw_mysekai
- Telegram vs Bark is decided by the routing in `config/push_map.json` (unconfigured players default to Telegram); see [Player push routing](#player-push-routing-optional)

### server — start the upload service (chunked upload + Reqable report server)

```bash
python cli.py server [--host 0.0.0.0] [--port 9478]
```

- Starts the FastAPI service; clients upload encrypted saves to `POST /uploadMySekai` (single POST or chunked; protocol details: [Upload API](#upload-api)), and Reqable can report HAR sessions to the built-in report endpoint (see [Reqable Report Server](#reqable-report-server))
- When all chunks arrive, the server automatically: merges the save → generates maps → archives to `data/archive/by-id/<user_id>/<timestamp>/` → pushes notifications per player routing. No manual intervention.
- Listens on `9478` by default; for public deployment, expose it as HTTPS via a reverse proxy — the hardcoded upload URL (including the port) in your client script must match your actual deployment

### Typical manual flow

```bash
python cli.py generate mysekai_xxx.bin       # 1. generate maps to data/latest/
python cli.py notify data/latest <task_id>   # 2. push (task_id = upload ID, e.g. chfto53c3)
```

## Directory structure

```
├── app/                       # core package
│   ├── config.py              # centralized paths / env vars / local config
│   ├── crypto.py              # MySekai save AES-128-CBC decryption
│   ├── parser.py              # msgpack parsing + station coordinate rotation (pure functions)
│   ├── render.py              # extract drop points → matplotlib drawing + rare-resource stats
│   ├── notify.py              # push: Telegram media groups / Bark, per-player routing
│   ├── server.py              # FastAPI chunked upload service
│   └── cli.py                 # CLI entry
├── assets/                    # static assets (committed to the repo)
│   ├── resourceId.csv         # item ID → name + icon (base64)
│   └── NotoSansSC-Regular.ttf # Chinese font (OFL license)
├── config/                    # local configs (real files not committed; see *.example.json)
│   ├── bark_map.example.json  # Bark alias → device key template
│   └── push_map.example.json  # player ID → push method template
├── data/                      # runtime data (whole directory gitignored)
│   ├── tmp/                   # chunk upload staging, cleaned after merge
│   ├── raw_mysekai/           # merged original (encrypted) saves, kept permanently
│   ├── archive/               # historical output archive by-id/<user>/<timestamp>/ (Bark links point here)
│   └── latest/                # most recent output
├── cli.py                     # unified entry
├── tests/                     # unit tests (pytest)
├── .env.example               # env var template (copy to .env and fill in)
└── requirements.txt           # runtime dependencies (pinned)
```

## Testing

```bash
python -m pytest
```

## Disclaimer

This tool is for personal learning and entertainment only. Do not use it for any commercial purpose or in ways that violate the game's terms of service. Game data and art assets belong to their respective owners.

## License

The project's code is licensed under the [MIT License](LICENSE) (Copyright © 2025 mouse233) — free to use, modify, and redistribute. See [LICENSE](LICENSE) for details.

> ⚠️ The license covers only the project's code: the game assets in `assets/` (e.g. the item icons inside `resourceId.csv`) and the game data belong to SEGA / Colorful Palette and other respective owners — **not covered by the MIT license** — please do not use them outside this tool.
