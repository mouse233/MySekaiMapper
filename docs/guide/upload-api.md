# Upload API

This endpoint receives the captured mysekai response body, uploaded in chunks to `POST /uploadMySekai` (the same protocol can also be debugged manually with curl). Headers:

| Header | Description |
| --- | --- |
| `X-Upload-Id` | Upload task ID (alphanumeric plus `-` / `_`, length 1~64), required |
| `X-Chunk-Index` | Chunk index, starting at 0, required |
| `X-Total-Chunks` | Total number of chunks (1~10), required |
| `X-Original-Url` | The client's original page URL, used to resolve the player ID (e.g. `https://.../user/123456...`); **optional** — if missing, the player ID is recorded as `unknown` |
| `X-Script-Version` | Client script version; ignored by the server, may be omitted |

The request body is the raw binary chunk data (no multipart needed).

## Limits

- Total file size ≤1MB (`MAX_TOTAL_SIZE`)
- Single chunk ≤1MB (`MAX_CHUNK_SIZE`, returns 413 if exceeded)
- Max 10 chunks (`MAX_CHUNKS`)

::: tip
With a total cap of only 1MB, **chunk sizes should be well below 1MB to make sense** (e.g. 256KB, so 10 chunks fill the full 1MB). If a client uses 1MB chunks, any file over 1MB gets rejected with 413 starting from the 2nd chunk — effectively degrading to single-chunk uploads.
:::

## Responses

| Status | Meaning |
| --- | --- |
| `200` | Chunk received, returns `OK`; when the last chunk arrives, the server automatically: merges the save → generates maps → archives to `data/archive/by-id/<user_id>/<timestamp>/` → pushes notifications. No manual intervention. |
| `400` | Invalid parameters (bad upload id format, chunk index out of range, total chunks not in 1~10) |
| `413` | Size limit exceeded (single chunk over 1MB, or cumulative total over 1MB) |

## curl examples

Saves ≤1MB can be uploaded in a single chunk (most common):

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

Chunked upload (256KB per chunk; up to 10 chunks fill the 1MB limit):

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

A `200 OK` per chunk means it was accepted; once the last chunk arrives, the server starts merging and the rest of the pipeline automatically. Replace `127.0.0.1:9478` with your actual service address; `X-Upload-Id` must match `^[a-zA-Z0-9_-]{1,64}$` (e.g. a random string from `openssl rand -hex 5`).
