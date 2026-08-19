# MySekaiMapper

A Go service that turns encrypted *Project SEKAI* MySekai saves into resource-gathering maps and sends them to Telegram or Bark.

📖 Documentation: <https://mouse233.github.io/MySekaiMapper/> · [简体中文](docs/zh-CN/index.md)

## Features

- AES-128-CBC decryption, MsgPack parsing, coordinate normalization, and PNG map rendering.
- Chunked upload endpoint and Reqable HAR report endpoint (`gzip`, `brotli`, `zstd`).
- Per-job storage, bounded background rendering, archive/latest publishing, and lifecycle logs.
- Telegram media-group delivery and Bark image-link delivery with per-player routing.
- Standalone `inspect`, `generate`, `notify`, and `serve` commands.

## Quick start

### 1. Configure

```bash
cp .env.example .env
```

Set `AES_KEY` and `AES_IV` to the 16-byte AES-128-CBC values used by the game. Configure Telegram and/or Bark only when notification is needed.

### 2. Build

Go 1.25 or newer is required.

```bash
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

### 3. Run the service

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The service accepts:

- `POST /uploadMySekai` for capture-client uploads;
- `POST /reqable/report` for Reqable HAR reports (enabled by default).

Accepted jobs are queued, rendered, archived under `data/archive/`, and published to `data/latest/`. Shell logs include task ID, `player_id`, rendering status, notification status, and elapsed time.

## Commands

```bash
# Inspect an encrypted save without generating maps.
bin/mysekaimapper inspect --input data/raw_mysekai/save.bin

# Generate maps manually; data/latest is the default output.
bin/mysekaimapper generate --input data/raw_mysekai/save.bin --output data/latest

# Send existing maps manually. player-id selects the push route.
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789
```

Use `--root /path/to/MySekaiMapper` when running the binary outside the repository.

## Configuration

| Variable | Purpose |
| --- | --- |
| `AES_KEY`, `AES_IV` | Required 16-byte AES-128-CBC key and IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | Optional Telegram delivery |
| `BARK_ICON`, `BARK_IMAGE_BASE`, `FALLBACK_IMAGE_BASE` | Optional Bark delivery and image URLs |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | Reqable report endpoint settings |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | Optional path overrides |

Create local routing files from `config/push_map.example.json` and `config/bark_map.example.json`. They contain player/device identifiers and are excluded from Git.

## Go refactor

The active runtime is Go-only. The module now follows the standard root layout with `cmd/`, `internal/`, `go.mod`, and `go.sum`; Python source, dependencies, and CI were removed. The archived Python reference remains available in the [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) branch and [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) tag.

HTTP endpoints, environment variables, output names, archive layout, and routing files remain compatible. Go uses a fixed-canvas renderer, so generated PNGs are not guaranteed to be pixel-identical to the previous Matplotlib output.

## License

Project code is licensed under [MIT](LICENSE). Game assets and game data under `assets/` belong to their respective owners and are not covered by this license.
