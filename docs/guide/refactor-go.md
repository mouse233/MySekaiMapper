# Go Refactor

The production implementation was migrated from Python to Go to reduce startup cost and resource use, provide bounded background rendering, and ship a single executable.

## What changed

- The repository root is now the Go module: `cmd/`, `internal/`, `go.mod`, and `go.sum`.
- The service includes AES/MsgPack parsing, HAR ingestion, rendering, storage, and Telegram/Bark notification.
- `notify` is available as a standalone Go command.
- Notifications enumerate all regular `site_*.png` outputs.
- Go CI runs tests and builds the service on pushes and pull requests.

## Compatibility notes

The HTTP endpoints, `.env` variables, output names, archive layout, and routing files remain compatible. Go renders fixed-canvas PNGs, so images are not guaranteed to be Matplotlib pixel-identical. The manual notification command uses explicit flags, especially `--player-id` for player-specific routing.

The archived Python implementation is preserved in the `legacy/python` branch and `python-v0.2.0` tag; it is not part of the active runtime.
