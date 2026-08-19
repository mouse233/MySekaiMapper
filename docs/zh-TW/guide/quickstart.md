<!-- GENERATED from doc/README.zh-TW.md; do not edit directly. -->

# 快速開始

請選擇符合您環境的通知方式：

- **路徑 A — 僅使用 Telegram**：最簡單的選項；不需要玩家路由檔或公開圖片伺服器。
- **路徑 B — 啟用 Bark**：設定 Bark 金鑰、玩家路由與用於圖片的公開靜態檔案伺服器。

### 1. 需求與建置

需要 Go **1.25 或更新版本**。

```bash
go version
cp .env.example .env
go test ./...
mkdir -p bin
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

`.env` 中的 `AES_KEY` 與 `AES_IV` 必須是 16 位元組的 AES-128-CBC 值。請勿提交 `.env` 或本機路由檔。

### 2. 設定 `.env`

| 變數 | 必要性 | 說明 |
| --- | --- | --- |
| `AES_KEY`, `AES_IV` | 是 | 16 位元組的 MySekai AES-128-CBC 金鑰與 IV |
| `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID` | 僅 Telegram | 來自 [@BotFather](https://t.me/BotFather) 的 Bot 憑證與目標聊天 ID |
| `BARK_ICON` | 選用 | 隨 Bark 通知附帶的圖示 URL |
| `BARK_IMAGE_BASE` | Bark 圖片 | 已封存地圖圖片的公開基底 URL |
| `FALLBACK_IMAGE_BASE` | 選用 | 未設定 `BARK_IMAGE_BASE` 時的圖片基底備援值 |
| `REPORT_ENABLED`, `REPORT_PATH`, `REPORT_MAX_SIZE`, `REPORT_TOKEN` | 選用 | Reqable 報告端點設定 |
| `MYSK_ASSETS_DIR`, `MYSK_CONFIG_DIR`, `MYSK_DATA_DIR` | 選用 | 覆寫儲存庫預設目錄 |

### 3. 路徑 A — 僅使用 Telegram

1. 在 `.env` 中設定 Telegram 變數：

    ```dotenv
    TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
    TELEGRAM_CHAT_ID=123456789
    ```

2. 可選：使用既有的加密存檔驗證解析與通知：

    ```bash
    bin/mysekaimapper generate --input data/raw_mysekai/mysekai.bin
    bin/mysekaimapper notify \
      --output data/latest \
      --task-id manual-001 \
      --player-id 1234567890123456789
    ```

3. 啟動服務以進行一般操作：

    ```bash
    bin/mysekaimapper serve --host 0.0.0.0 --port 9478
    ```

未出現在 `config/push_map.json` 中的玩家預設會使用 Telegram。路徑 A 不需要 Bark 對應檔、推送對應檔或公開圖片伺服器。

### 4. 路徑 B — 啟用 Bark

除路徑 A 的設定外（僅使用 Bark 的路由可省略 Telegram）：

1. 由 `config/bark_map.example.json` 建立 `config/bark_map.json`，將每個 Bark 別名對應至裝置金鑰。
2. 由 `config/push_map.example.json` 建立 `config/push_map.json`，將玩家 ID 對應至 Bark 別名、`telegram`、`none`，或它們的組合：

    ```json
    {
      "1234567890123456789": ["klee"],
      "1234567890123456790": ["telegram", "klee"],
      "1234567890123456791": "none"
    }
    ```

3. 透過公開 HTTP(S) 靜態檔案伺服器公開儲存庫的 `data/` 目錄，並將其公開根 URL 設為 `BARK_IMAGE_BASE`：

    ```dotenv
    BARK_IMAGE_BASE=https://maps.example.com
    ```

未設定的玩家預設會使用 Telegram。因此若未設定 Telegram，未設定的玩家將不會收到通知；僅使用 Bark 時，請明確指定 Bark 別名。
