# Quick Start

## Prerequisites

- Go 1.25+
- AES-128-CBC key and IV from the game client
- Optional Telegram or Bark credentials

## Configure

```bash
cp .env.example .env
```

Set `AES_KEY` and `AES_IV` to 16-byte values. Keep `.env`, `config/bark_map.json`, and `config/push_map.json` private.

## Build and run

```bash
mkdir -p bin
go test ./...
go build -o bin/mysekaimapper ./cmd/mysekaimapper
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The service accepts `POST /uploadMySekai` and, by default, `POST /reqable/report`. Accepted jobs render in a bounded background worker and publish output under `data/archive/` and `data/latest/`.

## Generate a map manually

```bash
bin/mysekaimapper generate --input data/raw_mysekai/save.bin --output data/latest
```
