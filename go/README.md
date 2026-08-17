# MySekaiMapper Go prototype

This directory contains an experimental Go implementation of the offline core:

- AES-128-CBC + PKCS#7 archive decryption;
- MsgPack harvesting-drop extraction and coordinate rotation;
- WebP icon loading from the existing `assets/resourceId.csv`;
- standalone PNG map generation and byte-compatible `rare_resources.txt` output.

The Go renderer preserves drop extraction, site rotation, resource icons, and
output filenames, but it deliberately uses a new fixed-canvas visual style. It
is **not** pixel-compatible with Matplotlib yet. It intentionally does **not**
include a captured save, `.env`, generated maps, the HTTP/HAR server, or network
notification code. The Python implementation remains the production reference
while this prototype is validated.

Input archives are capped at 1 MiB, matching the current Python upload API.

## Local validation

From the repository root:

```bash
go -C go test ./...
go -C go run ./cmd/mysekaimapper inspect \
  --input ../data/raw_mysekai/<local-save>.bin

go -C go run ./cmd/mysekaimapper generate \
  --input ../data/raw_mysekai/<local-save>.bin \
  --output ../data/go-rewrite-output/latest
```

A compiled binary can run outside the checkout by providing
`--root /path/to/MySekaiMapper`; otherwise the repository root is discovered
from the current working directory. Relative input paths are resolved from the
current working directory.

The commands load the repository `.env` only locally to obtain `AES_KEY` and
`AES_IV`; neither command prints them. Do not commit local saves or generated
outputs.
