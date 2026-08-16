# 上傳介面

用戶端把擷取的 mysekai 回應主體分片 POST 到 `POST /uploadMySekai`（手動用 curl 依同一協定除錯亦可）。header 如下：

| Header | 說明 |
| --- | --- |
| `X-Upload-Id` | 上傳任務 ID（僅字母數字與 `-` / `_`，長度 1~64），必填 |
| `X-Chunk-Index` | 分片序號，從 0 開始，必填 |
| `X-Total-Chunks` | 總分片數（1~10），必填 |
| `X-Original-Url` | 用戶端原始頁面 URL，用於解析玩家 ID（如 `https://.../user/123456...`）；**選填**，缺失時玩家 ID 記為 `unknown` |
| `X-Script-Version` | 用戶端腳本版本號；伺服器端忽略該 header，可不傳 |

請求主體為原始二進位分片資料（無需 multipart）。

## 限制

- 單一檔案總大小 ≤1MB（`MAX_TOTAL_SIZE`）
- 單一分片 ≤1MB（`MAX_CHUNK_SIZE`，超出限制回傳 413）
- 總分片數 ≤10（`MAX_CHUNKS`）

::: tip
總大小上限僅 1MB，**分片大小應明顯小於 1MB 才有意義**（例如 256KB，10 片可傳滿 1MB）。若用戶端用 1MB 分片，任何超過 1MB 的檔案都會從第 2 片起被 413 拒絕，實際上退化成只能單片上傳。
:::

## 回應

| 狀態碼 | 含義 |
| --- | --- |
| `200` | 分片已接收，回傳 `OK`；最後一片到達時伺服器端自動完成：合併存檔 → 產生地圖 → 歸檔到 `data/archive/by-id/<user_id>/<時間戳>/` → 推播通知，全程無需人工介入 |
| `400` | 參數不合法（upload id 格式錯誤、分片序號超出範圍、總分片數不在 1~10） |
| `413` | 超過大小限制（單分片超 1MB，或累計總大小超 1MB） |

## curl 範例

存檔 ≤1MB 時單片即可傳完（最常用）：

```bash
curl -X POST http://127.0.0.1:9478/uploadMySekai \
  -H "X-Upload-Id: demo12345" \
  -H "X-Chunk-Index: 0" \
  -H "X-Total-Chunks: 1" \
  -H "X-Original-Url: https://example.com/user/1234567890123456789" \
  --data-binary @mysekai.bin
```

分片上傳（每片 256KB，最多 10 片傳滿 1MB 上限）：

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

每個分片回傳 `200 OK` 即表示已接收；最後一片到達後伺服器端開始合併，其餘流水線自動完成。把 `127.0.0.1:9478` 替換為你的實際服務位址；`X-Upload-Id` 必須符合 `^[a-zA-Z0-9_-]{1,64}$`（例如用 `openssl rand -hex 5` 產生的隨機字串）。
