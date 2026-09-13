<!-- GENERATED from README.md; do not edit directly. -->

# Command-line reference

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
