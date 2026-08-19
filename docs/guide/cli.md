# CLI Reference

Build the executable once:

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

All commands accept `--root /path/to/MySekaiMapper` when the executable runs outside the repository.

## Inspect

```bash
bin/mysekaimapper inspect --input data/raw_mysekai/save.bin
```

Decrypts a save and prints an aggregate summary without writing maps.

## Generate

```bash
bin/mysekaimapper generate --input data/raw_mysekai/save.bin --output data/latest
```

Creates `site_*.png` and `rare_resources.txt`. If `--output` is omitted, the default is `data/latest`.

## Notify

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789
```

`--player-id` selects the entry in `config/push_map.json`. `--image-base` can override the public URL base used for Bark image links.

## Serve

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The service logs upload, queue, render, archive, notification, player ID, and elapsed-time events. Treat these logs as sensitive operational data.
