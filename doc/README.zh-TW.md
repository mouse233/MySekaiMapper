# MySekaiMapper

🌐 **Languages**: [English](../README.md) · [简体中文](README.zh-CN.md) · [繁體中文](README.zh-TW.md) · [日本語](README.ja-JP.md) · [한국어](README.ko-KR.md)

📖 **Documentation site**: <https://mouse233.github.io/MySekaiMapper/zh-TW/>

這是一項 Go 服務，可將已加密的 *Project SEKAI* MySekai 存檔轉換為資源採集地圖，並將結果傳送至 Telegram 或 Bark（Day.app）。

它可搭配 MitM 擷取用戶端或 Reqable 的 **Report Server** 使用：擷取工具上傳 MySekai 存檔後，服務會解密並解析內容、繪製地圖與稀有資源摘要、封存產物，並自動發送通知，無須手動處理。

一般的 MySekai 區域會產生 `site_5.png`（草原）、`site_6.png`（海灘）、`site_7.png`（花園）、`site_8.png`（紀念場所）及 `rare_resources.txt`。渲染器與通知程式亦能處理其他一般的 `site_*.png` 輸出。

此擷取流程已在 Nuverse 營運的 CN 與 TW 伺服器上驗證。其他地區是否可用，取決於其 API 路徑與存檔格式。

## 運作方式

```text
Game API response → MitM module / Reqable Report Server
    │  ① POST /uploadMySekai (single upload or ordered chunks)
    │  ② POST /reqable/report (HAR, optionally gzip / br / zstd)
    ▼
mysekaimapper serve
    ├─ AES-128-CBC decrypt + MsgPack parse + coordinate normalization
    ├─ render site_*.png + rare_resources.txt
    ├─ archive data/archive/by-id/<player_id>/<timestamp>/
    └─ publish data/latest/ and notify
         ├─ Telegram: upload local images as multipart media groups
         └─ Bark: send image URLs from a public static-file server
```

## 快速開始

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

## 執行服務

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

伺服器會印出就緒 URL，並為上傳／報告接受、佇列處理、解析、渲染、封存、通知、耗時、工作 ID 與 `player_id` 寫入生命週期日誌。它刻意不會記錄封存內容、密鑰、權杖或完整的通知 URL。

程序會處理 `SIGINT` 與 `SIGTERM`：先停止接受 HTTP 請求，接著最多等待 15 秒，以排空已接受的工作。

已編譯的二進位檔可在專案工作區外透過 `--root /path/to/MySekaiMapper` 執行；否則會從工作目錄尋找儲存庫根目錄。

## 上傳 API

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

## Reqable Report Server

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

## 通知與靜態檔案

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

## 命令列參考

先建置一次二進位檔：

```bash
go build -o bin/mysekaimapper ./cmd/mysekaimapper
```

所有指令預設都會載入 `.env`，並接受 `--env /path/to/file`。`--root` 可放在子指令之後的任意位置。

### `inspect`

```bash
bin/mysekaimapper inspect --input mysekai.bin
```

解密並解析存檔，接著輸出安全的彙總 JSON 摘要，不會寫入地圖。

### `generate`

```bash
bin/mysekaimapper generate \
  --input mysekai.bin \
  --output data/latest
```

解密封存檔、擷取掉落物，並寫入 `site_*.png` 與 `rare_resources.txt`。`--output` 預設為 `data/latest`；可用 `--assets` 覆寫素材目錄。

### `notify`

```bash
bin/mysekaimapper notify \
  --output data/latest \
  --task-id manual-001 \
  --player-id 1234567890123456789 \
  --image-base https://maps.example.com/latest
```

必須提供 `--output`。`--task-id` 與 `--player-id` 預設為 `unknown`；需要依玩家進行路由時，請傳入實際的玩家 ID。

### `serve`

```bash
bin/mysekaimapper serve --host 0.0.0.0 --port 9478
```

啟動上傳與報告 HTTP 端點。預設位址為 `0.0.0.0:9478`。

## 目錄結構

```text
.
├── cmd/mysekaimapper/       # CLI 進入點
├── internal/
│   ├── har/                 # Reqable HAR 解析與解壓縮
│   ├── mapper/              # AES、MsgPack、資源與渲染
│   ├── notify/              # Telegram 與 Bark 傳送
│   ├── server/              # 上傳與報告 HTTP 端點
│   └── service/             # 佇列、儲存與封存管線
├── assets/                  # 字型與資源圖示
├── config/                  # 本機路由範本
│   ├── bark_map.example.json
│   └── push_map.example.json
├── data/                    # 忽略的執行階段資料
│   ├── tmp/                 # 上傳暫存區
│   ├── raw_mysekai/         # 加密來源封存檔
│   ├── archive/             # 依玩家與時間戳記保存的歷史產物
│   └── latest/              # 最新產生的產物
├── docs/                    # VitePress 文件
├── go.mod / go.sum          # Go 模組定義
└── .env.example             # 設定範本
```

`data/`、`.env`、`config/bark_map.json` 與 `config/push_map.json` 是私密的執行階段資料，且會被 Git 忽略。

## 測試

```bash
go test ./...
go build -o /tmp/mysekaimapper ./cmd/mysekaimapper
npm run docs:build
```

GitHub Actions 會在推送與提取請求時執行 Go 測試套件與建置。

## Go 重構

目前執行階段僅使用 Go。此模組採用標準根目錄結構，包含 `cmd/`、`internal/`、`go.mod` 與 `go.sum`；Python 原始碼、相依項目與 CI 均已移除。封存的參考實作仍保留在 [`legacy/python`](https://github.com/mouse233/MySekaiMapper/tree/legacy/python) 分支與 [`python-v0.2.0`](https://github.com/mouse233/MySekaiMapper/tree/python-v0.2.0) 標籤中。

HTTP 端點、環境變數、輸出名稱、封存配置與路由檔格式皆維持相容。Go 渲染器使用固定畫布，因此產生的 PNG 不保證與先前 Matplotlib 輸出逐像素相同。

## 免責聲明

本工具僅供個人學習與娛樂使用。請勿將其用於商業用途，或以違反遊戲服務條款的方式使用。遊戲資料與素材歸各自的權利人所有。

## 授權條款

專案程式碼採用 [MIT](https://github.com/mouse233/MySekaiMapper/blob/feat/go-rewrite/LICENSE) 授權（Copyright © 2025 mouse233）。`assets/` 中的遊戲素材與遊戲資料歸各自的權利人所有，且不受本授權條款涵蓋。
