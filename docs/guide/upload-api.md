<!-- GENERATED from README.md; do not edit directly. -->

# Upload API

`POST /uploadMySekai` accepts the encrypted MySekai response body directly. A single upload is normally enough; ordered chunks remain supported for capture-client compatibility.

| Header | Required | Description |
| --- | --- | --- |
| `X-Upload-Id` | Yes | Task identifier matching `^[A-Za-z0-9_-]{1,64}$` |
| `X-Chunk-Index` | Yes | Zero-based chunk index |
| `X-Total-Chunks` | Yes | Total chunk count, from 1 through 10 |
| `X-Original-Url` | No | Original game URL; `/user/<id>` supplies the player route |
| `X-Script-Version` | No | Accepted for capture-client compatibility and ignored by the service |

The encrypted archive, each chunk, and the merged upload are limited to 1 MiB. A successfully accepted request returns plain-text `OK`; rendering and notification continue in the background.

### Single-upload example

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### Chunked-upload example

Use a shared `X-Upload-Id`, ordered indices, and at most ten chunks:

```bash
file=mysekai.bin
id=$(openssl rand -hex 5)
split -b 262144 -a 2 -d "$file" /tmp/ms_chunk_
total=$(ls /tmp/ms_chunk_* | wc -l | tr -d ' ')

i=0
for chunk in /tmp/ms_chunk_*; do
  curl -s -X POST http://127.0.0.1:9478/uploadMySekai \
    -H "X-Upload-Id: $id" \
    -H "X-Chunk-Index: $i" \
    -H "X-Total-Chunks: $total" \
    -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
    --data-binary @"$chunk"
  echo
  i=$((i + 1))
done
rm -f /tmp/ms_chunk_*
```

Common responses are `200 OK` for an accepted upload, `400 Bad Request` for invalid identifiers or chunk ranges, `413 Payload Too Large` for a size limit, and `422 Unprocessable Entity` for missing or non-integer required upload headers.
