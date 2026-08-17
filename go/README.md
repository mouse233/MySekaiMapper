# MySekaiMapper Go migration

This directory contains the staged Go migration of MySekaiMapper. It now covers:

- AES-128-CBC + PKCS#7 archive decryption and tolerant MsgPack drop extraction;
- WebP icon loading from the existing `assets/resourceId.csv` and PNG map output;
- Reqable HAR ingestion (`identity`, `gzip`, `x-gzip`, `br`, `zstd`, `zstandard`);
- the capture-client chunk endpoint and Reqable report endpoint;
- raw-save storage, per-job rendering, archive/latest promotion, and bounded render workers;
- local Bark/Telegram routing, including text, photo, and media-group requests.

The Python service remains the production reference during migration. The Go
renderer preserves output names and `rare_resources.txt` content semantics but
uses a fixed-canvas visual style, so PNGs are **not** Matplotlib pixel-compatible.

## Commands

From the repository root:

```bash
go -C go test ./...

go -C go run ./cmd/mysekaimapper inspect \
  --input ../data/raw_mysekai/<local-save>.bin

go -C go run ./cmd/mysekaimapper generate \
  --input ../data/raw_mysekai/<local-save>.bin \
  --output ../data/go-rewrite-output/latest

go -C go run ./cmd/mysekaimapper serve --host 127.0.0.1 --port 9478
```

`serve` provides:

- `POST /uploadMySekai` → plain-text `OK`;
- `POST /reqable/report` (or `REPORT_PATH`) → plain-text `ok`.

A compiled binary can run outside the checkout with
`--root /path/to/MySekaiMapper`; otherwise the repository root is discovered
from the current working directory.

## Local configuration

Commands load the repository `.env` locally for `AES_KEY` and `AES_IV`; values
are never printed. The service also recognizes the existing `REPORT_ENABLED`,
`REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN`, `TELEGRAM_BOT_TOKEN`,
`TELEGRAM_CHAT_ID`, `BARK_ICON`, `BARK_IMAGE_BASE`, and
`FALLBACK_IMAGE_BASE` settings. Optional path overrides are
`MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, and `MYSK_DATA_DIR`.

Bark and push routing maps are read from `config/bark_map.json` and
`config/push_map.json` (or the overridden config directory). Network calls are
made only after a successful local archive/render step.

## Safety and validation

- Uploads, HAR bodies, decoded archive candidates, and raw saves are capped at
  1 MiB by default; uploads permit at most ten chunks.
- HAR decompression and local JSON/config reads have explicit size bounds.
- Raw saves and chunk files use owner-only permissions; generated outputs are
  published through archive/latest staging swaps.
- Notification failures are isolated from rendering and do not print tokens,
  Bark keys, or full request URLs. Production notification endpoints must use
  HTTPS by default.
- `serve` defaults to `0.0.0.0` for LAN capture clients; bind to `127.0.0.1`
  when possible, and use a firewall/reverse proxy when exposing either upload
  endpoint beyond a trusted network.
- Tests use synthetic AES/MsgPack/HAR data and local `httptest` endpoints only.
  Do not commit raw saves, `.env`, generated maps, or cache directories.
