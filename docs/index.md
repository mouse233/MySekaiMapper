# MySekaiMapper

MySekaiMapper is a Go service that decrypts MySekai saves, renders resource-drop maps, archives the result, and sends it to Telegram or Bark.

## Start here

1. Install Go 1.25 or newer.
2. Copy `.env.example` to `.env` and set `AES_KEY` and `AES_IV`.
3. Build and run the service:

```bash
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

See the [Quick Start](./guide/quickstart), [CLI Reference](./guide/cli), and [Go Refactor](./guide/refactor-go).
