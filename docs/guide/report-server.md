# Reqable Report Server (optional)

Reqable's built-in **Report Server** feature (v2.20.0+) automatically POSTs each captured HTTP session to your own server in HAR JSON format, optionally compressed with gzip / brotli / zstd. Enable the matching endpoint with `REPORT_ENABLED=1`:

```bash
REPORT_ENABLED=1 python cli.py server
```

Configuration (`.env`):

| Variable | Default | Description |
| --- | --- | --- |
| `REPORT_ENABLED` | *(empty = off)* | Set to `1` / `true` to enable the report endpoint |
| `REPORT_PATH` | `/reqable/report` | Endpoint path; fill this into Reqable's "Upload Path" field |
| `REPORT_MAX_SIZE` | `8` | Max HAR body size in MB (the save itself must stay ≤1MB; base64 inflates it ~33%) |
| `REPORT_TOKEN` | *(empty)* | Optional shared token; when set, the endpoint requires the `X-Report-Token` header |

## How it works

For each report the server:

1. Decompresses the body (`Content-Encoding: gzip` / `br` / `zstd`) and parses the HAR.
2. Walks `log.entries` and takes the first session whose response body (fallback: request body) decrypts with `AES_KEY` / `AES_IV` and parses as a MySekai save — unrelated API traffic is skipped.
3. Resolves the player ID from the session URL (`/user/<id>`).
4. Saves the archive to `data/raw_mysekai/` and launches the same generate → archive → notify pipeline as chunked uploads.

::: warning
Reqable sends each session **exactly once and never retries**, so the endpoint answers `200` as fast as possible. Keep your server stable and watch the `[REPORT]` log lines.
:::

Only **one** archive per report is processed (the first valid entry), so a rule matching many endpoints won't cause duplicate pushes.

## Security

The protocol has no built-in auth. Since Reqable cannot attach custom headers, prefer embedding a random secret in `REPORT_PATH` (e.g. `/reqable/report/9f3a…`) or restricting access with a reverse proxy / firewall instead of relying on `REPORT_TOKEN`.

## Reqable configuration

- URL matching rule: `https://<game-api-host>/*` (or narrow it, e.g. `https://<game-api-host>/user/*/mysekai*`)
- Upload path: `http://<your-server>:9478/reqable/report`
- Compression: gzip / brotli / zstd — all three are supported by the server

## curl example

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```
