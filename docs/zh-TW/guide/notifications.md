<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 通知與靜態檔案

從 `config/push_map.example.json`、`config/bark_map.example.json` 建立本地設定。這些檔案包含玩家／裝置識別資訊，已被 Git 忽略。

### 玩家路由

`config/push_map.json` 將玩家 ID 對映至 `telegram`、`bark:<別名>`、`astrbot:<別名>`、`none`、含管道前綴的 `+tg` 字串或方法陣列：

```json
{
  "1234567890123456789": ["telegram"],
  "1234567890123456790": ["telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

沒有可用路由值的玩家預設使用 Telegram。

Bark 路由必須使用 `bark:<別名>`，與 AstrBot 的 `astrbot:<別名>` 一致。升級時，將 `push_map.json` 中的 `"sls"`、`"xufan"` 改為 `"bark:sls"`、`"bark:xufan"`；`bark_map.json` 的鍵仍為 `"sls"`、`"xufan"`，無需前綴。未加前綴的別名會回報路由錯誤，其他有效目標仍會繼續嘗試。支援 `"bark:klee+tg"` 字串簡寫，但建議使用 `["bark:klee", "telegram"]` 陣列。

### Telegram

Telegram 會將所有產生的常規 `site_*.png` 以本地 multipart 媒體群組的形式上傳。它需要 `TELEGRAM_BOT_TOKEN`、`TELEGRAM_CHAT_ID`，但不需要公開圖片伺服器。Telegram 失敗不會阻止已設定的 Bark 嘗試。

### AstrBot Push Lite（QQ）

在現有 AstrBot 實例安裝 [astrbot_plugin_push_lite](https://github.com/Raven95676/astrbot_plugin_push_lite)。QQ 透過 OneBot v11 接入，並設定插件 API token 與監聽連接埠（預設 `9966`）。從 `config/astrbot_map.example.json` 建立 `config/astrbot_map.json`：

```json
{
  "qq_me": {"platform_id": "qq_main", "type": "private", "qq": "123456789"},
  "qq_group": {"platform_id": "qq_main", "type": "group", "qq": "987654321"},
  "session": {"umo": "qq_main:FriendMessage:123456789"}
}
```

`platform_id` 是 AstrBot 的實際機器人／平台實例 ID，不一定是 `aiocqhttp`，也不是機器人的 QQ 號。`type` 為 `private`（私訊）或 `group`（群組），`qq` 必須以**字串**填寫接收人 QQ 號或群號。程式自動組成 `platform_id:FriendMessage:qq` 或 `platform_id:GroupMessage:qq`。只需用 `/sid` 確認一次平台 ID，無需每位接收人執行指令。亦可只填寫完整 `umo`，不可與其他三個欄位混用。參見 [AstrBot 會話文件](https://docs.astrbot.app/use/command.html#name)。

於 `config/push_map.json` 使用 `astrbot:<別名>`，可與既有管道組合：

```json
{
  "1234567890123456789": ["astrbot:qq_me", "astrbot:qq_group"],
  "1234567890123456790": ["astrbot:qq_me", "telegram", "bark:klee"],
  "1234567890123456791": "none"
}
```

於 `.env` 設定：

```dotenv
ASTRBOT_PUSH_URL=http://astrbot:9966
ASTRBOT_PUSH_TOKEN=your-plugin-api-token
ASTRBOT_ALLOW_INSECURE_HTTP=1
```

`ASTRBOT_PUSH_URL` 是服務**根位址**，不含 `/send`，支援反向代理路徑前綴。預設要求 HTTPS；僅在可信本機／Docker 私有網路啟用 `ASTRBOT_ALLOW_INSECURE_HTTP=1`。不同容器應使用同一網路的 AstrBot 服務名稱，`localhost` 是目前容器；主機部署則使用可達位址及插件對映連接埠。請求以 Bearer token 驗證，不會跟隨重新導向。

`serve` 與 `notify` 對每位接收人先提交文字摘要，再將所有常規 `site_*.png` 分別以 Base64 圖片提交。無需公開圖片伺服器或共享目錄。每張圖片編碼前上限 8 MiB，反向代理亦需允許相應請求大小。同一路由的重複 AstrBot 別名只提交一次；失敗不會阻止其他接收人或管道。

成功僅代表**已進入 Push Lite 佇列**，不代表 QQ 已送達，日誌顯示 `request queued`。本整合不使用回呼或自動重試；插件使用記憶體佇列，重新啟動可能遺失待發訊息，最終結果請查看 AstrBot 日誌。未設定有效路由的玩家仍預設 Telegram。`.env` 與 `config/astrbot_map.json` 已被 Git 忽略，請保持私密。

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
