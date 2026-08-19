<!-- GENERATED from README.md; do not edit directly. -->

# Reqable Report Server

Reqable v2.20.0+ can POST each captured HTTP session to this service as HAR JSON. The report endpoint is enabled by default and coexists with `/uploadMySekai`.

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| Variable | Default | Description |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | Set `0`, `false`, `no`, or `off` to disable reports |
| `REPORT_PATH` | `/reqable/report` | Endpoint path configured in Reqable |
| `REPORT_MAX_SIZE` | `1` | Maximum decompressed HAR body size in MiB |
| `REPORT_TOKEN` | empty | Optional value required in `X-Report-Token` |

### Processing flow

For each report, the service:

1. Decompresses `identity`, `gzip`, `br`, `zstd`, or `zstandard` content and parses the HAR. Streamed zstd frames without a content-size field are supported.
2. Walks `log.entries` and accepts the first response body (falling back to its request body) that decrypts with `AES_KEY`/`AES_IV` and validates as a MySekai archive.
3. Extracts `player_id` from `/user/<id>` in the matched session URL.
4. Saves the encrypted archive in `data/raw_mysekai/` and starts the same render → archive → notify pipeline used by uploads.

> Reqable reports each session once and does not retry. Keep the service available and watch `[REPORT]` logs. A syntactically valid HAR with no MySekai archive still receives `ok`; only the first valid archive in a report is processed.

### Configure Reqable

- **Matching rule**: `https://<game-api-domain>/api/user/*/mysekai*`
- **Server URL**: `http://<your-server>:9478/reqable/report` (or your custom `REPORT_PATH`)

| Server | Game API domain |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

The matching pattern has been verified for CN. If your region uses another MySekai API path, inspect its captured URL and adjust the rule.

### Security

Reqable cannot add the custom `X-Report-Token` header. Use a long random `REPORT_PATH` such as `/reqable/report/<random>` and restrict access through a reverse proxy or firewall; do not expose the default endpoint publicly without controls.

### Manual gzip HAR test

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```
