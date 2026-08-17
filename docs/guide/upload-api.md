# Upload API

This endpoint receives the captured mysekai response body via `POST /uploadMySekai` (a single POST; chunked upload is kept for compatibility). The same protocol can be debugged manually with curl. Headers:

| Header | Description |
| --- | --- |
| `X-Upload-Id` | Upload task ID (alphanumeric plus `-` / `_`, length 1~64), required |
| `X-Chunk-Index` | Chunk index, starting at 0 (always 0 for a single POST), required |
| `X-Total-Chunks` | Total number of chunks (1~10; use 1 for a single POST), required |
| `X-Original-Url` | The client's original page URL, used to resolve the player ID (e.g. `https://.../user/123456...`); **optional** — if missing, the player ID is recorded as `unknown` |
| `X-Script-Version` | Client script version; ignored by the server, may be omitted |

The request body is the raw binary save data (no multipart needed).

## Limits

- Total file size ≤1MB (`MAX_TOTAL_SIZE`)
- Single chunk ≤1MB (`MAX_CHUNK_SIZE`, returns 413 if exceeded)
- Max 10 chunks (`MAX_CHUNKS`)

::: tip
Current saves are ~200KB, so a **single POST** is all you need. Chunked upload is kept for compatibility with older capture clients; if used, keep each chunk well below 1MB (e.g. 256KB) so 10 chunks fill the 1MB cap.
:::

## Responses

| Status | Meaning |
| --- | --- |
| `200` | Save received, returns `OK`; the server automatically: merges the save (if chunked) → generates maps → archives to `data/archive/by-id/<user_id>/<timestamp>/` → pushes notifications. No manual intervention. |
| `400` | Invalid parameters (bad upload id format, chunk index out of range, total chunks not in 1~10) |
| `413` | Size limit exceeded (single chunk over 1MB, or cumulative total over 1MB) |

## curl examples

Single POST (all current saves fit in one request):

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

Chunked upload (optional, for compatibility; 256KB per chunk fills the 1MB cap with 10 chunks):

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
total=$(( ($(wc -c < "$file") + 262143) / 262144 ))
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_

i=0
for c in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H "X-Original-Url: https://example.com/user/1234567890123456789" \
    --data-binary @"$c"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

A `200 OK` means the save was accepted; the pipeline (merge if chunked → generate → archive → notify) runs automatically. Replace `127.0.0.1:9478` with your actual service address; `X-Upload-Id` must match `^[a-zA-Z0-9_-]{1,64}$` (e.g. a random string from `openssl rand -hex 5`).
