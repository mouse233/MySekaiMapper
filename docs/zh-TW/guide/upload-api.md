<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 上傳 API

`POST /uploadMySekai` 可直接接受已加密的 MySekai 回應主體。通常單次上傳即可；為相容擷取用戶端，仍支援依序傳送的分塊。

| 標頭 | 必要性 | 說明 |
| --- | --- | --- |
| `X-Upload-Id` | 是 | 符合 `^[A-Za-z0-9_-]{1,64}$` 的工作識別碼 |
| `X-Chunk-Index` | 是 | 從零開始的分塊索引 |
| `X-Total-Chunks` | 是 | 分塊總數，範圍從 1 到 10 |
| `X-Original-Url` | 否 | 原始遊戲 URL；`/user/<id>` 可提供玩家路由資訊 |
| `X-Script-Version` | 否 | 為相容擷取用戶端而接受，服務會忽略此值 |

加密封存檔、每個分塊與合併後的上傳皆限制為 1 MiB。成功接受的請求會回傳純文字 `OK`；渲染與通知則在背景繼續執行。

### 單次上傳範例

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H 'X-Upload-Id: demo12345' \
  -H 'X-Chunk-Index: 0' \
  -H 'X-Total-Chunks: 1' \
  -H 'X-Original-Url: https://example.com/user/1234567890123456789' \
  --data-binary @mysekai.bin
```

### 分塊上傳範例

請使用相同的 `X-Upload-Id`、依序的索引，且最多十個分塊：

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

常見回應包括：已接受上傳時的 `200 OK`、識別碼或分塊範圍無效時的 `400 Bad Request`、超過大小限制時的 `413 Payload Too Large`，以及缺少必要上傳標頭或其值非整數時的 `422 Unprocessable Entity`。
