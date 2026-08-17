# 快速上手

先完成安裝與 `.env` 基礎設定，再依你想要的推播方式選擇路徑：

- **路徑 A（僅 Telegram Bot 推播）**：設定最少，建議先跑通這條；
- **路徑 B（啟用 Bark 推播）**：在路徑 A 的基礎上，需要額外設定 Bark key、玩家路由與靜態檔案伺服器。

## 1. 安裝

```bash
python -m venv venv
venv/bin/pip install -r requirements.txt
# 可選:安裝 mysekai 命令(等同於 python cli.py ...)
venv/bin/pip install -e .
```

## 2. 設定 .env（必填項目）

```bash
cp .env.example .env
```

`AES_KEY` / `AES_IV` 為 MySekai 存檔的 AES-128-CBC 解密金鑰（各 16 位元組），無論走哪條路徑都必須填寫。其餘變數依選擇的路徑設定：

| 變數 | 必填 | 說明 |
| --- | --- | --- |
| `AES_KEY` / `AES_IV` | ✅ | MySekai 存檔的 AES-128-CBC 金鑰，各 16 位元組 |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | 選* | Telegram 推播（預設管道）需要，來自 [@BotFather](https://t.me/BotFather) |
| `BARK_ICON` | 選 | Bark 通知圖示 URL |
| `BARK_IMAGE_BASE` | 選 | 靜態檔案伺服器根位址（推播 Bark 圖片直連用，見下文） |
| `FALLBACK_IMAGE_BASE` | 選 | 未設定 `BARK_IMAGE_BASE` 時的圖片直連備用位址 |

::: warning
\* 若只想用 Bark 收通知：可留空 Telegram 設定，但**必須在 `config/push_map.json` 裡把玩家路由到 Bark 別名**，否則未設定的玩家預設走 Telegram，而 Telegram 缺設定時只會印出一行警告並跳過，結果是什麼都不推。
:::

## 3. 路徑 A：僅 Telegram Bot 推播（最簡）

適用場景：只要在 Telegram 收到地圖與統計，不折騰其他元件。

1. 在 `.env` 中填寫 Telegram 設定（來自 [@BotFather](https://t.me/BotFather)）：

   ```
   TELEGRAM_BOT_TOKEN=1234567890:AAAA-your-bot-token
   TELEGRAM_CHAT_ID=123456789
   ```

2. 手動跑一遍驗證：

   ```bash
   python cli.py generate <mysekai.bin>
   python cli.py notify data/latest <task_id>
   ```

3. 日常使用：啟動上傳服務，存檔送達後自動產生地圖並推播。兩種抓封包方式任選：

   - **MitM 模組**：按[上傳介面](/zh-TW/guide/upload-api)分片上傳存檔
   - **Reqable 上報伺服器**：設定匹配規則與上報路徑（見[Reqable 上報伺服器](/zh-TW/guide/report-server)）

   ```bash
   python cli.py server [--host 0.0.0.0] [--port 9478]
   ```

路徑 A **不需要**：`config/push_map.json`、`config/bark_map.json`、靜態檔案伺服器、`BARK_IMAGE_BASE`。未設定的玩家預設就推播到 Telegram。

## 4. 路徑 B：啟用 Bark 推播（需額外設定）

在路徑 A 的基礎上（Telegram 設定可保留，也可留空只推 Bark），依順序補齊：

1. **設定 Bark key**：在 `config/bark_map.json` 中為每個別名設定裝置 key（範本見同目錄 `bark_map.example.json`）。
2. **設定玩家路由**：在 `config/push_map.json` 中把玩家 ID 路由到 Bark 別名，例如：

   ```json
   {
     "1234567890123456789": ["klee"],
     "1234567890123456790": ["telegram", "klee"]
   }
   ```

   ::: warning
   **必須設定**：未設定的玩家預設走 Telegram；若此時 Telegram 又未設定，只會印出警告並跳過，結果什麼都不推。
   :::
3. **架設靜態檔案伺服器**：把專案的 `data/` 目錄暴露為公開網路可達的 HTTP(S) 服務，並在 `.env` 設定 `BARK_IMAGE_BASE=https://<域名或IP:端口>`。否則 Bark 通知不帶地圖圖片（詳見[靜態檔案伺服器](/zh-TW/guide/static-server)）。
4. 驗證與日常使用同路徑 A（第 2、3 步）。
