<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 通知與靜態檔案

從 `config/push_map.example.json`、`config/bark_map.example.json` 建立本地設定。這些檔案包含玩家／裝置識別資訊，已被 Git 忽略。

### 玩家路由

`config/push_map.json` 將玩家 ID 對映至 `telegram`、Bark 別名、`none`、`+tg` 字串或方法陣列：

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "klee"],
  "1234567890123456791": "none"
}
```

沒有可用路由值的玩家預設使用 Telegram。

### Telegram

Telegram 會將所有產生的常規 `site_*.png` 以本地 multipart 媒體群組的形式上傳。它需要 `TELEGRAM_BOT_TOKEN`、`TELEGRAM_CHAT_ID`，但不需要公開圖片伺服器。Telegram 失敗不會阻止已設定的 Bark 嘗試。

### Bark

Bark 會傳送稀有資源摘要，並針對所有產生的常規 `site_*.png` 分別發送通知。`config/bark_map.json` 負責別名與裝置金鑰的對映：

```json
{ "klee": "paste-your-bark-key-here" }
```

Bark 會自行擷取圖片 URL。自動服務工作應將 `BARK_IMAGE_BASE` 指向公開的 `data/` 根目錄，歸檔 URL 如下：

```text
https://maps.example.com/archive/by-id/<player_id>/<timestamp>/site_5.png
```

手動 `notify` 的圖片根路徑優先順序為 `--image-base`、`BARK_IMAGE_BASE`、`FALLBACK_IMAGE_BASE`；該根路徑應直接公開所選的輸出目錄。

### 靜態檔案伺服器

Bark 圖片不可使用 `localhost` 或 `127.0.0.1`。請使用公開的 HTTPS，例如：

```nginx
server {
    listen 443 ssl;
    server_name maps.example.com;
    root /path/to/MySekaiMapper/data;
}
```

```bash
caddy file-server --root /path/to/MySekaiMapper/data --listen :443
```

通知器會忽略輸出目錄中的符號連結，不會記錄憑據或完整的通知 URL。
