# Reqable Report Server

Reqable can POST captured HAR sessions directly to the Go service.

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

The default endpoint is `POST /reqable/report`. It accepts `identity`, `gzip`, `br`, `zstd`, and `zstandard` request bodies, including streamed zstd frames without a content-size field.

Environment variables:

- `REPORT_ENABLED=0` disables the endpoint.
- `REPORT_PATH=/your/private/path` changes the endpoint.
- `REPORT_MAX_SIZE=1` sets the decompressed-body limit in MiB.
- `REPORT_TOKEN` requires a matching `X-Report-Token` header.

If Reqable cannot add a custom token header, use a private random path and a network-level allowlist.
