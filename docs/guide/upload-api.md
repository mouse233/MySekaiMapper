# Upload API

`POST /uploadMySekai` accepts an encrypted MySekai save. It supports one request or ordered chunks.

Required headers:

| Header | Meaning |
| --- | --- |
| `X-Upload-Id` | 1–64 character upload/task identifier |
| `X-Chunk-Index` | Zero-based chunk index |
| `X-Total-Chunks` | Total number of chunks, at most 10 |
| `X-Original-Url` | Optional original game URL; `/user/<id>` is used for player routing |

The default maximum encrypted archive size is 1 MiB. A successful request receives plain-text `OK`; rendering and notification continue asynchronously. Check the service logs for the task lifecycle.
