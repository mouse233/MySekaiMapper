# Reqable 上報伺服器

Reqable 內建的「上報伺服器」功能（v2.20.0+）會把每個已捕獲的 HTTP 工作階段依 HAR JSON 格式自動 POST 到你自建的主機端，可選 gzip / brotli / zstd 壓縮。上報端點**預設開啟**，與分片上傳共存——`python cli.py server` 同時提供兩個介面；設 `REPORT_ENABLED=0` 可關閉：

```bash
python cli.py server
```

設定（`.env`）：

| 變數 | 預設值 | 說明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1`（開啟） | 設 `0` / `false` 關閉上報端點 |
| `REPORT_PATH` | `/reqable/report` | 端點路徑，填入 Reqable 的「上報路徑」 |
| `REPORT_MAX_SIZE` | `8` | HAR 請求主體大小上限（MB）；存檔本身需 ≤1MB，base64 膨脹約 33% |
| `REPORT_TOKEN` | （空） | 選用共享令牌；設定後端點要求請求頭 `X-Report-Token` 匹配 |

## 處理流程

每次上報，主機端會：

1. 依 `Content-Encoding`（gzip / br / zstd）解壓並解析 HAR。
2. 遍歷 `log.entries`，取第一個「回應主體（兜底：請求主體）能用 `AES_KEY` / `AES_IV` 解密並解析為 MySekai 存檔」的工作階段——命中規則但與存檔無關的流量會被跳過。
3. 從工作階段 URL 解析玩家 ID（`/user/<id>`）。
4. 存檔儲存到 `data/raw_mysekai/`，並啟動與分片上傳相同的 生成 → 歸檔 → 推播 流水線。

::: warning
Reqable 每個工作階段**只上報 1 次且失敗不重試**，因此端點會盡快回傳 `200`。請保持服務穩定，並留意 `[REPORT]` 日誌。
:::

每次上報只處理 **1 份**存檔（第一個有效條目），因此匹配多個介面的規則不會造成重複推播。

## 安全性

協定本身沒有鑑權。Reqable 無法附加自訂請求頭，建議把隨機字串拼進 `REPORT_PATH`（如 `/reqable/report/9f3a…`），或用反向代理 / 防火牆限制存取，而不是依賴 `REPORT_TOKEN`。

## Reqable 側設定

- URL 匹配規則：`https://<遊戲API網域>/api/user/*/mysekai*`
- 上報路徑：`http://<你的伺服器>:9478/reqable/report`
- 壓縮演算法：gzip / brotli / zstd 皆可（主機端三種都支援）

五個伺服器的遊戲 API 網域：

| 伺服器 | 遊戲 API 網域 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

推薦匹配規則：`https://<網域>/api/user/*/mysekai*`（CN 已實測驗證）。若你所在伺服器的 mysekai 介面路徑不同，請依實際路徑調整規則。

## curl 範例

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H "Content-Type: application/json" -H "Content-Encoding: gzip" \
  --data-binary @-
```
