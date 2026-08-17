# 上傳介面

用戶端把擷取的 mysekai 回應主體透過 `POST /uploadMySekai` 上傳（一次 POST 即可；分片上傳僅作相容保留）。手動用 curl 依同一協定除錯亦可。header 如下：

| Header | 說明 |
| --- | --- |
| `X-Upload-Id` | 上傳任務 ID（僅字母數字與 `-` / `_`，長度 1~64），必填 |
| `X-Chunk-Index` | 分片序號，從 0 開始（單片上傳恆為 0），必填 |
| `X-Total-Chunks` | 總分片數（1~10；單片上傳填 1），必填 |
| `X-Original-Url` | 用戶端原始頁面 URL，用於解析玩家 ID（如 `https://.../user/123456...`）；**選填**，缺失時玩家 ID 記為 `unknown` |
| `X-Script-Version` | 用戶端腳本版本號；伺服器端忽略該 header，可不傳 |

請求主體為原始二進位存檔資料（無需 multipart）。

## 限制

- 單一檔案總大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 單一分片 ≤1MB（`MAX_CHUNK_SIZE`，超出限制回傳 413）
- 總分片數 ≤10（`MAX_CHUNKS`）

::: tip
目前存檔約 200KB，**一次 POST 即可傳完**。分片上傳僅為相容舊抓封包用戶端保留；若使用分片，每片應明顯小於 1MB（例如 256KB），10 片可傳滿 1MB 上限。
:::

## 回應

| 狀態碼 | 含義 |
| --- | --- |
| `200` | 存檔已接收，回傳 `OK`；伺服器端自動完成：合併存檔（如分片）→ 產生地圖 → 歸檔到 `data/archive/by-id/<user_id>/<時間戳>/` → 推播通知，全程無需人工介入 |
| `400` | 參數不合法（upload id 格式錯誤、分片序號超出範圍、總分片數不在 1~10） |
| `413` | 超過大小限制（單分片超 1MB，或累計總大小超 1MB） |

## curl 範例

單次 POST（目前存檔一次即可傳完）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分片上傳（選用，相容舊用戶端；每片 256KB，10 片傳滿 1MB 上限）：

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

回傳 `200 OK` 即表示存檔已接收；流水線（如分片則先合併 → 產生地圖 → 歸檔 → 推播）自動完成。把 `127.0.0.1:9478` 替換為你的實際服務位址；`X-Upload-Id` 必須符合 `^[a-zA-Z0-9_-]{1,64}$`（例如用 `openssl rand -hex 5` 產生的隨機字串）。
