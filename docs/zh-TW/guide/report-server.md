<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# Reqable Report Server

Reqable v2.20.0+ 可將每個擷取到的 HTTP 工作階段以 HAR JSON POST 至此服務。報告端點預設啟用，並可與 `/uploadMySekai` 共存。

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

| 變數 | 預設值 | 說明 |
| --- | --- | --- |
| `REPORT_ENABLED` | `1` | 設為 `0`、`false`、`no` 或 `off` 可停用報告 |
| `REPORT_PATH` | `/reqable/report` | 在 Reqable 中設定的端點路徑 |
| `REPORT_MAX_SIZE` | `1` | HAR 內容解壓縮後的大小上限，單位為 MiB |
| `REPORT_TOKEN` | 空白 | 可選值；若設定，則要求 `X-Report-Token` 中包含該值 |

### 處理流程

對於每份報告，服務會：

1. 解壓縮 `identity`、`gzip`、`br`、`zstd` 或 `zstandard` 內容並剖析 HAR。支援不含內容大小欄位的串流 zstd frame。
2. 遍歷 `log.entries`，接受第一個可使用 `AES_KEY`/`AES_IV` 解密且通過 MySekai 封存檔驗證的回應主體（若無，則退回使用請求主體）。
3. 從相符工作階段 URL 中的 `/user/<id>` 擷取 `player_id`。
4. 將加密封存檔儲存至 `data/raw_mysekai/`，並啟動與上傳所使用的相同 render → archive → notify 管線。

> Reqable 只會為每個工作階段報告一次，且不會重試。請保持服務可用，並留意 `[REPORT]` 記錄。即使語法有效的 HAR 不含 MySekai 封存檔，仍會收到 `ok`；每份報告只會處理第一個有效封存檔。

### 設定 Reqable

- **比對規則**：`https://<game-api-domain>/api/user/*/mysekai*`
- **伺服器 URL**：`http://<your-server>:9478/reqable/report`（或您自訂的 `REPORT_PATH`）

| 伺服器 | 遊戲 API 網域 |
| --- | --- |
| JP | `https://production-game-api.sekai.colorfulpalette.org` |
| EN | `https://n-production-game-api.sekai-en.com` |
| TW | `https://mk-zian-obt-cdn.bytedgame.com` |
| KR | `https://mkkorea-obt-prod01-cdn.bytedgame.com` |
| CN | `https://mkcn-prod-public-60001-1.dailygn.com` |

此比對模式已針對 CN 驗證。若您的地區使用其他 MySekai API 路徑，請檢查其擷取到的 URL 並調整規則。

### 安全性

Reqable 無法新增自訂的 `X-Report-Token` 標頭。請使用較長的隨機 `REPORT_PATH`，例如 `/reqable/report/<random>`，並透過反向代理或防火牆限制存取；未採取控管措施時，請勿將預設端點公開至網際網路。

### 手動 gzip HAR 測試

```bash
gzip -c report.har.json | curl -X POST http://127.0.0.1:9478/reqable/report \
  -H 'Content-Type: application/json' \
  -H 'Content-Encoding: gzip' \
  --data-binary @-
```
